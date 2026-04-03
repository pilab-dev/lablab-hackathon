//go:generate oapi-codegen -package api -generate types,gin,spec internal/market/api.yaml > internal/api/generated.go

package api

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"

	"kraken-trader/internal/decision"
	"kraken-trader/internal/market"
	"kraken-trader/internal/state"
	"kraken-trader/internal/storage"
	"kraken-trader/pkg/kraken"
	"kraken-trader/pkg/logger"
)

func NewGinRouter(col *market.Collector, eng *decision.Engine, krakenClient *kraken.Client, stateMgr *state.MemoryManager, storageClient *storage.Client) *gin.Engine {
	router := gin.Default()
	server := &Server{collector: col, engine: eng, krakenClient: krakenClient, stateMgr: stateMgr, storageClient: storageClient}
	RegisterHandlersWithOptions(router, server, GinServerOptions{})
	return router
}

type Server struct {
	collector     *market.Collector
	engine        *decision.Engine
	krakenClient  *kraken.Client
	stateMgr      *state.MemoryManager
	storageClient *storage.Client
}

func NewServer(col *market.Collector, eng *decision.Engine, krakenClient *kraken.Client, stateMgr *state.MemoryManager, storageClient *storage.Client) *Server {
	return &Server{collector: col, engine: eng, krakenClient: krakenClient, stateMgr: stateMgr, storageClient: storageClient}
}

func (s *Server) GetHealth(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

func (s *Server) ListSubscriptions(c *gin.Context) {
	subs := s.collector.ListSubscriptions()
	result := make([]SubscriptionDetail, len(subs))
	errors := s.collector.GetSubscriptionErrors()
	for i, sym := range subs {
		createdAt, lastData, _ := s.collector.GetSubscriptionDetail(sym)
		isActive := s.collector.IsSubscribed(sym)
		detail := SubscriptionDetail{
			Symbol:    &sym,
			IsActive:  &isActive,
			CreatedAt: &createdAt,
			LastData:  &lastData,
		}
		if errMsg, hasError := errors[sym]; hasError {
			detail.LastError = &errMsg
		}
		result[i] = detail
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

func (s *Server) GetHistory(c *gin.Context, symbol string, params GetHistoryParams) {
	symbol, _ = url.QueryUnescape(symbol)
	pair := s.collector.FormatSymbol(symbol)

	timeframe := "1d"
	if params.Timeframe != nil {
		timeframe = string(*params.Timeframe)
	}

	limit := 100
	if params.Limit != nil {
		limit = *params.Limit
		if limit < 1 || limit > 1000 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 1000"})
			return
		}
	}

	historicalData, err := s.storageClient.GetHistoricalData(c.Request.Context(), pair, timeframe, limit)
	if err != nil {
		log.Error().Err(err).Str("symbol", symbol).Str("pair", pair).Msg("Failed to get historical data")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get historical data"})
		return
	}

	data := make([]OHLCData, len(historicalData))
	for i, h := range historicalData {
		open := float32(h.Open)
		high := float32(h.High)
		low := float32(h.Low)
		close := float32(h.Close)
		volume := float32(h.Volume)
		data[i] = OHLCData{
			Timestamp: &h.Timestamp,
			Open:      &open,
			High:      &high,
			Low:       &low,
			Close:     &close,
			Volume:    &volume,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"symbol":    pair,
		"timeframe": timeframe,
		"data":      data,
		"count":     len(data),
	})
}

func (s *Server) GetDashboard(c *gin.Context) {
	portfolio := s.stateMgr.GetPortfolioSummary()
	marketSnapshots := s.stateMgr.GetAllMarketSnapshots()
	active, stale, errored := s.collector.GetSubscriptionHealth()
	errors := s.collector.GetSubscriptionErrors()
	recentAlerts := s.stateMgr.GetRecentAlerts(10)

	var recentDecisions []PromptRecord
	if s.engine != nil {
		decisions, err := s.engine.GetPrompts(c.Request.Context(), 5)
		if err == nil && decisions != nil {
			recentDecisions = make([]PromptRecord, len(decisions))
			for i, d := range decisions {
				id := int(d.ID)
				sizePct := float32(d.SizePct)
				confidence := float32(d.Confidence)
				recentDecisions[i] = PromptRecord{
					Id:         &id,
					CreatedAt:  &d.CreatedAt,
					Type:       &d.Type,
					Pair:       &d.Pair,
					RawPrompt:  &d.RawPrompt,
					RawAnswer:  &d.RawAnswer,
					Answer:     &d.Answer,
					Action:     &d.Action,
					SizePct:    &sizePct,
					Confidence: &confidence,
					Success:    &d.Success,
				}
			}
		}
	}

	totalValueUSD := float32(portfolio.TotalValueUSD)
	openPositions := portfolio.OpenPositions
	updatedAt := portfolio.UpdatedAt

	balances := make(map[string]float32, len(portfolio.Balances))
	for k, v := range portfolio.Balances {
		balances[k] = float32(v)
	}

	marketData := make([]MarketSnapshot, 0, len(marketSnapshots))
	for _, snap := range marketSnapshots {
		sparkline := make([]float32, len(snap.PriceHistory))
		for i, p := range snap.PriceHistory {
			sparkline[i] = float32(p)
		}
		bid := float32(snap.Bid)
		ask := float32(snap.Ask)
		last := float32(snap.Last)
		vol := float32(snap.Volume24h)
		marketData = append(marketData, MarketSnapshot{
			Pair:      &snap.Pair,
			Bid:       &bid,
			Ask:       &ask,
			Last:      &last,
			Volume24h: &vol,
			UpdatedAt: &snap.UpdatedAt,
			Sparkline: &sparkline,
		})
	}

	activeCount := active
	staleCount := stale
	erroredCount := errored

	dashboard := DashboardOverview{
		Portfolio: &PortfolioSummary{
			TotalValueUsd: &totalValueUSD,
			Balances:      &balances,
			OpenPositions: &openPositions,
			UpdatedAt:     &updatedAt,
		},
		MarketSnapshot: &marketData,
		SubscriptionHealth: &SubscriptionHealth{
			Active:  &activeCount,
			Stale:   &staleCount,
			Errored: &erroredCount,
			Errors:  &errors,
		},
		RecentDecisions: &recentDecisions,
		RecentAlerts:    nil,
	}

	if len(recentAlerts) > 0 {
		alerts := make([]PriceAlert, len(recentAlerts))
		for i, a := range recentAlerts {
			pair := a["pair"].(string)
			changePct := a["change_pct"].(float64)
			current := a["current"].(float64)
			previous := a["previous"].(float64)
			updatedAt := a["updated_at"].(time.Time)
			direction := a["direction"].(string)
			cpct := float32(changePct)
			curr := float32(current)
			prev := float32(previous)
			alerts[i] = PriceAlert{
				Pair:      &pair,
				ChangePct: &cpct,
				Current:   &curr,
				Previous:  &prev,
				UpdatedAt: &updatedAt,
				Direction: &direction,
			}
		}
		dashboard.RecentAlerts = &alerts
	}

	c.JSON(http.StatusOK, dashboard)
}

func (s *Server) GetSignals(c *gin.Context) {
	snapshots := s.stateMgr.GetAllMarketSnapshots()
	signals := make([]SignalSummary, 0, len(snapshots))

	for _, snap := range snapshots {
		if snap.MomentumSignal != "" || snap.BreakoutSignal != "" || snap.VolumeSignal != "" {
			sym := snap.Pair
			momentum := snap.MomentumSignal
			breakout := snap.BreakoutSignal
			volume := snap.VolumeSignal
			updatedAt := snap.UpdatedAt
			signals = append(signals, SignalSummary{
				Symbol:         &sym,
				MomentumSignal: &momentum,
				BreakoutSignal: &breakout,
				VolumeSignal:   &volume,
				UpdatedAt:      &updatedAt,
			})
		}
	}

	c.JSON(http.StatusOK, gin.H{"signals": signals, "count": len(signals)})
}

func (s *Server) GetNews(c *gin.Context) {
	articles := s.stateMgr.GetNews()
	news := make([]NewsItem, len(articles))
	for i, a := range articles {
		id := a.ID
		title := a.Title
		summary := a.Summary
		source := a.Source
		ts := a.Timestamp
		news[i] = NewsItem{
			Id:        &id,
			Title:     &title,
			Summary:   &summary,
			Source:    &source,
			Timestamp: &ts,
		}
	}

	c.JSON(http.StatusOK, gin.H{"news": news, "count": len(news)})
}

func (s *Server) GetTrades(c *gin.Context, params GetTradesParams) {
	limit := 50
	if params.Limit != nil {
		limit = *params.Limit
		if limit < 1 || limit > 500 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "limit must be between 1 and 500"})
			return
		}
	}

	if s.storageClient == nil {
		c.JSON(http.StatusOK, gin.H{"trades": []TradeRecord{}, "count": 0})
		return
	}

	trades, err := s.storageClient.GetTradeHistory(c.Request.Context(), limit)
	if err != nil {
		log.Error().Err(err).Msg("Failed to get trade history")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get trade history"})
		return
	}

	result := make([]TradeRecord, len(trades))
	for i, t := range trades {
		pair := t.Pair
		action := t.Action
		orderType := t.OrderType
		size := float32(t.Size)
		price := float32(t.Price)
		cost := float32(t.Cost)
		fee := float32(t.Fee)
		mode := t.Mode
		reasoning := t.Reasoning
		confidence := float32(t.Confidence)
		ts := t.Timestamp
		result[i] = TradeRecord{
			Pair:       &pair,
			Action:     &action,
			OrderType:  &orderType,
			Size:       &size,
			Price:      &price,
			Cost:       &cost,
			Fee:        &fee,
			Mode:       &mode,
			Reasoning:  &reasoning,
			Confidence: &confidence,
			Timestamp:  &ts,
		}
	}

	c.JSON(http.StatusOK, gin.H{"trades": result, "count": len(result)})
}
