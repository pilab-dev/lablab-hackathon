package kraken

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"

	"github.com/rs/zerolog/log"
)

// DefaultTimeout is the default time limit for kraken-cli commands
const DefaultTimeout = 30 * time.Second

// ErrorEnvelope matches the standard kraken-cli JSON error structure
type ErrorEnvelope struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Client wraps the kraken-cli executable
type Client struct {
	binPath string
}

// NewClient creates a new Kraken CLI wrapper.
// If binPath is empty, it assumes "kraken" is in the system PATH.
func NewClient(binPath string) *Client {
	if binPath == "" {
		binPath = "kraken"
	}
	return &Client{
		binPath: binPath,
	}
}

// RunRaw executes a kraken-cli command and returns the raw JSON bytes.
// It automatically appends "-o json" to the arguments.
func (c *Client) RunRaw(ctx context.Context, args ...string) ([]byte, error) {
	// Ensure we always request JSON output
	fullArgs := append(args, "-o", "json")

	cmd := exec.CommandContext(ctx, c.binPath, fullArgs...)

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// If the command failed, or even if it succeeded but we want to check for API errors
	if err != nil {
		// Try to parse the stdout as a structured error envelope
		if stdout.Len() > 0 {
			var errEnv ErrorEnvelope
			if jsonErr := json.Unmarshal(stdout.Bytes(), &errEnv); jsonErr == nil && errEnv.Error != "" {
				return nil, fmt.Errorf("kraken-cli error [%s]: %s", errEnv.Error, errEnv.Message)
			}
		}
		// Fallback error reporting
		return nil, fmt.Errorf("command execution failed: %w, stderr: %s", err, stderr.String())
	}

	// Even on success (exit code 0), check for error envelope in response body
	if stdout.Len() > 0 {
		var errEnv ErrorEnvelope
		if jsonErr := json.Unmarshal(stdout.Bytes(), &errEnv); jsonErr == nil && errEnv.Error != "" {
			return nil, fmt.Errorf("kraken-cli error [%s]: %s", errEnv.Error, errEnv.Message)
		}
	}

	return stdout.Bytes(), nil
}

// Run executes a kraken-cli command and unmarshals the JSON output into the target interface.
func (c *Client) Run(ctx context.Context, target interface{}, args ...string) error {
	out, err := c.RunRaw(ctx, args...)
	if err != nil {
		return err
	}

	// Unmarshal into the provided struct
	if err := json.Unmarshal(out, target); err != nil {
		return fmt.Errorf("failed to parse kraken-cli json response: %w\nRaw output: %s", err, string(out))
	}

	return nil
}

// RunStream executes a long-running kraken-cli command (like WebSockets)
// and streams the NDJSON output line-by-line to a callback function.
func (c *Client) RunStream(ctx context.Context, callback func([]byte), args ...string) error {
	fullArgs := append(args, "-o", "json")

	cmd := exec.CommandContext(ctx, c.binPath, fullArgs...)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return fmt.Errorf("failed to get stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return fmt.Errorf("failed to start streaming command: %w", err)
	}

	// Read line by line asynchronously
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Bytes()
			// Skip empty lines
			if len(bytes.TrimSpace(line)) > 0 {
				callback(line)
			}
		}

		// When scanner ends, wait for the command to finish
		if err := cmd.Wait(); err != nil {
			log.Error().Err(err).Msg("Command wait failed")
		}
	}()

	return nil
}

type BalanceResponse struct {
	Error  interface{}        `json:"error"`
	Result map[string]float64 `json:"result,omitempty"`
}

func (c *Client) GetBalance(ctx context.Context) (map[string]float64, error) {
	var resp BalanceResponse
	if err := c.Run(ctx, &resp, "balance"); err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("balance error: %v", resp.Error)
	}

	return resp.Result, nil
}

func (c *Client) GetAssets(ctx context.Context) (map[string]map[string]interface{}, error) {
	log.Info().Msg("Kraken GetAssets: calling 'assets' command")

	raw, err := c.RunRaw(ctx, "assets")
	if err != nil {
		log.Error().Err(err).Msg("Kraken GetAssets: RunRaw failed")
		return nil, err
	}

	log.Info().Str("raw", string(raw)[:200]).Msg("Kraken GetAssets: raw result")

	var result map[string]map[string]interface{}
	if err := json.Unmarshal(raw, &result); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal assets")
		return nil, err
	}

	log.Info().Int("count", len(result)).Msg("Got assets from kraken")

	return result, nil
}

func (c *Client) GetPairs(ctx context.Context) (map[string]map[string]any, error) {
	log.Info().Msg("Kraken GetPairs: calling 'pairs' command")

	raw, err := c.RunRaw(ctx, "pairs")
	if err != nil {
		log.Error().Err(err).Msg("Kraken GetPairs: RunRaw failed")
		return nil, err
	}

	log.Info().Str("raw", string(raw)[:200]).Msg("Kraken GetPairs: raw result")

	var result map[string]map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		log.Error().Err(err).Msg("Failed to unmarshal pairs")
		return nil, err
	}

	log.Info().Int("count", len(result)).Msg("Got pairs from kraken")

	return result, nil
}

type OrderResponse struct {
	Error  interface{} `json:"error"`
	Result struct {
		TxID           string `json:"txid"`
		ClOrdID        string `json:"cl_ord_id,omitempty"`
		OrderEvents    []any  `json:"orderEvents,omitempty"`
		Desc           string `json:"descr"`
		Params         any    `json:"params,omitempty"`
		CanceledOrders []any  `json:"canceledOrders,omitempty"`
	} `json:"result,omitempty"`
}

type OrderParams struct {
	Pair        string `json:"pair"`
	Side        string `json:"side"`
	Type        string `json:"type"`
	OrderType   string `json:"ordertype"`
	Volume      string `json:"volume"`
	Price       string `json:"price,omitempty"`
	Price2      string `json:"price2,omitempty"`
	DisplayVol  string `json:"displayvol,omitempty"`
	Trigger     string `json:"trigger,omitempty"`
	Leverage    string `json:"leverage,omitempty"`
	ReduceOnly  bool   `json:"reduce_only,omitempty"`
	TimeInForce string `json:"timeinforce,omitempty"`
	StartTime   string `json:"start_time,omitempty"`
	ExpireTime  string `json:"expire_time,omitempty"`
	UserRef     int32  `json:"userref,omitempty"`
	ClOrdID     string `json:"cl_ord_id,omitempty"`
	OFlags      string `json:"oflags,omitempty"`
	STPType     string `json:"stptype,omitempty"`
	Validate    bool   `json:"validate,omitempty"`
	AssetClass  string `json:"asset_class,omitempty"`
}

func (c *Client) OrderBuy(ctx context.Context, params OrderParams) (*OrderResponse, error) {
	params.Side = "buy"
	if params.OrderType == "" {
		params.OrderType = "limit"
	}
	return c.orderExecute(ctx, "order", "buy", params)
}

func (c *Client) OrderSell(ctx context.Context, params OrderParams) (*OrderResponse, error) {
	params.Side = "sell"
	if params.OrderType == "" {
		params.OrderType = "limit"
	}
	return c.orderExecute(ctx, "order", "sell", params)
}

func (c *Client) orderExecute(ctx context.Context, command, subcommand string, params OrderParams) (*OrderResponse, error) {
	args := []string{command, subcommand, params.Pair, params.Volume}

	if params.Type != "" {
		args = append(args, "--type", params.Type)
	}
	if params.OrderType != "" && params.OrderType != "limit" {
		args = append(args, "--type", params.OrderType)
	}
	if params.Price != "" {
		args = append(args, "--price", params.Price)
	}
	if params.Price2 != "" {
		args = append(args, "--price2", params.Price2)
	}
	if params.DisplayVol != "" {
		args = append(args, "--displayvol", params.DisplayVol)
	}
	if params.Trigger != "" {
		args = append(args, "--trigger", params.Trigger)
	}
	if params.Leverage != "" {
		args = append(args, "--leverage", params.Leverage)
	}
	if params.ReduceOnly {
		args = append(args, "--reduce-only")
	}
	if params.TimeInForce != "" {
		args = append(args, "--timeinforce", params.TimeInForce)
	}
	if params.StartTime != "" {
		args = append(args, "--start-time", params.StartTime)
	}
	if params.ExpireTime != "" {
		args = append(args, "--expire-time", params.ExpireTime)
	}
	if params.UserRef != 0 {
		args = append(args, "--userref", fmt.Sprintf("%d", params.UserRef))
	}
	if params.ClOrdID != "" {
		args = append(args, "--cl-ord-id", params.ClOrdID)
	}
	if params.OFlags != "" {
		args = append(args, "--oflags", params.OFlags)
	}
	if params.STPType != "" {
		args = append(args, "--stptype", params.STPType)
	}
	if params.Validate {
		args = append(args, "--validate")
	}
	if params.AssetClass != "" {
		args = append(args, "--asset-class", params.AssetClass)
	}

	var resp OrderResponse
	err := c.Run(ctx, &resp, args...)
	if err != nil {
		return nil, err
	}

	if resp.Error != nil {
		return nil, fmt.Errorf("order error: %v", resp.Error)
	}

	return &resp, nil
}

type AmendParams struct {
	TxID         string `json:"txid,omitempty"`
	ClOrdID      string `json:"cl_ord_id,omitempty"`
	OrderQty     string `json:"order_qty,omitempty"`
	DisplayQty   string `json:"display_qty,omitempty"`
	LimitPrice   string `json:"limit_price,omitempty"`
	TriggerPrice string `json:"trigger_price,omitempty"`
	Pair         string `json:"pair,omitempty"`
	PostOnly     bool   `json:"post_only,omitempty"`
}

func (c *Client) OrderAmend(ctx context.Context, params AmendParams) (*OrderResponse, error) {
	args := []string{"order", "amend"}

	if params.TxID != "" {
		args = append(args, "--txid", params.TxID)
	}
	if params.ClOrdID != "" {
		args = append(args, "--cl-ord-id", params.ClOrdID)
	}
	if params.OrderQty != "" {
		args = append(args, "--order-qty", params.OrderQty)
	}
	if params.DisplayQty != "" {
		args = append(args, "--display-qty", params.DisplayQty)
	}
	if params.LimitPrice != "" {
		args = append(args, "--limit-price", params.LimitPrice)
	}
	if params.TriggerPrice != "" {
		args = append(args, "--trigger-price", params.TriggerPrice)
	}
	if params.Pair != "" {
		args = append(args, "--pair", params.Pair)
	}
	if params.PostOnly {
		args = append(args, "--post-only")
	}

	var resp OrderResponse
	err := c.Run(ctx, &resp, args...)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("amend error: %v", resp.Error)
	}
	return &resp, nil
}

type EditParams struct {
	TXID   string `json:"txid"`
	Volume string `json:"volume,omitempty"`
	Price  string `json:"price,omitempty"`
	Pair   string `json:"pair,omitempty"`
}

func (c *Client) OrderEdit(ctx context.Context, params EditParams) (*OrderResponse, error) {
	args := []string{"order", "edit", params.TXID}

	if params.Volume != "" {
		args = append(args, "--volume", params.Volume)
	}
	if params.Price != "" {
		args = append(args, "--price", params.Price)
	}
	if params.Pair != "" {
		args = append(args, "--pair", params.Pair)
	}

	var resp OrderResponse
	err := c.Run(ctx, &resp, args...)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("edit error: %v", resp.Error)
	}
	return &resp, nil
}

func (c *Client) OrderCancel(ctx context.Context, txids []string, clOrdID string) (*OrderResponse, error) {
	args := []string{"order", "cancel"}

	if clOrdID != "" {
		args = append(args, "--cl-ord-id", clOrdID)
	}
	args = append(args, txids...)

	var resp OrderResponse
	err := c.Run(ctx, &resp, args...)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("cancel error: %v", resp.Error)
	}
	return &resp, nil
}

func (c *Client) OrderCancelAll(ctx context.Context) (*OrderResponse, error) {
	args := []string{"order", "cancel-all"}

	var resp OrderResponse
	err := c.Run(ctx, &resp, args...)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("cancel-all error: %v", resp.Error)
	}
	return &resp, nil
}

func (c *Client) OrderCancelAfter(ctx context.Context, timeoutSeconds int) (*OrderResponse, error) {
	args := []string{"order", "cancel-after", fmt.Sprintf("%d", timeoutSeconds)}

	var resp OrderResponse
	err := c.Run(ctx, &resp, args...)
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, fmt.Errorf("cancel-after error: %v", resp.Error)
	}
	return &resp, nil
}
