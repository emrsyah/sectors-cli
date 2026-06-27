package sectors

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// logTransport writes a one-line trace per request to w (stderr), including the
// status, wall-clock duration, and whether the cache served it.
type logTransport struct {
	base http.RoundTripper
	w    io.Writer
}

// NewLogTransport wraps base to emit request traces to w. Use for --verbose.
func NewLogTransport(base http.RoundTripper, w io.Writer) http.RoundTripper {
	if base == nil {
		base = http.DefaultTransport
	}
	return &logTransport{base: base, w: w}
}

func (t *logTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	start := time.Now()
	resp, err := t.base.RoundTrip(req)
	dur := time.Since(start).Round(time.Millisecond)

	status := "ERR"
	cache := ""
	if resp != nil {
		status = fmt.Sprintf("%d", resp.StatusCode)
		if resp.Header.Get(CacheHeader) == "hit" {
			cache = " (cache hit)"
		}
	}
	fmt.Fprintf(t.w, "%s %s -> %s %s%s\n", req.Method, req.URL, status, dur, cache)
	return resp, err
}

// dryRunTransport never performs a request; it returns a 200 whose body
// describes what would have been sent (with the API key redacted). Use for
// --dry-run.
type dryRunTransport struct{}

// NewDryRunTransport returns a transport that echoes requests instead of sending.
func NewDryRunTransport() http.RoundTripper { return &dryRunTransport{} }

func (t *dryRunTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	headers := map[string]string{}
	for k := range req.Header {
		v := req.Header.Get(k)
		if http.CanonicalHeaderKey(k) == "Authorization" {
			v = "***redacted***"
		}
		headers[k] = v
	}
	desc, _ := json.Marshal(map[string]any{
		"dry_run": true,
		"method":  req.Method,
		"url":     req.URL.String(),
		"headers": headers,
	})
	return &http.Response{
		StatusCode:    http.StatusOK,
		Status:        "200 OK",
		Header:        http.Header{"Content-Type": {"application/json"}},
		Body:          io.NopCloser(bytes.NewReader(desc)),
		ContentLength: int64(len(desc)),
		Request:       req,
	}, nil
}
