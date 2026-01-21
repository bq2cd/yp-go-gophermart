package integration

import (
	"context"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

type APITestContext struct {
	APIClient       *APIClient
	SecureAPIClient *SecureAPIClient
	accrualClient   *fakes.TestAccrualClient
	orderProcessor  *workers.OrderProcessor
}

func (c *APITestContext) SetupAccrualClient(data fakes.TestAccrualData) {
	c.accrualClient.Setup(data, fakes.TestAccrualErrorMap{}, fakes.TestAccrualDelayMap{})
}

func (c *APITestContext) StartOrderProcessing(ctx context.Context) {
	go c.orderProcessor.Run(ctx)
}

func (c *APITestContext) HasOrderProcessingFinished() bool {
	return c.orderProcessor.HasFinished()
}
