package feed

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// Replayer drives the same ingest pipeline from on-disk files. Each file
// becomes one delivery; filenames sort lexically so timestamp-prefixed
// names (e.g. 2026-05-14T12-00-00Z_odds_change_42.json) replay in order.
//
// Supported extensions: .json (raw) and .json.gz (GZIP). Both go through
// the same Processor.DecodeBody path used by the live consumer.
type Replayer struct {
	Dir       string
	Processor *Processor
	// Pace pauses between deliveries; zero means as-fast-as-possible.
	Pace time.Duration
	// Source is the prefix used for raw_messages.source. Default "replay".
	Source string
}

// Run iterates the directory in name order, calling Processor.Process for
// each file. Stops on the first non-recoverable error or when ctx is done.
func (r *Replayer) Run(ctx context.Context) (int, error) {
	if r.Processor == nil {
		return 0, fmt.Errorf("feed: replayer needs Processor")
	}
	if r.Dir == "" {
		return 0, fmt.Errorf("feed: replayer needs Dir")
	}
	src := r.Source
	if src == "" {
		src = "replay"
	}

	entries, err := os.ReadDir(r.Dir)
	if err != nil {
		return 0, fmt.Errorf("feed: replayer read dir: %w", err)
	}
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		n := e.Name()
		if strings.HasSuffix(n, ".json") || strings.HasSuffix(n, ".json.gz") {
			names = append(names, n)
		}
	}
	sort.Strings(names)

	count := 0
	for _, name := range names {
		select {
		case <-ctx.Done():
			return count, ctx.Err()
		default:
		}
		full := filepath.Join(r.Dir, name)
		body, err := readFile(full)
		if err != nil {
			return count, fmt.Errorf("feed: replayer read %s: %w", name, err)
		}
		meta := DeliveryMeta{
			Source:     fmt.Sprintf("%s.%s", src, name),
			Queue:      inferQueue(name),
			RoutingKey: "",
		}
		if _, err := r.Processor.Process(ctx, body, meta); err != nil {
			return count, fmt.Errorf("feed: replayer process %s: %w", name, err)
		}
		count++
		if r.Pace > 0 {
			t := time.NewTimer(r.Pace)
			select {
			case <-ctx.Done():
				t.Stop()
				return count, ctx.Err()
			case <-t.C:
			}
		}
	}
	return count, nil
}

func readFile(path string) ([]byte, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	return io.ReadAll(f)
}

// inferQueue extracts a "live" / "prematch" hint from the filename when
// present so the audit row keeps the original queue context.
func inferQueue(name string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "_live"):
		return "live"
	case strings.Contains(lower, "_prematch"):
		return "prematch"
	}
	return ""
}
