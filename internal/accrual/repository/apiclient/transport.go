package apiclient

import (
	"net/http"

	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

// WithTransport returns an option to override transport configuration
// of a [Client].
func WithTransport(transport http.RoundTripper) option.Option[Client] {
	return func(c *Client) {
		c.httpClient.SetTransport(transport)
	}
}
