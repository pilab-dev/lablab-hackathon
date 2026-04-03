package market

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"kraken-trader/internal/state"
)

type wsTickExpectation struct {
	Pair string  `json:"pair"`
	Bid  float64 `json:"bid"`
	Ask  float64 `json:"ask"`
	Last float64 `json:"last"`
}

type wsHandlerCase struct {
	Name             string             `json:"name"`
	Input            string             `json:"input"`
	PreSubscriptions map[string]bool    `json:"pre_subscriptions"`
	ExpectSubscribed map[string]bool    `json:"expect_is_subscribed"`
	ExpectTick       *wsTickExpectation `json:"expect_tick"`
}

func TestWSHandlerCasesFromJSON(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("testdata", "ws_handler_cases.json"))
	if err != nil {
		t.Fatalf("failed to read ws handler cases json: %v", err)
	}

	var cases []wsHandlerCase
	if err := json.Unmarshal(raw, &cases); err != nil {
		t.Fatalf("failed to unmarshal ws handler cases json: %v", err)
	}
	if len(cases) == 0 {
		t.Fatal("no cases found in ws_handler_cases.json")
	}

	for _, tc := range cases {
		t.Run(tc.Name, func(t *testing.T) {
			mem := state.NewMemoryManager()
			c := &Collector{
				state:         mem,
				subscriptions: make(map[string]bool),
				pairsCache:    make(map[string]string),
			}

			for k, v := range tc.PreSubscriptions {
				c.subscriptions[k] = v
			}

			// Must not panic for any case.
			c.handleWSTick([]byte(tc.Input))

			for sym, want := range tc.ExpectSubscribed {
				got := c.IsSubscribed(sym)
				if got != want {
					t.Fatalf("IsSubscribed(%q): expected %v, got %v (subscriptions=%v)", sym, want, got, c.subscriptions)
				}
			}

			if tc.ExpectTick != nil {
				snap, ok := mem.GetMarketSnapshot(tc.ExpectTick.Pair)
				if !ok {
					t.Fatalf("expected state for pair %q to exist", tc.ExpectTick.Pair)
				}
				if snap.Bid != tc.ExpectTick.Bid || snap.Ask != tc.ExpectTick.Ask || snap.Last != tc.ExpectTick.Last {
					t.Fatalf("unexpected tick for %q: expected bid=%v ask=%v last=%v, got bid=%v ask=%v last=%v",
						tc.ExpectTick.Pair,
						tc.ExpectTick.Bid, tc.ExpectTick.Ask, tc.ExpectTick.Last,
						snap.Bid, snap.Ask, snap.Last,
					)
				}
			}
		})
	}
}
