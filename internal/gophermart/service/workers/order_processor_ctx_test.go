package workers_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers/mocks"
)

type TestOrderProcessorContext struct {
	InputOrders     []domain.OrderID
	OrderRepoData   mocks.TestOrderRepoData
	OrderRepoErrors mocks.TestOrderRepoErrorMap
	OrderRepoDelays mocks.TestOrderRepoDelayMap
	AccrualData     mocks.TestAccrualData
	AccrualErrors   mocks.TestAccrualErrorMap
	AccrualDelays   mocks.TestAccrualDelayMap
}

func NewTestOrderProcessorContext() *TestOrderProcessorContext {
	return &TestOrderProcessorContext{
		InputOrders:     []domain.OrderID{},
		OrderRepoData:   mocks.TestOrderRepoData{},
		OrderRepoErrors: mocks.TestOrderRepoErrorMap{},
		OrderRepoDelays: mocks.TestOrderRepoDelayMap{},
		AccrualData:     mocks.TestAccrualData{},
		AccrualErrors:   mocks.TestAccrualErrorMap{},
		AccrualDelays:   mocks.TestAccrualDelayMap{},
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
	orderRepo *mocks.TestOrderRepository,
	accrualClient *mocks.TestAccrualClient,
) {
	orderRepo.Setup(pc.OrderRepoData, pc.OrderRepoErrors, pc.OrderRepoDelays)
	accrualClient.Setup(pc.AccrualData, pc.AccrualErrors, pc.AccrualDelays)
}
