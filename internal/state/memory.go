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

// UpdateTick updates the live price from a WebSocket stream in microseconds
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
