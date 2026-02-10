package workers_test

import (
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

type TestOrderProcessorContext struct {
	OrderRepoData   fakes.TestOrderRepoData
	OrderRepoErrors fakes.TestOrderRepoErrorMap
	OrderRepoDelays fakes.TestOrderRepoDelayMap
	AccrualData     fakes.TestAccrualData
	AccrualErrors   fakes.TestAccrualErrorMap
	AccrualDelays   fakes.TestAccrualDelayMap
}

func NewTestOrderProcessorContext() *TestOrderProcessorContext {
	return &TestOrderProcessorContext{
		OrderRepoData:   fakes.TestOrderRepoData{},
		OrderRepoErrors: fakes.TestOrderRepoErrorMap{},
		OrderRepoDelays: fakes.TestOrderRepoDelayMap{},
		AccrualData:     fakes.TestAccrualData{},
		AccrualErrors:   fakes.TestAccrualErrorMap{},
		AccrualDelays:   fakes.TestAccrualDelayMap{},
	}
}

func (pc *TestOrderProcessorContext) SetupMocks(
	orderRepo *fakes.TestOrderRepository,
	accrualClient *fakes.TestAccrualClient,
) {
	orderRepo.Setup(pc.OrderRepoData, pc.OrderRepoErrors, pc.OrderRepoDelays)
	accrualClient.Setup(pc.AccrualData, pc.AccrualErrors, pc.AccrualDelays)
}
