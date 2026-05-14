package webapi

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

// decompressResponse reads resp.Body. Bodies arriving with
// Content-Encoding: gzip OR a leading gzip magic header are inflated
// transparently; everything else passes through.
func decompressResponse(resp *http.Response) ([]byte, error) {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("webapi: read body: %w", err)
	}
	if isGzipResponse(resp, body) {
		r, err := gzip.NewReader(bytes.NewReader(body))
		if err != nil {
			return nil, fmt.Errorf("webapi: gzip reader: %w", err)
		}
		defer r.Close()
		out, err := io.ReadAll(r)
		if err != nil {
			return nil, fmt.Errorf("webapi: gzip read: %w", err)
		}
		return out, nil
	}
	return body, nil
}

func isGzipResponse(resp *http.Response, body []byte) bool {
	if strings.EqualFold(resp.Header.Get("Content-Encoding"), "gzip") {
		return true
	}
	return len(body) >= 2 && body[0] == 0x1f && body[1] == 0x8b
}
