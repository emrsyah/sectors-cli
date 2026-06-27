package sectors

import (
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"time"
)

// retryTransport is an http.RoundTripper that transparently retries transient
// failures (network errors, 429, 502, 503, 504) on idempotent requests, with
// exponential backoff + jitter, honoring the server's Retry-After header.
type retryTransport struct {
	base       http.RoundTripper
	maxRetries int
	baseWait   time.Duration
	maxWait    time.Duration
}

// NewRetryTransport wraps base with retry behavior. maxRetries <= 0 disables
// retries and returns base unchanged. A nil base uses http.DefaultTransport.
func NewRetryTransport(base http.RoundTripper, maxRetries int, maxWait time.Duration) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	if maxRetries <= 0 {
		return base
	}
	if maxWait <= 0 {
		maxWait = 10 * time.Second
	}
	return &retryTransport{
		base:       base,
		maxRetries: maxRetries,
		baseWait:   200 * time.Millisecond,
		maxWait:    maxWait,
	}
}

func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error

	for attempt := 0; ; attempt++ {
		resp, err = t.base.RoundTrip(req)

		if attempt >= t.maxRetries || !shouldRetry(req.Method, resp, err) {
			return resp, err
		}

		wait := t.backoff(attempt, resp)
		// Discard the failed response body so the connection can be reused.
		if resp != nil {
			_, _ = io.Copy(io.Discard, resp.Body)
			_ = resp.Body.Close()
		}

		select {
		case <-time.After(wait):
		case <-req.Context().Done():
			return nil, req.Context().Err()
		}
	}
}

// shouldRetry reports whether a request should be retried. Only idempotent
// methods are retried (all Sectors endpoints are GET).
func shouldRetry(method string, resp *http.Response, err error) bool {
	if method != http.MethodGet && method != http.MethodHead {
		return false
	}
	if err != nil {
		return true // transport/network error
	}
	switch resp.StatusCode {
	case http.StatusTooManyRequests, // 429
		http.StatusInternalServerError, // 500 (this API emits transient 500s)
		http.StatusBadGateway,          // 502
		http.StatusServiceUnavailable,  // 503
		http.StatusGatewayTimeout:      // 504
		return true
	}
	return false
}

// backoff returns how long to wait before the next attempt. It honors a
// Retry-After header when present, otherwise uses exponential backoff with
// full jitter, capped at maxWait.
func (t *retryTransport) backoff(attempt int, resp *http.Response) time.Duration {
	if resp != nil {
		if d, ok := retryAfter(resp.Header.Get("Retry-After")); ok {
			return clamp(d, t.maxWait)
		}
	}
	// Exponential: baseWait * 2^attempt, then full jitter in [0, exp].
	exp := t.baseWait << attempt
	if exp <= 0 || exp > t.maxWait {
		exp = t.maxWait
	}
	return time.Duration(rand.Int63n(int64(exp) + 1))
}

// retryAfter parses a Retry-After header (delta-seconds or HTTP-date).
func retryAfter(v string) (time.Duration, bool) {
	if v == "" {
		return 0, false
	}
	if secs, err := strconv.Atoi(v); err == nil {
		return time.Duration(secs) * time.Second, true
	}
	if when, err := http.ParseTime(v); err == nil {
		if d := time.Until(when); d > 0 {
			return d, true
		}
		return 0, true
	}
	return 0, false
}

func clamp(d, max time.Duration) time.Duration {
	if d > max {
		return max
	}
	if d < 0 {
		return 0
	}
	return d
}
