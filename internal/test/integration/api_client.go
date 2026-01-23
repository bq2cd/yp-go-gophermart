package integration

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler/api"
	"resty.dev/v3"
)

type APIError struct {
	Status int
	Body   []byte
}

func (e *APIError) Error() string {
	return fmt.Sprintf("API error: status=%d, body=%s", e.Status, string(e.Body))
}

/////////////////////////////////////////////////////////////////////////////////

type APIRequest struct {
	*resty.Request
}

func (r *APIRequest) OperationRegisterUser(login, password string) (*APIResponse, error) {
	req := api.LoginPassword{
		Login:    login,
		Password: password,
	}

	resp, err := r.SetBody(req).Post("/api/user/register")

	return &APIResponse{resp}, err
}

func (r *APIRequest) OperationAuthenticateUser(login, password string) (*APIResponse, error) {
	req := api.LoginPassword{
		Login:    login,
		Password: password,
	}

	resp, err := r.SetBody(req).Post("/api/user/login")

	return &APIResponse{resp}, err
}

/////////////////////////////////////////////////////////////////////////////////

type APIResponse struct {
	*resty.Response
}

func DecodeAPIResponse[T any](resp *APIResponse) (T, error) {
	var out T

	err := json.Unmarshal(resp.Bytes(), &out)

	return out, err
}

/////////////////////////////////////////////////////////////////////////////////

type APIClient struct {
	client *resty.Client
}

func NewAPIClient(baseURL string) *APIClient {
	client := resty.New()

	client.SetBaseURL(baseURL)
	client.SetResponseBodyUnlimitedReads(true)
	client.SetDisableWarn(true)

	return &APIClient{client: client}
}

func (c *APIClient) R() *APIRequest {
	return &APIRequest{c.client.R()}
}

/////////////////////////////////////////////////////////////////////////////////

type SecureAPIRequest struct {
	*resty.Request
}

func (r *SecureAPIRequest) WithDebug() *SecureAPIRequest {
	r.SetDebug(true)

	return r
}

func (r *SecureAPIRequest) OperationUploadOrder(orderID string) (*APIResponse, error) {
	resp, err := r.SetBody(orderID).Post("/api/user/orders")

	return &APIResponse{resp}, err
}

func (r *SecureAPIRequest) OperationListOrders() (*APIResponse, error) {
	resp, err := r.Get("/api/user/orders")

	return &APIResponse{resp}, err
}

func (r *SecureAPIRequest) OperationGetBalance() (*APIResponse, error) {
	resp, err := r.Get("/api/user/balance")

	return &APIResponse{resp}, err
}

func (r *SecureAPIRequest) OperationListWithdrawals() (*APIResponse, error) {
	resp, err := r.Get("/api/user/withdrawals")

	return &APIResponse{resp}, err
}

func (r *SecureAPIRequest) OperationWithdraw(req api.WithdrawalRequest) (*APIResponse, error) {
	resp, err := r.SetBody(req).Post("/api/user/balance/withdraw")

	return &APIResponse{resp}, err
}

/////////////////////////////////////////////////////////////////////////////////

type SecureAPIClient struct {
	api          *APIClient
	tokenPerUser map[string]string
}

func NewSecureAPIClient(baseURL string) *SecureAPIClient {
	return &SecureAPIClient{
		api:          NewAPIClient(baseURL),
		tokenPerUser: map[string]string{},
	}
}

func (c *SecureAPIClient) R(login string) *SecureAPIRequest {
	req := c.api.R().
		SetAuthScheme("Bearer").
		SetAuthToken(c.tokenPerUser[login])

	return &SecureAPIRequest{req}
}

func (c *SecureAPIClient) EnsureUserExists(login, password string) error {
	var apiErr *APIError

	err := c.registerOrAuthenticate(login, password, c.api.R().OperationRegisterUser)
	if err == nil {
		return nil
	}

	if !errors.As(err, &apiErr) {
		return err
	}

	if apiErr.Status != http.StatusConflict {
		return err
	}

	return c.registerOrAuthenticate(login, password, c.api.R().OperationAuthenticateUser)
}

func (c *SecureAPIClient) registerOrAuthenticate(
	login, password string,
	operationFn func(string, string) (*APIResponse, error),
) error {
	resp, err := operationFn(login, password)
	if err != nil {
		return err
	}

	err = c.checkResponseStatus(resp)
	if err != nil {
		return err
	}

	c.saveAuthToken(login, resp)

	return nil
}

func (c *SecureAPIClient) checkResponseStatus(resp *APIResponse) error {
	if resp.IsSuccess() {
		return nil
	}

	return &APIError{
		Status: resp.StatusCode(),
		Body:   resp.Bytes(),
	}
}

func (c *SecureAPIClient) saveAuthToken(login string, resp *APIResponse) {
	header := resp.Header().Get(api.AuthorizationHeaderName)
	token := strings.TrimPrefix(header, api.AuthorizationHeaderValuePrefix)

	c.tokenPerUser[login] = token
}
