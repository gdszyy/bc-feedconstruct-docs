package feed_test

import (
	"bytes"
	"compress/gzip"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/gdszyy/bc-feedconstruct-docs/backend/internal/feed"
)

// Given a GZIP-compressed JSON body
// When DecodeBody is called
// Then it returns the uncompressed JSON bytes
func TestGiven_GzipBody_When_Decode_Then_PlainJSONReturned(t *testing.T) {
	plain := []byte(`{"objectType":13,"matchId":42}`)
	var buf bytes.Buffer
	w := gzip.NewWriter(&buf)
	_, err := w.Write(plain)
	require.NoError(t, w.Close())
	require.NoError(t, err)

	out, err := feed.DecodeBody(buf.Bytes())
	require.NoError(t, err)
	require.Equal(t, plain, out)
}

// Given an uncompressed JSON body (test fixtures)
// When DecodeBody is called
// Then it passes through unchanged
func TestGiven_PlainBody_When_Decode_Then_Passthrough(t *testing.T) {
	plain := []byte(`{"objectType":4,"matchId":7}`)
	out, err := feed.DecodeBody(plain)
	require.NoError(t, err)
	require.Equal(t, plain, out)
}

// Given a body that starts with the GZIP magic but is otherwise garbage
// When DecodeBody is called
// Then it surfaces a gzip error
func TestGiven_TruncatedGzip_When_Decode_Then_ReturnsError(t *testing.T) {
	bad := []byte{0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00}
	_, err := feed.DecodeBody(bad)
	require.Error(t, err)
}
