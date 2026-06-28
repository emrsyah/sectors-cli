package sectors

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// CacheHeader is set on responses served (or filled) by the cache transport so
// outer layers (e.g. verbose logging) can tell hits from misses.
const CacheHeader = "X-Sectors-Cache"

// CacheConfig configures the on-disk response cache.
type CacheConfig struct {
	Dir string        // directory for cache files
	TTL time.Duration // 0 = classify per endpoint; >0 = uniform TTL
}

type cacheTransport struct {
	base http.RoundTripper
	cfg  CacheConfig
}

// NewCacheTransport wraps base with an on-disk GET cache. A response is served
// from disk while fresh; otherwise the request goes through and a 2xx result is
// stored. Non-GET requests pass straight through.
func NewCacheTransport(base http.RoundTripper, cfg CacheConfig) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if cfg.Dir == "" {
		return base // nowhere to cache → no-op
	}
	return &cacheTransport{base: base, cfg: cfg}
}

type cacheEntry struct {
	URL      string
	Status   int
	Header   map[string][]string
	Body     []byte
	StoredAt time.Time
}

// cacheHead is the on-disk metadata, serialized as a single JSON line. The
// response body is NOT a field here: it is appended raw after the header line
// (see store/readFresh). Keeping the body out of the JSON envelope avoids the
// ~33% size blowup and the base64 encode/decode + full body re-validation that
// a `[]byte`/`json.RawMessage` field would incur on every cache hit.
type cacheHead struct {
	URL      string              `json:"url"`
	Status   int                 `json:"status"`
	Header   map[string][]string `json:"header"`
	StoredAt time.Time           `json:"stored_at"`
}

func (t *cacheTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Method != http.MethodGet {
		return t.base.RoundTrip(req)
	}

	ttl := t.cfg.TTL
	if ttl == 0 {
		ttl = classifyTTL(req.URL.Path)
	}
	file := filepath.Join(t.cfg.Dir, cacheKey(req)+".json")

	if e, ok := readFresh(file, ttl); ok {
		resp := e.toResponse(req)
		resp.Header.Set(CacheHeader, "hit")
		return resp, nil
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}
	if resp.StatusCode >= 200 && resp.StatusCode < 300 {
		body, rerr := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if rerr == nil {
			store(file, &cacheEntry{
				URL:      req.URL.String(),
				Status:   resp.StatusCode,
				Header:   map[string][]string{"Content-Type": resp.Header["Content-Type"]},
				Body:     body,
				StoredAt: time.Now(),
			})
			resp.Body = io.NopCloser(bytes.NewReader(body))
		}
	}
	resp.Header.Set(CacheHeader, "miss")
	return resp, nil
}

// cacheKey hashes method + URL + auth so different keys / queries never collide.
func cacheKey(req *http.Request) string {
	h := sha256.New()
	io.WriteString(h, req.Method+"\n"+req.URL.String()+"\n"+req.Header.Get("Authorization"))
	return hex.EncodeToString(h.Sum(nil))
}

func readFresh(file string, ttl time.Duration) (*cacheEntry, bool) {
	raw, err := os.ReadFile(file)
	if err != nil {
		return nil, false
	}
	// Frame is "<json header>\n<raw body>". A missing newline means a
	// legacy/corrupt file; treat it as a miss so it gets refetched and rewritten.
	nl := bytes.IndexByte(raw, '\n')
	if nl < 0 {
		return nil, false
	}
	var h cacheHead
	if err := json.Unmarshal(raw[:nl], &h); err != nil {
		return nil, false
	}
	if time.Since(h.StoredAt) > ttl {
		return nil, false
	}
	return &cacheEntry{URL: h.URL, Status: h.Status, Header: h.Header, Body: raw[nl+1:], StoredAt: h.StoredAt}, true
}

func store(file string, e *cacheEntry) {
	if err := os.MkdirAll(filepath.Dir(file), 0o700); err != nil {
		return
	}
	// The HTTP header marshals to a single JSON line (header values can't contain
	// newlines), so the first '\n' cleanly separates metadata from the raw body.
	head, err := json.Marshal(cacheHead{URL: e.URL, Status: e.Status, Header: e.Header, StoredAt: e.StoredAt})
	if err != nil {
		return
	}
	raw := make([]byte, 0, len(head)+1+len(e.Body))
	raw = append(raw, head...)
	raw = append(raw, '\n')
	raw = append(raw, e.Body...)
	// Atomic write: temp then rename, so a reader never sees a partial file.
	tmp := file + ".tmp"
	if os.WriteFile(tmp, raw, 0o600) == nil {
		_ = os.Rename(tmp, file)
	}
}

func (e *cacheEntry) toResponse(req *http.Request) *http.Response {
	header := http.Header{}
	for k, v := range e.Header {
		header[k] = v
	}
	return &http.Response{
		StatusCode:    e.Status,
		Status:        http.StatusText(e.Status),
		Header:        header,
		Body:          io.NopCloser(bytes.NewReader(e.Body)),
		ContentLength: int64(len(e.Body)),
		Request:       req,
	}
}

// classifyTTL picks a cache lifetime by endpoint volatility. Reference data
// (slug/tag lists) is stable for a day; intraday data is cached only briefly;
// everything else (reports, financials) gets a few minutes.
func classifyTTL(path string) time.Duration {
	reference := []string{
		"/industries/", "/subindustries/", "/subsectors/", "/tags/",
		"/sectors/", "/brokers/", "/mining/commodities/",
		"list_companies_with_segments",
	}
	volatile := []string{
		"/daily/", "/news/", "/filings/", "/suspensions/", "/most-traded/",
		"/top-changes/", "/broker-activity/", "/broker-summary/",
		"/foreign-flow/", "/idx-total/", "/index-daily/",
	}
	for _, p := range reference {
		if strings.Contains(path, p) {
			return 24 * time.Hour
		}
	}
	for _, p := range volatile {
		if strings.Contains(path, p) {
			return time.Minute
		}
	}
	return 5 * time.Minute
}
