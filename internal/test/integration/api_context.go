package integration

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/app"
)

type APITestContext struct {
	APIClient       *APIClient
	SecureAPIClient *SecureAPIClient
	accrualServer   *TestAccrualServer
	orderProcessor  *app.OrderProcessor
}

func (c *APITestContext) SetupAccrualServer(data TestAccrualData) {
	c.accrualServer.SetupData(data)
}

func (c *APITestContext) StartOrderProcessing(ctx context.Context) {
	go c.orderProcessor.Run(ctx) //nolint:errcheck // Error is never returned.
}

func (c *APITestContext) HasOrderProcessingFinished() bool {
	return c.orderProcessor.HasFinished()
}
