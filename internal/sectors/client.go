package sectors

import (
	"context"
	"net/http"
)

// DefaultBaseURL is the production Sectors Financial API server.
const DefaultBaseURL = "https://api.sectors.app"

// New constructs a typed Sectors API client with API-key authentication.
//
// apiKey is sent verbatim in the `Authorization` header on every request, per
// the API's ApiKeyAuth security scheme. If baseURL is empty, DefaultBaseURL is
// used. A nil httpClient falls back to the generated client's default.
func New(baseURL, apiKey string, httpClient *http.Client) (*ClientWithResponses, error) {
	if baseURL == "" {
		baseURL = DefaultBaseURL
	}

	opts := []ClientOption{
		WithRequestEditorFn(func(_ context.Context, req *http.Request) error {
			req.Header.Set("Authorization", apiKey)
			return nil
		}),
	}
	if httpClient != nil {
		opts = append(opts, WithHTTPClient(httpClient))
	}

	return NewClientWithResponses(baseURL, opts...)
}
