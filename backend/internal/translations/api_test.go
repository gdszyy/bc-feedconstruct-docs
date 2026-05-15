package translations_test

import (
	"bytes"
	"compress/gzip"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/translations"
)

// Given the Languages endpoint returns gzip-compressed JSON
// When the client decodes the response
// Then the LanguageModel list is correctly parsed.
func TestGiven_GzippedLanguagesResponse_When_Decoded_Then_LanguagesParsed(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/Translation/Languages", r.URL.Path)
		require.Equal(t, "gzip", r.Header.Get("Accept-Encoding"))

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Content-Encoding", "gzip")
		gz := gzip.NewWriter(w)
		_, _ = gz.Write([]byte(`[
			{"Id":"en","Name":"English","IsDefault":true},
			{"Id":"zh","Name":"Chinese","IsDefault":false}
		]`))
		require.NoError(t, gz.Close())
	}))
	defer srv.Close()

	c := translations.NewClient(translations.ClientOptions{BaseURL: srv.URL})
	got, err := c.Languages(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "en", got[0].ID)
	require.Equal(t, "English", got[0].Name)
	require.True(t, got[0].IsDefault)
	require.Equal(t, "zh", got[1].ID)
}

// Given the ByLanguage endpoint returns a TranslationResponseModel
// When the client decodes it
// Then every TranslationModel is mapped onto the local Translation type
func TestGiven_ByLanguageResponse_When_Decoded_Then_TranslationsMapped(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/Translation/ByLanguage/en", r.URL.Path)
		_, _ = w.Write([]byte(`{
			"LanguageId":"en",
			"Translations":[
				{"Id":1,"Text":"Match Result"},
				{"Id":2,"Text":"Total Goals"}
			]
		}`))
	}))
	defer srv.Close()

	c := translations.NewClient(translations.ClientOptions{BaseURL: srv.URL})
	got, err := c.ByLanguage(context.Background(), "en")
	require.NoError(t, err)
	require.Len(t, got, 2)
	require.Equal(t, "en", got[0].LanguageID)
	require.Equal(t, int64(1), got[0].TranslationID)
	require.Equal(t, "Match Result", got[0].Text)
}

// Given the ById endpoint returns one TranslationModel
// When the client decodes it
// Then the single Translation is returned with ok=true
func TestGiven_ByIdResponse_When_Decoded_Then_SingleTranslation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.Equal(t, "/api/Translation/ById/en", r.URL.Path)
		require.Equal(t, "42", r.URL.Query().Get("id"))
		_, _ = w.Write([]byte(`{
			"LanguageId":"en",
			"Translations":[{"Id":42,"Text":"Both Teams To Score"}]
		}`))
	}))
	defer srv.Close()

	c := translations.NewClient(translations.ClientOptions{BaseURL: srv.URL})
	got, ok, err := c.ByID(context.Background(), "en", 42)
	require.NoError(t, err)
	require.True(t, ok)
	require.Equal(t, "Both Teams To Score", got.Text)
	require.Equal(t, int64(42), got.TranslationID)
	require.Equal(t, "en", got.LanguageID)
}

// Given the ById endpoint returns an empty translations array
// When the client decodes it
// Then (Translation{}, false, nil) is returned
func TestGiven_ByIdEmptyResponse_When_Decoded_Then_NotFound(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"LanguageId":"en","Translations":[]}`))
	}))
	defer srv.Close()

	c := translations.NewClient(translations.ClientOptions{BaseURL: srv.URL})
	_, ok, err := c.ByID(context.Background(), "en", 9999)
	require.NoError(t, err)
	require.False(t, ok)
}

// Given the upstream responds with HTTP 429 and Retry-After
// When the client issues any GET
// Then the typed HTTPError is returned and IsRateLimited reports true
func TestGiven_RateLimitedUpstream_When_Called_Then_RateLimitErrorReturned(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Retry-After", "120")
		http.Error(w, "rate limited", http.StatusTooManyRequests)
	}))
	defer srv.Close()

	c := translations.NewClient(translations.ClientOptions{BaseURL: srv.URL})
	_, err := c.Languages(context.Background())
	require.Error(t, err)
	require.True(t, translations.IsRateLimited(err))
}

// Sanity test for non-gzipped responses (some test fixtures don't enable gzip).
func TestGiven_PlainResponse_When_Decoded_Then_StillWorks(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		var buf bytes.Buffer
		buf.WriteString(`[{"Id":"en","Name":"English","IsDefault":true}]`)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(buf.Bytes())
	}))
	defer srv.Close()

	c := translations.NewClient(translations.ClientOptions{BaseURL: srv.URL})
	got, err := c.Languages(context.Background())
	require.NoError(t, err)
	require.Len(t, got, 1)
}
