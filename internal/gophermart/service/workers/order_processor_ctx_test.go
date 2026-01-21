package workers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

type TestOrderProcessorContext struct {
	InputOrders     []domain.OrderID
	OrderRepoData   fakes.TestOrderRepoData
	OrderRepoErrors fakes.TestOrderRepoErrorMap
	OrderRepoDelays fakes.TestOrderRepoDelayMap
	AccrualData     fakes.TestAccrualData
	AccrualErrors   fakes.TestAccrualErrorMap
	AccrualDelays   fakes.TestAccrualDelayMap
}

func NewTestOrderProcessorContext() *TestOrderProcessorContext {
	return &TestOrderProcessorContext{
		InputOrders:     []domain.OrderID{},
		OrderRepoData:   fakes.TestOrderRepoData{},
		OrderRepoErrors: fakes.TestOrderRepoErrorMap{},
		OrderRepoDelays: fakes.TestOrderRepoDelayMap{},
		AccrualData:     fakes.TestAccrualData{},
		AccrualErrors:   fakes.TestAccrualErrorMap{},
		AccrualDelays:   fakes.TestAccrualDelayMap{},
	}
}

func (pc *TestOrderProcessorContext) EnqueueOrders(orderProcessor *workers.OrderProcessor) {
	GinkgoHelper()

	Expect(pc.InputOrders).NotTo(BeEmpty())

	userID := domain.UserID("dummy-user")
	for _, orderID := range pc.InputOrders {
		order, ok := pc.OrderRepoData.Orders[orderID]
		if ok {
			userID = order.UserID
		}

		accepted := orderProcessor.EnqueueOrder(userID, orderID)
		Expect(accepted).To(BeTrue(), "order must be enqueued")
	}
}

func (pc *TestOrderProcessorContext) SetupMocks(
	orderRepo *fakes.TestOrderRepository,
	accrualClient *fakes.TestAccrualClient,
) {
	orderRepo.Setup(pc.OrderRepoData, pc.OrderRepoErrors, pc.OrderRepoDelays)
	accrualClient.Setup(pc.AccrualData, pc.AccrualErrors, pc.AccrualDelays)
}
