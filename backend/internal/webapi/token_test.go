package webapi_test

import (
	"context"
	"encoding/json"
	"errors"
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

// FeedConstruct WebAPI Token contract — see docs/01_data_feed/rmq-web-api/002_access.md
//
// Given valid FC_API_USER and FC_API_PASS
// When webapi.Client.Token() is called for the first time
// Then a non-empty token string is returned and cached in memory
func TestGiven_ValidCreds_When_TokenCalled_Then_NonEmptyAndCached(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		require.Equal(t, "/api/DataService/Token", r.URL.Path)
		require.Equal(t, http.MethodPost, r.Method)
		var body struct {
			Params []struct {
				UserName string
				Password string
			}
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		require.Equal(t, "user-x", body.Params[0].UserName)
		require.Equal(t, "pw-y", body.Params[0].Password)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Token":"abc-123","ResultCode":0}`))
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{
		BaseURL:  srv.URL,
		Username: "user-x",
		Password: "pw-y",
	})

	tok, err := c.Token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "abc-123", tok)

	tok2, err := c.Token(context.Background())
	require.NoError(t, err)
	require.Equal(t, "abc-123", tok2)
	require.EqualValues(t, 1, atomic.LoadInt32(&hits), "second call must hit cache, not transport")
}

// Given a cached token approaching its 24h expiry
// When the client refreshes >=1h before expiry
// Then exactly one refresh request is in flight even under concurrent callers
func TestGiven_TokenNearExpiry_When_ConcurrentRefresh_Then_SingleInFlight(t *testing.T) {
	var hits int32
	gate := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		<-gate
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Token":"refreshed","ResultCode":0}`))
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{
		BaseURL:       srv.URL,
		Username:      "user",
		Password:      "pw",
		TokenLifetime: time.Millisecond,
		RefreshBefore: 0,
	})

	const callers = 8
	var wg sync.WaitGroup
	tokens := make([]string, callers)
	errs := make([]error, callers)
	wg.Add(callers)
	for i := 0; i < callers; i++ {
		go func(i int) {
			defer wg.Done()
			tokens[i], errs[i] = c.Token(context.Background())
		}(i)
	}
	time.Sleep(20 * time.Millisecond)
	close(gate)
	wg.Wait()

	for i, e := range errs {
		require.NoError(t, e, "caller %d", i)
		require.Equal(t, "refreshed", tokens[i])
	}
	require.EqualValues(t, 1, atomic.LoadInt32(&hits), "exactly one HTTP refresh under concurrent callers")
}

// Given the WebAPI returns an auth error
// When Token() is called
// Then the error is wrapped with cause and producer_health.detail records "webapi_token_failed"
//      AND no token value is logged
func TestGiven_AuthError_When_Token_Then_ErrorWrappedAndCredentialsNotLogged(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"Error":{"Key":"InvalidUsernamePassword","Message":"Invalid Username and/or password"},"Objects":[],"ResultCode":15}`))
	}))
	defer srv.Close()

	c := webapi.NewClient(webapi.Options{
		BaseURL:  srv.URL,
		Username: "user-x",
		Password: "super-secret-pw",
	})

	_, err := c.Token(context.Background())
	require.Error(t, err)

	var tokenErr *webapi.TokenError
	require.True(t, errors.As(err, &tokenErr), "should be wrapped TokenError")
	require.Equal(t, 15, tokenErr.ResultCode)
	require.Equal(t, "InvalidUsernamePassword", tokenErr.Key)

	msg := err.Error()
	require.NotContains(t, msg, "super-secret-pw")
	require.NotContains(t, msg, "user-x")
	require.True(t, strings.Contains(strings.ToLower(msg), "webapi_token_failed"),
		"error must surface canonical webapi_token_failed marker (got %q)", msg)
}
