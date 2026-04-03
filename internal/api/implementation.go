//go:generate oapi-codegen -package api -generate types,gin,spec internal/market/api.yaml > internal/api/generated.go

package api

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"kraken-trader/internal/decision"
	"kraken-trader/internal/market"
	"kraken-trader/pkg/kraken"
	"kraken-trader/pkg/logger"
)

func NewGinRouter(col *market.Collector, eng *decision.Engine, krakenClient *kraken.Client) *gin.Engine {
	router := gin.Default()
	server := &Server{collector: col, engine: eng, krakenClient: krakenClient}
	RegisterHandlersWithOptions(router, server, GinServerOptions{})
	return router
}

type Server struct {
	collector    *market.Collector
	engine       *decision.Engine
	krakenClient *kraken.Client
}

func NewServer(col *market.Collector, eng *decision.Engine, krakenClient *kraken.Client) *Server {
	return &Server{collector: col, engine: eng, krakenClient: krakenClient}
}

func (s *Server) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) ListSubscriptions(c *gin.Context) {
	subs := s.collector.ListSubscriptions()
	result := make([]SubscriptionDetail, len(subs))
	for i, sym := range subs {
		createdAt, lastData, _ := s.collector.GetSubscriptionDetail(sym)
		isActive := s.collector.IsSubscribed(sym)
		result[i] = SubscriptionDetail{
			Symbol:    &sym,
			IsActive:  &isActive,
			CreatedAt: &createdAt,
			LastData:  &lastData,
		}
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": result, "count": len(result)})
}

func (s *Server) AddSubscription(c *gin.Context) {
	var req AddSubscriptionJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}
	wsSymbol := s.collector.FormatSymbol(req.Symbol)
	ok := s.collector.AddSubscription(wsSymbol)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "symbol already subscribed"})
		return
	}
	s.collector.RestartWebSocket()
	c.JSON(http.StatusOK, gin.H{"status": "subscribed", "symbol": wsSymbol})
}

func (s *Server) RemoveSubscription(c *gin.Context, symbol string) {
	if symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}
	wsSymbol := s.collector.FormatSymbol(symbol)
	ok := s.collector.RemoveSubscription(wsSymbol)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "symbol not found"})
		return
	}
	s.collector.RestartWebSocket()
	c.JSON(http.StatusOK, gin.H{"status": "unsubscribed", "symbol": wsSymbol})
}

func (s *Server) ListSubscriptionsDetail(c *gin.Context) {
	subs := s.collector.ListSubscriptions()
	result := make([]SubscriptionDetail, len(subs))
	for i, sym := range subs {
		createdAt, lastData, _ := s.collector.GetSubscriptionDetail(sym)
		isActive := s.collector.IsSubscribed(sym)
		result[i] = SubscriptionDetail{
			Symbol:    &sym,
			IsActive:  &isActive,
			CreatedAt: &createdAt,
			LastData:  &lastData,
		}
	}
	c.JSON(http.StatusOK, gin.H{"subscriptions": result, "count": len(result)})
}

func (s *Server) ListPrompts(c *gin.Context, params ListPromptsParams) {
	if s.engine == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "decision engine not available"})
		return
	}
	limit := 20
	if params.Limit != nil {
		limit = *params.Limit
	}
	prompts, err := s.engine.GetPrompts(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get prompts"})
		return
	}
	if prompts == nil {
		c.JSON(http.StatusOK, gin.H{"prompts": []PromptRecord{}, "count": 0})
		return
	}
	c.JSON(http.StatusOK, gin.H{"prompts": prompts, "count": len(prompts)})
}

func (s *Server) GetLogLevel(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"level": logger.GetConsoleLevel()})
}

func (s *Server) SetLogLevel(c *gin.Context) {
	var req SetLogLevelJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if err := logger.SetConsoleLevel(string(req.Level)); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid log level"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"status": "ok", "level": req.Level})
}

func (s *Server) GetAssets(c *gin.Context, params GetAssetsParams) {
	assets, err := s.krakenClient.GetAssets(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assets"})
		return
	}

	result := make([]AssetInfo, 0, len(assets))
	enabledOnly := true
	if params.EnabledOnly != nil {
		enabledOnly = *params.EnabledOnly
	}

	for name, data := range assets {
		if enabledOnly {
			status, ok := data["status"].(string)
			if !ok || status != "enabled" {
				continue
			}
		}

		asset := AssetInfo{
			Altname: &name,
		}

		if v, ok := data["aclass"].(string); ok {
			asset.Aclass = &v
		}
		if v, ok := data["altname"].(string); ok {
			asset.Altname = &v
		}
		if v, ok := data["decimals"].(float64); ok {
			i := int(v)
			asset.Decimals = &i
		}
		if v, ok := data["display_decimals"].(float64); ok {
			i := int(v)
			asset.DisplayDecimals = &i
		}
		if v, ok := data["status"].(string); ok {
			asset.Status = &v
		}
		if v, ok := data["margin_rate"].(string); ok {
			asset.MarginRate = &v
		}
		if v, ok := data["collateral_value"].(float64); ok {
			f := float32(v)
			asset.CollateralValue = &f
		}

		result = append(result, asset)
	}

	c.JSON(http.StatusOK, gin.H{"assets": result, "count": len(result)})
}

func init() {
	_ = time.Time{}
}
