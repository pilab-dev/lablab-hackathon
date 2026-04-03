package market

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeSymbol(t *testing.T) {
	tests := []struct {
		input    string
		cache    map[string]string
		expected string
	}{
		{"ETH/USD", nil, "ETH/USD"},
		{"eth/usd", nil, "ETH/USD"},
		{"ETHUSD", nil, "ETH/USD"},
		{"BTCUSDT", nil, "BTC/USDT"},
		{"BTC/USDT", nil, "BTC/USDT"},
		{"ETH", nil, "ETH/USD"},
		{"BTC", nil, "BTC/USD"},
		{"XBT/USD", nil, "XBT/USD"},
		{"1INCH/USD", nil, "1INCH/USD"},
		{"SOLUSDC", nil, "SOL/USDC"},
		{"ETH", map[string]string{"ETH": "ETH/EUR"}, "ETH/EUR"},
		{"  eth  ", nil, "ETH/USD"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := NormalizeSymbol(tt.input, tt.cache)
			assert.Equal(t, tt.expected, result)
		})
	}
}
