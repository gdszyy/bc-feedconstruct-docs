package webapi_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/webapi"
)

func readJSON(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	b, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	defer r.Body.Close()
	var out map[string]any
	require.NoError(t, json.Unmarshal(b, &out))
	return out
}

func gzipJSON(t *testing.T, payload any) []byte {
	t.Helper()
	plain, err := json.Marshal(payload)
	require.NoError(t, err)
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err = w.Write(plain)
	require.NoError(t, err)
	require.NoError(t, w.Close())
	return buf.Bytes()
}

// Given valid FC_API_USER and FC_API_PASS
// When Token() is called for the first time
// Then a non-empty token is returned and cached so the second call
//      issues no additional Token request
func TestGiven_ValidCreds_When_TokenCalled_Then_NonEmptyAndCached(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/Token", r.URL.Path)
		atomic.AddInt32(&hits, 1)
		body := readJSON(t, r)
		require.Equal(t, "u", body["username"])
		require.Equal(t, "p", body["password"])
		_, _ = io.WriteString(w, `{"token":"abc-123"}`)
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "p", webapi.Options{})
	tok, err := c.Token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "abc-123", tok)

	// Second call within TTL must not hit the server.
	tok2, err := c.Token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "abc-123", tok2)
	require.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

// Given a token near expiry under concurrent callers
// When several goroutines call Token()
// Then exactly one HTTP request is in flight (singleflight)
func TestGiven_ConcurrentCallers_When_TokenRefresh_Then_SingleInFlight(t *testing.T) {
	var hits int32
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		<-release // hold response until all callers have arrived
		_, _ = io.WriteString(w, `{"token":"xyz"}`)
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "p", webapi.Options{})

	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := c.Token(context.Background())
			require.NoError(t, err)
		}()
	}
	time.Sleep(80 * time.Millisecond) // allow goroutines to register
	close(release)
	wg.Wait()

	require.Equal(t, int32(1), atomic.LoadInt32(&hits))
}

// Given the WebAPI returns 401 Unauthorized
// When Token() is called
// Then the error is wrapped without leaking the password to logs
func TestGiven_AuthError_When_Token_Then_ErrorWrappedAndCredentialsNotLogged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "bad credentials", http.StatusUnauthorized)
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "supersecret", webapi.Options{})
	_, err := c.Token(context.Background())
	require.Error(t, err)
	require.False(t, strings.Contains(err.Error(), "supersecret"),
		"password must not appear in error: %q", err.Error())
	require.Contains(t, err.Error(), "status 401")
}

// Given a cached token within TTL
// When DataSnapshot is called without changesFrom
// Then it issues a snapshot POST with isLive=true and no getChangesFrom
func TestGiven_CachedToken_When_FullSnapshot_Then_NoChangesFromInBody(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/Token":
			_, _ = io.WriteString(w, `{"token":"t"}`)
		case "/DataSnapshot":
			body := readJSON(t, r)
			require.Equal(t, "t", body["token"])
			require.Equal(t, true, body["isLive"])
			_, hasCF := body["getChangesFrom"]
			require.False(t, hasCF, "full snapshot must not carry getChangesFrom")
			w.Header().Set("Content-Encoding", "gzip")
			_, _ = w.Write(gzipJSON(t, map[string]any{"matches": []any{}}))
		default:
			t.Fatalf("unexpected path %s", r.URL.Path)
		}
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "p", webapi.Options{})
	res, err := c.DataSnapshot(context.Background(), true, nil)
	require.NoError(t, err)
	require.True(t, res.IsLive)
	require.Contains(t, string(res.BodyJSON), `"matches"`)
}

// Given the server returns 429 with Retry-After
// When DataSnapshot is called
// Then a RateLimitedError is returned carrying the retry-after seconds
func TestGiven_RateLimited_When_DataSnapshot_Then_RateLimitedError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Token" {
			_, _ = io.WriteString(w, `{"token":"t"}`)
			return
		}
		w.Header().Set("Retry-After", "7")
		http.Error(w, "slow down", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "p", webapi.Options{})
	_, err := c.DataSnapshot(context.Background(), true, nil)
	require.Error(t, err)
	var rl *webapi.RateLimitedError
	for cur := err; cur != nil; {
		if asRL, ok := cur.(*webapi.RateLimitedError); ok {
			rl = asRL
			break
		}
		type unwrapper interface{ Unwrap() error }
		if u, ok := cur.(unwrapper); ok {
			cur = u.Unwrap()
		} else {
			break
		}
	}
	require.NotNil(t, rl, "expected RateLimitedError in error chain")
	require.Equal(t, 7*time.Second, rl.RetryAfter)
}

// Given an incremental call
// When DataSnapshot is called with a changesFrom timestamp
// Then the request body carries getChangesFrom as an RFC3339 string
func TestGiven_ChangesFrom_When_IncrementalSnapshot_Then_RFC3339Sent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Token" {
			_, _ = io.WriteString(w, `{"token":"t"}`)
			return
		}
		body := readJSON(t, r)
		cf, ok := body["getChangesFrom"].(string)
		require.True(t, ok, "expected getChangesFrom string in %v", body)
		_, err := time.Parse(time.RFC3339, cf)
		require.NoError(t, err, "expected RFC3339, got %q", cf)
		_, _ = w.Write([]byte(`{"matches":[]}`))
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "p", webapi.Options{})
	cf := time.Date(2026, 5, 14, 12, 0, 0, 0, time.UTC)
	_, err := c.DataSnapshot(context.Background(), false, &cf)
	require.NoError(t, err)
}

// Given GetMatchByID is called for a specific id
// When the server returns a JSON body
// Then it surfaces unchanged to the caller
func TestGiven_GetMatchByID_When_Called_Then_BodyReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/Token" {
			_, _ = io.WriteString(w, `{"token":"t"}`)
			return
		}
		require.Equal(t, "/GetMatchByID", r.URL.Path)
		body := readJSON(t, r)
		require.EqualValues(t, 42, body["matchId"])
		_, _ = fmt.Fprintln(w, `{"matchId":42,"name":"A vs B"}`)
	}))
	defer srv.Close()

	c := webapi.New(srv.URL, "u", "p", webapi.Options{})
	body, err := c.GetMatchByID(context.Background(), 42)
	require.NoError(t, err)
	require.Contains(t, string(body), `"name":"A vs B"`)
}
