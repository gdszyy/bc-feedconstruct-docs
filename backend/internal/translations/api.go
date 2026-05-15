package translations

import (
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// ClientOptions configures the Translation Web API client.
type ClientOptions struct {
	// BaseURL is the FeedConstruct Translation API root (e.g.
	// https://translation.feedconstruct.com). Required.
	BaseURL string

	// HTTPClient is reused when set; defaults to a Client with a
	// 30s timeout.
	HTTPClient *http.Client
}

// Client implements API against the FeedConstruct Translation Web API.
// It is safe for concurrent use.
type Client struct {
	base string
	http *http.Client
}

// NewClient constructs a Client from options. The BaseURL trailing
// slash is stripped to keep URL assembly predictable.
func NewClient(opts ClientOptions) *Client {
	hc := opts.HTTPClient
	if hc == nil {
		hc = &http.Client{Timeout: 30 * time.Second}
	}
	return &Client{
		base: strings.TrimRight(opts.BaseURL, "/"),
		http: hc,
	}
}

// HTTPError describes a non-2xx response from the Translation API.
type HTTPError struct {
	Path       string
	StatusCode int
	RetryAfter int
}

// Error implements error.
func (e *HTTPError) Error() string {
	return fmt.Sprintf("translation_api: %s -> http %d (retry_after=%ds)", e.Path, e.StatusCode, e.RetryAfter)
}

// IsRateLimited reports whether err is a 429 from the Translation API.
func IsRateLimited(err error) bool {
	var he *HTTPError
	return errors.As(err, &he) && he.StatusCode == http.StatusTooManyRequests
}

// languageModel mirrors the LanguageModel wire shape.
type languageModel struct {
	Id        string `json:"Id"`
	Name      string `json:"Name"`
	IsDefault bool   `json:"IsDefault"`
}

// translationModel mirrors the TranslationModel wire shape.
type translationModel struct {
	Id   int64  `json:"Id"`
	Text string `json:"Text"`
}

// translationResponse wraps a list of TranslationModel objects for a
// given languageId. Both ByLanguage and ById return this shape; the
// ById response is expected to carry at most one Translation entry.
type translationResponse struct {
	LanguageId   string             `json:"LanguageId"`
	Translations []translationModel `json:"Translations"`
}

// Languages returns the upstream list of languages.
func (c *Client) Languages(ctx context.Context) ([]Language, error) {
	body, err := c.do(ctx, "/api/Translation/Languages")
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var raw []languageModel
	if err := json.NewDecoder(body).Decode(&raw); err != nil {
		return nil, fmt.Errorf("translation_api: decode languages: %w", err)
	}
	out := make([]Language, 0, len(raw))
	for _, r := range raw {
		id := strings.TrimSpace(r.Id)
		if id == "" {
			continue
		}
		out = append(out, Language{
			ID:        id,
			Name:      strings.TrimSpace(r.Name),
			IsDefault: r.IsDefault,
		})
	}
	return out, nil
}

// ByLanguage returns every translation for the given language.
func (c *Client) ByLanguage(ctx context.Context, languageID string) ([]Translation, error) {
	languageID = strings.TrimSpace(languageID)
	if languageID == "" {
		return nil, errors.New("translation_api: empty languageID")
	}
	body, err := c.do(ctx, "/api/Translation/ByLanguage/"+url.PathEscape(languageID))
	if err != nil {
		return nil, err
	}
	defer body.Close()
	var tr translationResponse
	if err := json.NewDecoder(body).Decode(&tr); err != nil {
		return nil, fmt.Errorf("translation_api: decode by language: %w", err)
	}
	return mapTranslations(languageID, tr.Translations), nil
}

// ByID returns one translation by id. The bool return is false when the
// upstream response carries an empty Translations array.
func (c *Client) ByID(ctx context.Context, languageID string, translationID int64) (Translation, bool, error) {
	languageID = strings.TrimSpace(languageID)
	if languageID == "" {
		return Translation{}, false, errors.New("translation_api: empty languageID")
	}
	path := fmt.Sprintf("/api/Translation/ById/%s?id=%d", url.PathEscape(languageID), translationID)
	body, err := c.do(ctx, path)
	if err != nil {
		return Translation{}, false, err
	}
	defer body.Close()
	var tr translationResponse
	if err := json.NewDecoder(body).Decode(&tr); err != nil {
		return Translation{}, false, fmt.Errorf("translation_api: decode by id: %w", err)
	}
	if len(tr.Translations) == 0 {
		return Translation{}, false, nil
	}
	first := tr.Translations[0]
	return Translation{
		LanguageID:    languageID,
		TranslationID: first.Id,
		Text:          first.Text,
	}, true, nil
}

// do dispatches a GET and returns a reader that transparently
// decompresses gzip responses. The caller is responsible for closing
// the returned ReadCloser.
func (c *Client) do(ctx context.Context, path string) (io.ReadCloser, error) {
	if c.base == "" {
		return nil, errors.New("translation_api: BaseURL not configured")
	}
	full := c.base + path
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, full, nil)
	if err != nil {
		return nil, fmt.Errorf("translation_api: build request: %w", err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Accept-Encoding", "gzip")

	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("translation_api: transport %s: %w", path, err)
	}
	if resp.StatusCode >= 400 {
		retry, _ := strconv.Atoi(resp.Header.Get("Retry-After"))
		_ = resp.Body.Close()
		return nil, &HTTPError{Path: path, StatusCode: resp.StatusCode, RetryAfter: retry}
	}
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		gz, err := gzip.NewReader(resp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("translation_api: gzip %s: %w", path, err)
		}
		return &gzipBody{Reader: gz, raw: resp.Body}, nil
	}
	return resp.Body, nil
}

// gzipBody bundles a gzip.Reader with the underlying response body so
// Close releases both resources.
type gzipBody struct {
	*gzip.Reader
	raw io.ReadCloser
}

// Close releases the gzip stream and the network connection.
func (b *gzipBody) Close() error {
	_ = b.Reader.Close()
	return b.raw.Close()
}

func mapTranslations(languageID string, in []translationModel) []Translation {
	out := make([]Translation, 0, len(in))
	for _, t := range in {
		out = append(out, Translation{
			LanguageID:    languageID,
			TranslationID: t.Id,
			Text:          t.Text,
		})
	}
	return out
}

// Ensure the concrete Client satisfies the API interface.
var _ API = (*Client)(nil)
