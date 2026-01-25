package apiclient

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"

	"resty.dev/v3"

	"github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

// Client wraps HTTP client and provides an interface to interact with an accrual system.
type Client struct {
	httpClient *resty.Client
}

// NewClient creates an instance of [Client].
func NewClient(baseURL string, options ...option.Option[Client]) *Client {
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

	resp, err := c.sendGetOrderStatusRequest(ctx, orderID)
	if err != nil {
		return order, err
	}

	orderResp, err := c.processGetOrderStatusResponse(resp)
	if err != nil {
		return order, err
	}

	return orderResp.ToOrder()
}

func (c *Client) sendGetOrderStatusRequest(
	ctx context.Context,
	orderID domain.OrderID,
) (*resty.Response, error) {
	var result OrderResponse

	req := c.httpClient.R().
		WithContext(ctx).
		SetPathParam("orderId", orderID.String()).
		SetResult(result)

	resp, err := req.Get("/api/orders/{orderId}")

	slog.DebugContext(ctx, "accrual system response",
		slog.Uint64("order_id", uint64(orderID)),
		slog.Int("status_code", resp.StatusCode()),
		slog.Any("error", err),
	)

	if err != nil {
		return resp, fmt.Errorf("cannot send HTTP request: %w", err)
	}

	return resp, nil
}

func (c *Client) processGetOrderStatusResponse(resp *resty.Response) (OrderResponse, error) {
	var orderResp OrderResponse

	switch resp.StatusCode() {
	case http.StatusOK:
		orderRespPtr, ok := resp.Result().(*OrderResponse)
		if !ok {
			return orderResp, ErrUnexpectedAPIResultType
		}

		return *orderRespPtr, nil
	case http.StatusNoContent:
		return orderResp, domain.ErrOrderNotFound
	case http.StatusTooManyRequests:
		return orderResp, domain.ErrRateLimitExceeded
	default:
		return orderResp, fmt.Errorf("%w: %s", ErrUnexpectedHTTPStatus, resp.Status())
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
