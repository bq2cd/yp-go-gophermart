package apiclient

import (
	"context"
	"fmt"
	"net/http"

	"resty.dev/v3"

	"github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
)

// Client wraps HTTP client and provides an interface to interact with an accrual system.
type Client struct {
	httpClient *resty.Client
}

// NewClient creates an instance of [Client].
func NewClient(baseURL string, options ...Option) *Client {
	client := &Client{
		httpClient: resty.New().SetBaseURL(baseURL),
	}

	client.setRetryConditions()
	client.applyRetryConfig(DefaultRetryConfig())

	for _, opt := range options {
		opt(client)
	}

	return client
}

// GetOrderStatus returns order information for a given order ID.
// It will return [domain.ErrOrderNotFound] if such order does not exist.
// It will return [domain.ErrRateLimitExceeded] if an accrual system server is overloaded.
// Other errors might be returned if network connection fails.
func (c *Client) GetOrderStatus(ctx context.Context, orderID domain.OrderID) (domain.Order, error) {
	var order domain.Order

	resp, err := c.sendGetOrderStatusRequest(ctx, orderID, &order)
	if err != nil {
		return order, err
	}

	err = c.processGetOrderStatusResponse(resp)
	if err != nil {
		return order, err
	}

	return order, nil
}

func (c *Client) sendGetOrderStatusRequest(
	ctx context.Context,
	orderID domain.OrderID,
	result *domain.Order,
) (*resty.Response, error) {
	req := c.httpClient.R().
		WithContext(ctx).
		SetPathParam("orderId", orderID.String()).
		SetResult(result)

	resp, err := req.Get("/api/orders/{orderId}")
	if err != nil {
		return resp, fmt.Errorf("cannot send HTTP request: %w", err)
	}

	return resp, nil
}

func (c *Client) processGetOrderStatusResponse(resp *resty.Response) error {
	switch resp.StatusCode() {
	case http.StatusOK:
		return nil
	case http.StatusNoContent:
		return domain.ErrOrderNotFound
	case http.StatusTooManyRequests:
		return domain.ErrRateLimitExceeded
	default:
		return fmt.Errorf("%w: %s", ErrUnexpectedHTTPStatus, resp.Status())
	}
}

func (c *Client) setRetryConditions() {
	c.httpClient.SetRetryDefaultConditions(false)

	c.httpClient.AddRetryConditions(
		func(resp *resty.Response, err error) bool {
			if err != nil {
				return false
			}

			switch resp.StatusCode() {
			case http.StatusTooManyRequests:
				return true
			default:
				return false
			}
		},
	)
}

func (c *Client) applyRetryConfig(config RetryConfig) {
	c.httpClient.
		SetRetryCount(config.Count).
		SetRetryWaitTime(config.WaitTime).
		SetRetryMaxWaitTime(config.MaxWaitTime)
}
