package sectors

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// benchBody builds a representative ~50KB JSON response body so the cache
// benchmarks reflect real report-sized payloads.
func benchBody(tb testing.TB) []byte {
	tb.Helper()
	rows := make([]map[string]any, 0, 600)
	for i := 0; i < 600; i++ {
		rows = append(rows, map[string]any{
			"symbol": "BBCA.JK", "name": "PT Bank Central Asia Tbk",
			"close": 9875.5, "volume": 1234567, "i": i,
		})
	}
	body, err := json.Marshal(map[string]any{"results": rows, "count": len(rows)})
	if err != nil {
		tb.Fatal(err)
	}
	return body
}

func benchEntry(body []byte) *cacheEntry {
	return &cacheEntry{
		URL:    "https://api.sectors.app/v2/company/report/BBCA/",
		Status: 200,
		Header: map[string][]string{"Content-Type": {"application/json"}},
		Body:   body, StoredAt: time.Now(),
	}
}

// BenchmarkCacheReadFresh covers the cache-hit read path, which runs on every
// cached invocation. The framed on-disk format (header line + raw body) keeps
// this off the base64 decode + full-body JSON validation that a serialized
// `[]byte` body would force.
func BenchmarkCacheReadFresh(b *testing.B) {
	file := filepath.Join(b.TempDir(), "entry.json")
	store(file, benchEntry(benchBody(b)))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, ok := readFresh(file, time.Hour); !ok {
			b.Fatal("unexpected miss")
		}
	}
}

func BenchmarkCacheStore(b *testing.B) {
	file := filepath.Join(b.TempDir(), "entry.json")
	e := benchEntry(benchBody(b))
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		store(file, e)
	}
}

// TestCache_RoundTripsRawBody guards the framed format: the stored body must
// survive byte-for-byte and the file must not base64-inflate it.
func TestCache_RoundTripsRawBody(t *testing.T) {
	body := benchBody(t)
	file := filepath.Join(t.TempDir(), "entry.json")
	store(file, benchEntry(body))

	got, ok := readFresh(file, time.Hour)
	if !ok {
		t.Fatal("readFresh reported a miss for a fresh entry")
	}
	if !bytes.Equal(got.Body, body) {
		t.Fatalf("body round-trip mismatch: got %d bytes, want %d", len(got.Body), len(body))
	}
	if got.Status != 200 || got.URL == "" {
		t.Errorf("metadata lost: status=%d url=%q", got.Status, got.URL)
	}
}

// TestCache_LegacyFileIsMiss ensures a pre-framing (single-line JSON) cache file
// is treated as a miss rather than mis-parsed, so old caches refetch cleanly.
func TestCache_LegacyFileIsMiss(t *testing.T) {
	legacy, _ := json.Marshal(map[string]any{
		"url": "u", "status": 200, "body": "anything", "stored_at": time.Now(),
	})
	file := filepath.Join(t.TempDir(), "entry.json")
	if err := os.WriteFile(file, legacy, 0o600); err != nil {
		t.Fatal(err)
	}
	if _, ok := readFresh(file, time.Hour); ok {
		t.Error("legacy single-line cache file should be treated as a miss")
	}
}
