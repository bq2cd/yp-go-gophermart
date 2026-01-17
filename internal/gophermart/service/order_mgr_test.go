package service_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/mocks"
)

// Ensure [service.OrderManager] implements [handler.OrderService].
var _ handler.OrderService = (*service.OrderManager)(nil)

var _ = Describe("OrderManager", func() {
	var (
		orderRepo      *mocks.MockOrderRepository
		orderProcessor *mocks.MockOrderProcessor
		orderMgr       *service.OrderManager
		userID         domain.UserID
		err            error
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())

		orderRepo = mocks.NewMockOrderRepository(ctrl)
		orderProcessor = mocks.NewMockOrderProcessor(ctrl)
		orderMgr = service.NewOrderManager(orderRepo, orderProcessor)
		userID = exampleUserID
	})

	Describe("creating an order", func() {
		var (
			orderID domain.OrderID
		)

		BeforeEach(func() {
			orderID = domain.OrderID(123456789)
		})

		JustBeforeEach(func() {
			err = orderMgr.CreateOrder(GinkgoT().Context(), userID, orderID)
		})

		When("order is brand new", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					CreateOrder(mockCtx(), userID, orderID).
					Return(true, userID, nil)

				orderProcessor.EXPECT().
					EnqueueOrder(userID, orderID).
					Return(true)
			})

			It("should succeed", func() {
				Expect(err).To(Succeed())
			})
		})

		When("order has been uploaded by the same user", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					CreateOrder(mockCtx(), userID, orderID).
					Return(false, userID, nil)
			})

			It("should return ErrOrderIDAlreadyExists with owner matching current user", func() {
				Expect(err).To(MatchError(func(errCheck error) bool {
					var errOrderExists *domain.OrderIDAlreadyExistsError
					if !errors.As(errCheck, &errOrderExists) {
						return false
					}

					return errOrderExists.CreatedBy == userID
				}, fmt.Sprintf("owner must match %v", userID)))
			})
		})

		When("order has been uploaded by another user", func() {
			var (
				createdBy domain.UserID
			)

			BeforeEach(func() {
				createdBy = domain.UserID("another-user")

				orderRepo.EXPECT().
					CreateOrder(mockCtx(), userID, orderID).
					Return(false, createdBy, nil)
			})

			It("should return ErrOrderIDAlreadyExists with owner matching another user", func() {
				Expect(err).To(MatchError(func(errCheck error) bool {
					var errOrderExists *domain.OrderIDAlreadyExistsError
					if !errors.As(errCheck, &errOrderExists) {
						return false
					}

					return errOrderExists.CreatedBy == createdBy
				}, fmt.Sprintf("owner must match %v", createdBy)))
			})
		})

		When("order repository fails on order creation", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					CreateOrder(mockCtx(), userID, orderID).
					Return(false, userID, errors.New("order is not worthy"))
			})
			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("order is not worthy")))
			})
		})

		When("order processor is shutting down", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					CreateOrder(mockCtx(), userID, orderID).
					Return(true, userID, nil)

				orderProcessor.EXPECT().
					EnqueueOrder(userID, orderID).
					Return(false)
			})
			It("should return an error", func() {
				Expect(err).To(MatchError(service.ErrOrderProcessorShuttingDown))
			})
		})
	})

	Describe("listing user's orders", func() {
		var (
			orders []domain.Order
		)

		JustBeforeEach(func() {
			orders, err = orderMgr.GetOrders(GinkgoT().Context(), userID)
		})

		When("user has no orders", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					GetOrders(mockCtx(), userID).
					Return(nil, nil)
			})

			It("should return empty array without errors", func() {
				Expect(err).To(Succeed())
				Expect(orders).To(BeEmpty())
			})
		})

		When("user has some orders", func() {
			var expectedOrders []domain.Order

			BeforeEach(func() {
				mockOrders := getExampleOrdersUnsorted()

				expectedOrders = slices.Clone(mockOrders)
				domain.SortByTimestampFromNewestToOldest(expectedOrders)
				Expect(expectedOrders).NotTo(Equal(mockOrders))

				orderRepo.EXPECT().
					GetOrders(mockCtx(), userID).
					Return(mockOrders, nil)
			})

			It("should return sorted array from the newest to the oldest", func() {
				Expect(err).To(Succeed())
				Expect(orders).To(Equal(expectedOrders))
			})
		})

		When("order repository fails on order retrieval", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					GetOrders(mockCtx(), userID).
					Return(nil, errors.New("no orders for you, sir"))
			})

			It("should return empty array and an error", func() {
				Expect(err).To(MatchError(ContainSubstring("no orders for you, sir")))
				Expect(orders).To(BeEmpty())
			})
		})
	})

	Describe("getting orders' accrual points", func() {
		var (
			orderIDs                         []domain.OrderID
			actualAccruals, expectedAccruals map[domain.OrderID]float64
		)

		BeforeEach(func() {
			orderIDs = service.GetOrderIDs(getExampleOrdersUnsorted())
		})

		JustBeforeEach(func() {
			actualAccruals, err = orderMgr.GetAccruals(GinkgoT().Context(), userID, orderIDs)
		})

		When("no order has accrual points", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					GetOrderAccruals(mockCtx(), userID, orderIDs).
					Return(nil, nil)
			})

			It("should succeed", func() {
				Expect(err).To(Succeed())
				Expect(actualAccruals).To(BeEmpty())
			})
		})

		When("some orders have accrual points", func() {
			BeforeEach(func() {
				expectedAccruals = getExampleAccruals(getExampleOrdersUnsorted())

				orderRepo.EXPECT().
					GetOrderAccruals(mockCtx(), userID, orderIDs).
					Return(expectedAccruals, nil)
			})

			It("should return only orders with accrual points", func() {
				Expect(err).To(Succeed())
				Expect(actualAccruals).To(Equal(expectedAccruals))
			})
		})

		When("order repository fails", func() {
			BeforeEach(func() {
				orderRepo.EXPECT().
					GetOrderAccruals(mockCtx(), userID, orderIDs).
					Return(nil, errors.New("accruals are not ready"))
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("accruals are not ready")))
				Expect(actualAccruals).To(BeEmpty())
			})
		})
	})

})

func getExampleOrdersUnsorted() []domain.Order {
	return []domain.Order{
		{
			ID:        123456789,
			Status:    domain.OrderStatusProcessed,
			CreatedAt: time.Date(2025, 11, 5, 10, 11, 12, 0, time.UTC),
		},
		{
			ID:        456789123,
			Status:    domain.OrderStatusInvalid,
			CreatedAt: time.Date(2025, 11, 7, 9, 10, 11, 0, time.UTC),
		},
		{
			ID:        789123456,
			Status:    domain.OrderStatusProcessing,
			CreatedAt: time.Date(2025, 11, 15, 8, 9, 10, 0, time.UTC),
		},
		{
			ID:        234567891,
			Status:    domain.OrderStatusNew,
			CreatedAt: time.Date(2025, 11, 17, 7, 8, 9, 0, time.UTC),
		},
		{
			ID:        345678912,
			Status:    domain.OrderStatusProcessed,
			CreatedAt: time.Date(2025, 11, 3, 12, 13, 14, 0, time.UTC),
		},
		{
			ID:        678912345,
			Status:    domain.OrderStatusInvalid,
			CreatedAt: time.Date(2025, 11, 1, 17, 16, 15, 0, time.UTC),
		},
	}
}

func getExampleAccruals(orders []domain.Order) map[domain.OrderID]float64 {
	numOrdersWithAccruals := max(1, rand.IntN(len(orders)))
	expectedAccruals := map[domain.OrderID]float64{}

	for range numOrdersWithAccruals {
		order := orders[rand.IntN(len(orders))]
		expectedAccruals[order.ID] = rand.Float64() * 100
	}

	return expectedAccruals
}
