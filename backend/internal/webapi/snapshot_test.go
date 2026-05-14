package webapi_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/webapi"
)

// FeedConstruct WebAPI DataSnapshot contract — see docs/01_data_feed/rmq-web-api/033_webmethods.md
//
// Given an authenticated client and isLive=true with getChangesFrom=15
// When DataSnapshot is called
// Then GET /api/DataService/DataSnapshot is issued with token, isLive=true,
//      getChangesFrom=15 query params, and the Objects list is returned.
func TestGiven_ValidToken_When_DataSnapshotIsLiveTrue_Then_QueryParamsAndObjectsReturned(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/DataService/Token":
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Token":"tk","ResultCode":0}`))
		case "/api/DataService/DataSnapshot":
			atomic.AddInt32(&calls, 1)
			require.Equal(t, http.MethodGet, r.Method)
			q := r.URL.Query()
			require.Equal(t, "tk", q.Get("token"))
			require.Equal(t, "true", strings.ToLower(q.Get("isLive")))
			require.Equal(t, "15", q.Get("getChangesFrom"))
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Objects":[{"Id":1},{"Id":2}],"ResultCode":0}`))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{BaseURL: srv.URL, Username: "u", Password: "p"})
	objs, err := c.DataSnapshot(context.Background(), true, 15)
	require.NoError(t, err)
	require.Len(t, objs, 2)
	require.EqualValues(t, 1, atomic.LoadInt32(&calls))
}

// Given DataSnapshot is called with getChangesFrom=0
// When the request is built
// Then the getChangesFrom query parameter is OMITTED (full snapshot semantics)
//
// Per FeedConstruct doc: getChangesFrom is optional; without it the full snapshot is returned.
func TestGiven_ZeroGetChangesFrom_When_DataSnapshot_Then_ParamOmitted(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/DataService/Token" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"Token":"tk","ResultCode":0}`))
			return
		}
		q := r.URL.Query()
		require.False(t, q.Has("getChangesFrom"), "getChangesFrom=0 must NOT be serialised")
		_, _ = w.Write([]byte(`{"Objects":[],"ResultCode":0}`))
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{BaseURL: srv.URL, Username: "u", Password: "p"})
	_, err := c.DataSnapshot(context.Background(), false, 0)
	require.NoError(t, err)
}

// Given the WebAPI responds with HTTP 429 Too Many Requests
// When DataSnapshot is called
// Then a webapi.RateLimitError is returned (caller decides backoff)
func TestGiven_HTTP429_When_DataSnapshot_Then_RateLimitErrorReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/DataService/Token" {
			_, _ = w.Write([]byte(`{"Token":"tk","ResultCode":0}`))
			return
		}
		w.Header().Set("Retry-After", "2")
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{BaseURL: srv.URL, Username: "u", Password: "p"})
	_, err := c.DataSnapshot(context.Background(), true, 5)
	require.Error(t, err)
	var rl *webapi.RateLimitError
	require.ErrorAs(t, err, &rl)
	require.Equal(t, 2, rl.RetryAfterSeconds)
}

// Given DataSnapshot is told isLive=false with getChangesFrom=60
// When the request is built
// Then the isLive parameter is serialised as the literal lowercase "false"
func TestGiven_IsLiveFalse_When_DataSnapshot_Then_LowercaseFalseSent(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/DataService/Token" {
			_, _ = w.Write([]byte(`{"Token":"tk","ResultCode":0}`))
			return
		}
		q := r.URL.Query()
		require.Equal(t, "false", q.Get("isLive"))
		require.Equal(t, "60", q.Get("getChangesFrom"))
		// Sanity: numeric is decimal, not float
		_, parseErr := strconv.Atoi(q.Get("getChangesFrom"))
		require.NoError(t, parseErr)
		_, _ = w.Write([]byte(`{"Objects":[],"ResultCode":0}`))
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{BaseURL: srv.URL, Username: "u", Password: "p"})
	_, err := c.DataSnapshot(context.Background(), false, 60)
	require.NoError(t, err)
}
