//go:generate oapi-codegen -package api -generate types,gin,spec internal/market/api.yaml > internal/api/generated.go

package api

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"kraken-trader/internal/decision"
	"kraken-trader/internal/market"
	"kraken-trader/internal/state"
	"kraken-trader/pkg/kraken"
	"kraken-trader/pkg/logger"
)

func NewGinRouter(col *market.Collector, eng *decision.Engine, krakenClient *kraken.Client, stateMgr *state.MemoryManager) *gin.Engine {
	router := gin.Default()
	server := &Server{collector: col, engine: eng, krakenClient: krakenClient, stateMgr: stateMgr}
	RegisterHandlersWithOptions(router, server, GinServerOptions{})
	return router
}

type Server struct {
	collector    *market.Collector
	engine       *decision.Engine
	krakenClient *kraken.Client
	stateMgr     *state.MemoryManager
}

func NewServer(col *market.Collector, eng *decision.Engine, krakenClient *kraken.Client, stateMgr *state.MemoryManager) *Server {
	return &Server{collector: col, engine: eng, krakenClient: krakenClient, stateMgr: stateMgr}
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
	ok := s.collector.AddSubscription(req.Symbol)
	if !ok {
		c.JSON(http.StatusConflict, gin.H{"error": "symbol already subscribed"})
		return
	}
	s.collector.RestartWebSocket()
	c.JSON(http.StatusOK, gin.H{"status": "subscribed", "symbol": req.Symbol})
}

func (s *Server) DeleteSubscription(c *gin.Context) {
	var req DeleteSubscriptionJSONBody
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	if req.Symbol == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "symbol is required"})
		return
	}
	ok := s.collector.RemoveSubscription(req.Symbol)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "symbol not found"})
		return
	}
	s.collector.RestartWebSocket()
	c.JSON(http.StatusOK, gin.H{"status": "unsubscribed", "symbol": req.Symbol})
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
	log.Info().Msg("GetAssets called - fetching from kraken")
	assets, err := s.krakenClient.GetAssets(c.Request.Context())
	if err != nil {
		log.Error().Err(err).Msg("GetAssets failed")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get assets"})
		return
	}

	log.Info().Int("count", len(assets)).Msg("Got assets from kraken")
	if len(assets) == 0 {
		log.Warn().Msg("Assets map is empty!")
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

func (s *Server) GetPairs(c *gin.Context) {
	pairs := s.collector.GetAvailablePairs()
	result := make([]TradingPair, 0, len(pairs))
	for _, p := range pairs {
		sym := p.Symbol
		alt := p.Altname
		ws := p.WsName
		result = append(result, TradingPair{
			Symbol:  &sym,
			Altname: &alt,
			WsName:  &ws,
		})
	}
	c.JSON(http.StatusOK, gin.H{"pairs": result, "count": len(result)})
}

func (s *Server) GetTicker(c *gin.Context, symbol string) {
	symbol, _ = url.QueryUnescape(symbol)
	pair := s.collector.FormatSymbol(symbol)
	state, ok := s.stateMgr.GetMarketSnapshot(pair)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "symbol not found or no data"})
		return
	}
	bid := float32(state.Bid)
	ask := float32(state.Ask)
	last := float32(state.Last)
	vol := float32(state.Volume24h)
	c.JSON(http.StatusOK, TickerData{
		Symbol:    &state.Pair,
		Bid:       &bid,
		Ask:       &ask,
		Last:      &last,
		Volume:    &vol,
		UpdatedAt: &state.UpdatedAt,
	})
}
