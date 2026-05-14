package webapi

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// MethodPaths maps logical WebAPI methods to URL paths under Client.Base.
// Defaults match FC documentation shape. Override in tests or when FC
// publishes a different path.
type MethodPaths struct {
	Token         string
	DataSnapshot  string
	GetMatchByID  string
}

// DefaultPaths returns the FC-documented paths.
func DefaultPaths() MethodPaths {
	return MethodPaths{
		Token:        "/Token",
		DataSnapshot: "/DataSnapshot",
		GetMatchByID: "/GetMatchByID",
	}
}

// Options tunes a Client.
type Options struct {
	HTTPClient *http.Client
	Paths      MethodPaths

	// TokenTTL is the cache lifetime. FC tokens last 24h; we refresh
	// proactively before expiry. Default: 23h.
	TokenTTL time.Duration

	// Now overrides time.Now for tests.
	Now func() time.Time
}

// Client talks to a FeedConstruct WebAPI endpoint.
type Client struct {
	Base string
	User string
	Pass string

	httpc *http.Client
	paths MethodPaths
	ttl   time.Duration
	now   func() time.Time

	tokenMu     sync.Mutex
	cachedTok   string
	tokenExp    time.Time
	refreshing  bool
	refreshDone chan struct{}
	refreshErr  error
}

// New returns a Client. Base / User / Pass come from FC_API_BASE etc.
func New(base, user, pass string, opts Options) *Client {
	if opts.HTTPClient == nil {
		opts.HTTPClient = &http.Client{Timeout: 20 * time.Second}
	}
	if opts.Paths == (MethodPaths{}) {
		opts.Paths = DefaultPaths()
	}
	if opts.TokenTTL == 0 {
		opts.TokenTTL = 23 * time.Hour
	}
	if opts.Now == nil {
		opts.Now = time.Now
	}
	return &Client{
		Base:  strings.TrimRight(base, "/"),
		User:  user,
		Pass:  pass,
		httpc: opts.HTTPClient,
		paths: opts.Paths,
		ttl:   opts.TokenTTL,
		now:   opts.Now,
	}
}

// Token returns a cached token or refreshes it. Concurrent callers see
// at most one in-flight refresh (singleflight).
func (c *Client) Token(ctx context.Context) (string, error) {
	c.tokenMu.Lock()
	if c.cachedTok != "" && c.now().Before(c.tokenExp) {
		t := c.cachedTok
		c.tokenMu.Unlock()
		return t, nil
	}
	if c.refreshing {
		ch := c.refreshDone
		c.tokenMu.Unlock()
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-ch:
		}
		c.tokenMu.Lock()
		defer c.tokenMu.Unlock()
		if c.refreshErr != nil {
			return "", c.refreshErr
		}
		return c.cachedTok, nil
	}
	c.refreshing = true
	c.refreshDone = make(chan struct{})
	c.tokenMu.Unlock()

	tok, exp, err := c.fetchToken(ctx)

	c.tokenMu.Lock()
	defer c.tokenMu.Unlock()
	c.refreshing = false
	c.refreshErr = err
	if err == nil {
		c.cachedTok = tok
		c.tokenExp = exp
	}
	close(c.refreshDone)
	if err != nil {
		return "", err
	}
	return tok, nil
}

type tokenResponse struct {
	Token string `json:"token,omitempty"`
	// FC sometimes wraps in {"Token":"..."} (capital T).
	TokenCap string `json:"Token,omitempty"`
}

func (c *Client) fetchToken(ctx context.Context) (string, time.Time, error) {
	body, _ := json.Marshal(map[string]string{
		"username": c.User,
		"password": c.Pass,
	})
	req, err := c.newRequest(ctx, c.paths.Token, body)
	if err != nil {
		return "", time.Time{}, err
	}
	respBody, err := c.do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("webapi: token: %w", err)
	}
	var tr tokenResponse
	if err := json.Unmarshal(respBody, &tr); err != nil {
		// FC may return the bare string token, accept that too.
		s := strings.TrimSpace(string(respBody))
		if len(s) > 1 && s[0] == '"' && s[len(s)-1] == '"' {
			s = s[1 : len(s)-1]
		}
		if s != "" && !strings.Contains(s, "{") {
			return s, c.now().Add(c.ttl), nil
		}
		return "", time.Time{}, fmt.Errorf("webapi: token: parse: %w", err)
	}
	tok := tr.Token
	if tok == "" {
		tok = tr.TokenCap
	}
	if tok == "" {
		return "", time.Time{}, errors.New("webapi: token: empty token in response")
	}
	return tok, c.now().Add(c.ttl), nil
}

// SnapshotResult is the raw JSON body of a DataSnapshot call.
type SnapshotResult struct {
	IsLive       bool
	GeneratedAt  time.Time
	ChangesFrom  *time.Time
	BodyJSON     []byte
}

// DataSnapshot fetches a snapshot. When changesFrom is nil the call is a
// full snapshot; otherwise it asks for incremental changes since that time.
func (c *Client) DataSnapshot(ctx context.Context, isLive bool, changesFrom *time.Time) (SnapshotResult, error) {
	tok, err := c.Token(ctx)
	if err != nil {
		return SnapshotResult{}, err
	}
	payload := map[string]any{
		"token":  tok,
		"isLive": isLive,
	}
	if changesFrom != nil {
		payload["getChangesFrom"] = changesFrom.UTC().Format(time.RFC3339)
	}
	body, _ := json.Marshal(payload)

	req, err := c.newRequest(ctx, c.paths.DataSnapshot, body)
	if err != nil {
		return SnapshotResult{}, err
	}
	respBody, err := c.do(req)
	if err != nil {
		return SnapshotResult{}, fmt.Errorf("webapi: data snapshot: %w", err)
	}
	return SnapshotResult{
		IsLive:      isLive,
		GeneratedAt: c.now(),
		ChangesFrom: changesFrom,
		BodyJSON:    respBody,
	}, nil
}

// GetMatchByID fetches a single match payload for event-scope recovery.
func (c *Client) GetMatchByID(ctx context.Context, matchID int64) ([]byte, error) {
	tok, err := c.Token(ctx)
	if err != nil {
		return nil, err
	}
	body, _ := json.Marshal(map[string]any{
		"token":   tok,
		"matchId": matchID,
	})
	req, err := c.newRequest(ctx, c.paths.GetMatchByID, body)
	if err != nil {
		return nil, err
	}
	respBody, err := c.do(req)
	if err != nil {
		return nil, fmt.Errorf("webapi: get match %d: %w", matchID, err)
	}
	return respBody, nil
}

// RateLimitedError is returned when the server replies 429 so callers can
// schedule retries with exponential backoff.
type RateLimitedError struct {
	RetryAfter time.Duration
}

func (e *RateLimitedError) Error() string {
	return fmt.Sprintf("webapi: rate limited (retry after %s)", e.RetryAfter)
}

func (c *Client) newRequest(ctx context.Context, path string, body []byte) (*http.Request, error) {
	url := c.Base + path
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")
	return req, nil
}

func (c *Client) do(req *http.Request) ([]byte, error) {
	resp, err := c.httpc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		ra := parseRetryAfter(resp.Header.Get("Retry-After"))
		_, _ = io.Copy(io.Discard, resp.Body)
		return nil, &RateLimitedError{RetryAfter: ra}
	}
	body, err := decompressResponse(resp)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, fmt.Errorf("status %d: %s", resp.StatusCode, snippet(body))
	}
	return body, nil
}

func parseRetryAfter(h string) time.Duration {
	if h == "" {
		return 5 * time.Second
	}
	// Honor only the simple "seconds" form; HTTP-date form is rare for APIs.
	var n int
	for _, ch := range h {
		if ch < '0' || ch > '9' {
			break
		}
		n = n*10 + int(ch-'0')
	}
	if n <= 0 {
		return 5 * time.Second
	}
	return time.Duration(n) * time.Second
}

func snippet(b []byte) string {
	const max = 240
	if len(b) <= max {
		return string(b)
	}
	return string(b[:max]) + "...(truncated)"
}
