package sectors

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func cacheClient(t *testing.T, ttl time.Duration) *http.Client {
	t.Helper()
	rt := NewCacheTransport(http.DefaultTransport, CacheConfig{Dir: t.TempDir(), TTL: ttl})
	return &http.Client{Transport: rt}
}

func TestCache_ServesSecondCallFromDisk(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		_, _ = w.Write([]byte(`{"v":1}`))
	}))
	defer srv.Close()

	c := cacheClient(t, time.Minute)

	r1, err := c.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	b1, _ := io.ReadAll(r1.Body)
	r1.Body.Close()
	if r1.Header.Get(CacheHeader) != "miss" {
		t.Errorf("first call should be a miss, got %q", r1.Header.Get(CacheHeader))
	}

	r2, err := c.Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	b2, _ := io.ReadAll(r2.Body)
	r2.Body.Close()
	if r2.Header.Get(CacheHeader) != "hit" {
		t.Errorf("second call should be a hit, got %q", r2.Header.Get(CacheHeader))
	}

	if string(b1) != string(b2) || string(b2) != `{"v":1}` {
		t.Errorf("bodies differ: %q vs %q", b1, b2)
	}
	if got := atomic.LoadInt32(&hits); got != 1 {
		t.Errorf("server hit %d times, want 1 (second served from cache)", got)
	}
}

func TestCache_ExpiredEntryRefetches(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	c := cacheClient(t, time.Nanosecond) // expires immediately

	_, _ = c.Get(srv.URL)
	time.Sleep(time.Millisecond)
	r, _ := c.Get(srv.URL)
	r.Body.Close()
	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Errorf("expired cache should refetch: server hits = %d, want 2", got)
	}
}

func TestCache_DoesNotStoreErrors(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&hits, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := cacheClient(t, time.Minute)
	r1, _ := c.Get(srv.URL)
	r1.Body.Close()
	r2, _ := c.Get(srv.URL)
	r2.Body.Close()
	if got := atomic.LoadInt32(&hits); got != 2 {
		t.Errorf("5xx must not be cached: hits = %d, want 2", got)
	}
}

func TestCache_DisabledWithoutDir(t *testing.T) {
	if _, ok := NewCacheTransport(http.DefaultTransport, CacheConfig{}).(*cacheTransport); ok {
		t.Error("empty Dir should yield base transport, not a cacheTransport")
	}
}

func TestClassifyTTL(t *testing.T) {
	cases := map[string]time.Duration{
		"/v2/subsectors/":          24 * time.Hour,
		"/v2/tags/":                24 * time.Hour,
		"/v2/daily/BBCA/":          time.Minute,
		"/v2/news/":                time.Minute,
		"/v2/company/report/BBCA/": 5 * time.Minute,
	}
	for path, want := range cases {
		if got := classifyTTL(path); got != want {
			t.Errorf("classifyTTL(%s) = %v, want %v", path, got, want)
		}
	}
}

func TestDryRunTransport_NeverSends(t *testing.T) {
	var hits int32
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		atomic.AddInt32(&hits, 1)
	}))
	defer srv.Close()

	c := &http.Client{Transport: NewDryRunTransport()}
	req, _ := http.NewRequest("GET", srv.URL+"/v2/x/", nil)
	req.Header.Set("Authorization", "secret-key")
	resp, err := c.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	body, _ := io.ReadAll(resp.Body)
	resp.Body.Close()

	if atomic.LoadInt32(&hits) != 0 {
		t.Error("dry-run must not hit the network")
	}
	s := string(body)
	if !contains(s, `"dry_run":true`) || !contains(s, "redacted") || contains(s, "secret-key") {
		t.Errorf("dry-run body wrong (key must be redacted): %s", s)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
