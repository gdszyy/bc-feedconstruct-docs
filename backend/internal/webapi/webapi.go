// Package webapi is the FeedConstruct WebAPI client (Token, DataSnapshot,
// descriptions, Sport/Region/Competition/MarketTypes/SelectionTypes/EventTypes/Periods).
// See docs/01_data_feed/rmq-web-api/033_webmethods.md for the contract.
package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Options configures Client behaviour. BaseURL/Username/Password are
// required; everything else has documentation-aligned defaults.
type Options struct {
	BaseURL    string
	Username   string
	Password   string
	HTTPClient *http.Client

	// TokenLifetime is how long a freshly minted token is considered
	// valid before a proactive refresh is attempted. FeedConstruct
	// rotates tokens every 24h, so the default is 24h.
	TokenLifetime time.Duration

	// RefreshBefore biases the cache to refresh earlier than the
	// documented expiry, providing slack under clock skew. The
	// FeedConstruct doc recommends ≥1h.
	RefreshBefore time.Duration

	// Now overrides time.Now in tests.
	Now func() time.Time
}

// Client is a thin, allocation-light FeedConstruct WebAPI client. It is
// safe for concurrent use; concurrent Token() refreshes collapse onto a
// single in-flight request.
type Client struct {
	opts Options
	http *http.Client

	mu          sync.Mutex
	token       string
	tokenExpiry time.Time

	refreshMu sync.Mutex
}

// NewClient constructs a configured Client. Required Options are
// validated lazily on the first call so tests can construct freely.
func NewClient(opts Options) *Client {
	if opts.TokenLifetime <= 0 {
		opts.TokenLifetime = 24 * time.Hour
	}
	if opts.RefreshBefore < 0 {
		opts.RefreshBefore = 0
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{opts: opts, http: hc}
}

// TokenError is returned when /api/DataService/Token reports a non-zero
// ResultCode. The Error string deliberately omits credentials and uses
// the canonical "webapi_token_failed" marker so log scrapers can match
// it without grepping URL paths.
type TokenError struct {
	ResultCode int
	Key        string
	Message    string
}

func (e *TokenError) Error() string {
	return fmt.Sprintf("webapi_token_failed: code=%d key=%s msg=%s", e.ResultCode, e.Key, e.Message)
}

// RateLimitError is returned when the WebAPI replies HTTP 429.
type RateLimitError struct {
	Path              string
	RetryAfterSeconds int
}

func (e *RateLimitError) Error() string {
	return fmt.Sprintf("webapi_rate_limited: path=%s retry_after=%ds", e.Path, e.RetryAfterSeconds)
}

type tokenResponse struct {
	Token      string `json:"Token"`
	ResultCode int    `json:"ResultCode"`
	Error      *struct {
		Key     string `json:"Key"`
		Message string `json:"Message"`
	} `json:"Error,omitempty"`
}

// Token returns a valid auth token, refreshing when the cache is empty
// or within RefreshBefore of expiry.
func (c *Client) Token(ctx context.Context) (string, error) {
	if tok, ok := c.cachedToken(); ok {
		return tok, nil
	}
	c.refreshMu.Lock()
	defer c.refreshMu.Unlock()
	if tok, ok := c.cachedToken(); ok {
		return tok, nil
	}
	tok, exp, err := c.fetchToken(ctx)
	if err != nil {
		return "", err
	}
	c.mu.Lock()
	c.token = tok
	c.tokenExpiry = exp
	c.mu.Unlock()
	return tok, nil
}

func (c *Client) cachedToken() (string, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.token == "" {
		return "", false
	}
	deadline := c.tokenExpiry.Add(-c.opts.RefreshBefore)
	if !c.opts.Now().Before(deadline) {
		return "", false
	}
	return c.token, true
}

func (c *Client) fetchToken(ctx context.Context) (string, time.Time, error) {
	payload, _ := json.Marshal(map[string]any{
		"Params": []map[string]string{{
			"UserName": c.opts.Username,
			"Password": c.opts.Password,
		}},
	})
	u := c.opts.BaseURL + "/api/DataService/Token"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, bytes.NewReader(payload))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("webapi_token_failed: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := c.http.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("webapi_token_failed: transport: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("webapi_token_failed: read body: %w", err)
	}
	var tr tokenResponse
	if err := json.Unmarshal(body, &tr); err != nil {
		return "", time.Time{}, fmt.Errorf("webapi_token_failed: decode: %w", err)
	}
	if tr.ResultCode != 0 || tr.Token == "" {
		te := &TokenError{ResultCode: tr.ResultCode}
		if tr.Error != nil {
			te.Key = tr.Error.Key
			te.Message = tr.Error.Message
		}
		return "", time.Time{}, te
	}
	return tr.Token, c.opts.Now().Add(c.opts.TokenLifetime), nil
}

// snapshotResponse mirrors the wire envelope; Objects stays as raw JSON
// so callers can decode against their own Match shape.
type snapshotResponse struct {
	Objects    []json.RawMessage `json:"Objects"`
	ResultCode int               `json:"ResultCode"`
}

// DataSnapshot returns the raw Object list from the WebAPI. When
// getChangesFrom <= 0 the parameter is omitted and the full snapshot is
// returned, per the FeedConstruct doc.
func (c *Client) DataSnapshot(ctx context.Context, isLive bool, getChangesFrom int) ([]json.RawMessage, error) {
	tok, err := c.Token(ctx)
	if err != nil {
		return nil, err
	}
	q := url.Values{}
	q.Set("token", tok)
	if isLive {
		q.Set("isLive", "true")
	} else {
		q.Set("isLive", "false")
	}
	if getChangesFrom > 0 {
		q.Set("getChangesFrom", strconv.Itoa(getChangesFrom))
	}
	full := c.opts.BaseURL + "/api/DataService/DataSnapshot?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return nil, fmt.Errorf("webapi_snapshot_failed: build: %w", err)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("webapi_snapshot_failed: transport: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusTooManyRequests {
		retry, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		return nil, &RateLimitError{Path: "/api/DataService/DataSnapshot", RetryAfterSeconds: retry}
	}
	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("webapi_snapshot_failed: http %d", resp.StatusCode)
	}
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("webapi_snapshot_failed: read body: %w", err)
	}
	var sr snapshotResponse
	if err := json.Unmarshal(body, &sr); err != nil {
		return nil, fmt.Errorf("webapi_snapshot_failed: decode: %w", err)
	}
	if sr.ResultCode != 0 {
		return nil, fmt.Errorf("webapi_snapshot_failed: result_code=%d", sr.ResultCode)
	}
	return sr.Objects, nil
}

// IsRateLimited reports whether err is a RateLimitError.
func IsRateLimited(err error) bool {
	var rl *RateLimitError
	return errors.As(err, &rl)
}
