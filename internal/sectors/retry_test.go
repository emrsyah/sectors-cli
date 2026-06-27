package sectors

import (
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// newClient builds an http.Client whose transport retries quickly (tiny waits).
func newRetryClient(maxRetries int) *http.Client {
	rt := NewRetryTransport(http.DefaultTransport, maxRetries, 5*time.Millisecond).(*retryTransport)
	rt.baseWait = time.Millisecond
	return &http.Client{Transport: rt}
}

func TestRetry_RecoversAfterTransientFailures(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if atomic.AddInt32(&calls, 1) <= 2 {
			w.WriteHeader(http.StatusServiceUnavailable) // 503 twice
			return
		}
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`ok`))
	}))
	defer srv.Close()

	resp, err := newRetryClient(3).Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 {
		t.Errorf("calls = %d, want 3 (2 failures + 1 success)", got)
	}
	body, _ := io.ReadAll(resp.Body)
	if string(body) != "ok" {
		t.Errorf("body = %q", body)
	}
}

func TestRetry_ExhaustsAndReturnsLastResponse(t *testing.T) {
	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusTooManyRequests)
	}))
	defer srv.Close()

	resp, err := newRetryClient(2).Get(srv.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("status = %d, want 429", resp.StatusCode)
	}
	if got := atomic.LoadInt32(&calls); got != 3 { // 1 initial + 2 retries
		t.Errorf("calls = %d, want 3", got)
	}
}

func TestRetry_DisabledWhenZero(t *testing.T) {
	// NewRetryTransport with 0 returns the base transport unchanged.
	if _, ok := NewRetryTransport(http.DefaultTransport, 0, time.Second).(*retryTransport); ok {
		t.Error("maxRetries=0 should return base transport, not a retryTransport")
	}
}

func TestShouldRetry(t *testing.T) {
	cases := []struct {
		method string
		status int
		err    error
		want   bool
	}{
		{"GET", 200, nil, false},
		{"GET", 429, nil, true},
		{"GET", 500, nil, true},
		{"GET", 503, nil, true},
		{"GET", 404, nil, false},
		{"GET", 400, nil, false},
		{"GET", 0, io.EOF, true},  // network error
		{"POST", 503, nil, false}, // non-idempotent never retried
	}
	for _, c := range cases {
		var resp *http.Response
		if c.status != 0 {
			resp = &http.Response{StatusCode: c.status}
		}
		if got := shouldRetry(c.method, resp, c.err); got != c.want {
			t.Errorf("shouldRetry(%s, %d, %v) = %v, want %v", c.method, c.status, c.err, got, c.want)
		}
	}
}

func TestRetryAfter(t *testing.T) {
	if d, ok := retryAfter("2"); !ok || d != 2*time.Second {
		t.Errorf("seconds: got %v, %v", d, ok)
	}
	if _, ok := retryAfter(""); ok {
		t.Error("empty should not parse")
	}
	if _, ok := retryAfter("garbage"); ok {
		t.Error("garbage should not parse")
	}
}
