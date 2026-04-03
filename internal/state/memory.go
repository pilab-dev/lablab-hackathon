package state

import (
	"sync"
	"time"
)

// PairState holds the latest in-memory market data for a single trading pair
type PairState struct {
	Pair      string
	Bid       float64
	Ask       float64
	Last      float64
	Volume24h float64
	UpdatedAt time.Time

	// Track recent price history in RAM for fast trigger calculations
	PriceHistory []float64 // Last 60 seconds of prices

	// PRISM Technical Signals
	MomentumSignal string
	BreakoutSignal string
	VolumeSignal   string
}

// NewsArticle represents a standard news item in memory
type NewsArticle struct {
	ID        string
	Title     string
	Summary   string
	Source    string
	Timestamp time.Time
}

// PortfolioState holds the current account balances
type PortfolioState struct {
	TotalValueUSD float64
	Balances      map[string]float64
	OpenPositions int
	UpdatedAt     time.Time
}

// MemoryManager is the high-performance RAM cache for the trading engine
type MemoryManager struct {
	mu        sync.RWMutex
	market    map[string]*PairState
	portfolio *PortfolioState
	news      []NewsArticle // Ring buffer of recent news

	// Channels for event triggers
	PriceAlertCh chan string // Sends the pair name if a sudden spike occurs
}

// NewMemoryManager initializes the thread-safe state manager
func NewMemoryManager() *MemoryManager {
	return &MemoryManager{
		market:       make(map[string]*PairState),
		portfolio:    &PortfolioState{Balances: make(map[string]float64)},
		PriceAlertCh: make(chan string, 10),
	}
}

// UpdateTick updates the live price from a WebSocket stream
func (m *MemoryManager) UpdateTick(pair string, bid, ask, last float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.market[pair]
	if !exists {
		state = &PairState{
			Pair:         pair,
			PriceHistory: make([]float64, 0, 60),
		}
		m.market[pair] = state
	}

	state.Bid = bid
	state.Ask = ask
	state.Last = last
	state.UpdatedAt = time.Now()

	// Maintain a rolling window of the last 60 ticks
	state.PriceHistory = append(state.PriceHistory, last)
	if len(state.PriceHistory) > 60 {
		state.PriceHistory = state.PriceHistory[1:]
	}

	// Trigger Engine Logic: Check for a sudden price spike (> 0.5% in window)
	if len(state.PriceHistory) == 60 {
		oldest := state.PriceHistory[0]
		changePct := ((last - oldest) / oldest) * 100

		// If price moves more than 0.5% (up or down), fire an alert to wake up the LLM
		if changePct >= 0.5 || changePct <= -0.5 {
			select {
			case m.PriceAlertCh <- pair:
				// Clear history to prevent duplicate alerts firing every millisecond
				state.PriceHistory = make([]float64, 0, 60)
			default:
				// channel full, skip
			}
		}
	}
}

// UpdateSignal updates the technical signals for a specific pair
func (m *MemoryManager) UpdateSignal(pair, momentum, breakout, volume string) {
	// Normalize pair name (e.g., BTC to BTCUSD to match websocket data)
	if len(pair) == 3 {
		pair = pair + "USD"
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	state, exists := m.market[pair]
	if !exists {
		state = &PairState{
			Pair:         pair,
			PriceHistory: make([]float64, 0, 60),
		}
		m.market[pair] = state
	}

	state.MomentumSignal = momentum
	state.BreakoutSignal = breakout
	state.VolumeSignal = volume
	state.UpdatedAt = time.Now()
}

// UpdateNews stores the latest news articles in memory
func (m *MemoryManager) UpdateNews(articles []NewsArticle) {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Keep only the latest 20 articles to prevent memory/context bloat
	if len(articles) > 20 {
		m.news = articles[:20]
	} else {
		m.news = articles
	}
}

// GetMarketSnapshot returns a thread-safe copy of the current market state for the LLM Prompt
func (m *MemoryManager) GetMarketSnapshot(pair string) (PairState, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	state, exists := m.market[pair]
	if !exists {
		return PairState{}, false
	}
	return *state, true
}

// GetPortfolio returns a thread-safe copy of the current portfolio state
func (m *MemoryManager) GetPortfolio() PortfolioState {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return *m.portfolio
}

// UpdateBalance updates the account balances
func (m *MemoryManager) UpdateBalance(balances map[string]float64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.portfolio.Balances = balances
	m.portfolio.UpdatedAt = time.Now()
}

// GetPortfolioSummary returns portfolio with calculated TotalValueUSD
// Uses current market prices to value non-USD balances
func (m *MemoryManager) GetPortfolioSummary() PortfolioState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	balances := make(map[string]float64, len(m.portfolio.Balances))
	for k, v := range m.portfolio.Balances {
		balances[k] = v
	}

	totalUSD := 0.0
	for asset, amount := range balances {
		if asset == "USD" || asset == "ZUSD" || asset == "USDT" || asset == "USDC" {
			totalUSD += amount
			continue
		}

		pairKey := asset + "USD"
		if state, exists := m.market[pairKey]; exists && state.Last > 0 {
			totalUSD += amount * state.Last
		}
	}

	return PortfolioState{
		TotalValueUSD: totalUSD,
		Balances:      balances,
		OpenPositions: m.portfolio.OpenPositions,
		UpdatedAt:     m.portfolio.UpdatedAt,
	}
}

// GetAllMarketSnapshots returns a copy of all pair states
func (m *MemoryManager) GetAllMarketSnapshots() map[string]PairState {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]PairState, len(m.market))
	for k, v := range m.market {
		result[k] = *v
	}
	return result
}

// GetNews returns a copy of recent news articles
func (m *MemoryManager) GetNews() []NewsArticle {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make([]NewsArticle, len(m.news))
	copy(result, m.news)
	return result
}

// GetRecentAlerts returns recent price alerts based on price history analysis
func (m *MemoryManager) GetRecentAlerts(limit int) []map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	alerts := make([]map[string]interface{}, 0, limit)
	for _, state := range m.market {
		if len(state.PriceHistory) >= 2 {
			recent := state.PriceHistory[len(state.PriceHistory)-1]
			older := state.PriceHistory[0]
			if older > 0 {
				changePct := ((recent - older) / older) * 100
				if changePct >= 0.5 || changePct <= -0.5 {
					alerts = append(alerts, map[string]interface{}{
						"pair":       state.Pair,
						"change_pct": changePct,
						"current":    recent,
						"previous":   older,
						"updated_at": state.UpdatedAt,
						"direction":  "up",
					})
				}
			}
		}
	}

	if len(alerts) > limit {
		alerts = alerts[:limit]
	}

	return alerts
}
