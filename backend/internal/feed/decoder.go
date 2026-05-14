package feed

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
)

// gzipMagic is the standard RFC 1952 GZIP header magic.
var gzipMagic = []byte{0x1f, 0x8b}

// DecodeBody returns the JSON bytes carried in a FeedConstruct delivery.
// FeedConstruct compresses both RMQ deliveries and WebAPI responses with
// GZIP; uncompressed bodies (e.g. test fixtures) pass through unchanged.
func DecodeBody(body []byte) ([]byte, error) {
	if !bytes.HasPrefix(body, gzipMagic) {
		return body, nil
	}
	r, err := gzip.NewReader(bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("feed: gzip reader: %w", err)
	}
	defer r.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("feed: gzip read: %w", err)
	}
	return out, nil
}
