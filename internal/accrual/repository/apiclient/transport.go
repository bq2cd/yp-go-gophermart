package apiclient

import "net/http"

// WithTransport returns an option to override transport configuration
// of a [Client].
func WithTransport(transport http.RoundTripper) Option {
	return func(c *Client) {
		c.httpClient.SetTransport(transport)
	}
}
