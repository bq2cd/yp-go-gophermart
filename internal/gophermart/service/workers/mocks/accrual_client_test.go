package mocks_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers/mocks"
)

var _ = Describe("AccrualClient", func() {
	var (
		accrualClient *mocks.TestAccrualClient
		testData      mocks.TestAccrualData
		testErrors    mocks.TestAccrualErrorMap
		testDelays    mocks.TestAccrualDelayMap
	)

	BeforeEach(func() {
		accrualClient = mocks.NewTestAccrualClient()
		testData = mocks.TestAccrualData{}
		testErrors = mocks.TestAccrualErrorMap{}
		testDelays = mocks.TestAccrualDelayMap{}
	})

	Describe("GetOrderStatus", func() {
		var (
			orderID              accdomain.OrderID
			order, expectedOrder accdomain.Order
			err                  error
			actionFn             func(context.Context)
			actionCtx            context.Context
		)

		BeforeEach(func() {
			actionFn = func(ctx context.Context) {
				order, err = accrualClient.GetOrderStatus(ctx, orderID)
			}
			actionCtx = GinkgoT().Context() //nolint:fatcontext
			orderID = accdomain.OrderID(1234)
			expectedOrder = accdomain.Order{}
		})

		JustBeforeEach(func() {
			accrualClient.Setup(testData, testErrors, testDelays)

			actionFn(actionCtx)
		})

		JustAfterEach(func() {
			Expect(order).To(Equal(expectedOrder))
		})

		When("order does not exist", func() {
			It("should return ErrOrderNotFound", func() {
				Expect(err).To(MatchError(accdomain.ErrOrderNotFound))
			})
		})

		When("action context expired", func() {
			BeforeEach(func() {
				ctx, cancel := context.WithCancel(GinkgoT().Context())

				cancel()

				actionCtx = ctx //nolint:fatcontext
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
			})
		})

		Context("order exists", func() {
			BeforeEach(func() {
				expectedOrder = accdomain.Order{
					ID:            orderID,
					Status:        accdomain.OrderStatusProcessed,
					AccrualPoints: 3.25,
				}

				testData.Merge(mocks.TestAccrualData{
					domain.OrderID(orderID): {Status: expectedOrder.Status, Accrual: expectedOrder.AccrualPoints},
				})
			})

			When("no test errors are configured", func() {
				It("should return expected order data", func() {
					Expect(err).To(Succeed())
				})
			})

			When("some test errors are configured", func() {
				BeforeEach(func() {
					testErrors.Merge(mocks.TestAccrualErrorMap{
						domain.OrderID(orderID): {GetOrderStatus: mocks.NewTestError(2)},
					})
				})

				It("should return configured errors", func() {
					Expect(err).To(MatchError(ContainSubstring("test error 2")))
					Expect(order).To(Equal(accdomain.Order{}))

					actionFn(actionCtx)

					Expect(err).To(MatchError(ContainSubstring("test error 1")))
					Expect(order).To(Equal(accdomain.Order{}))

					actionFn(actionCtx)

					Expect(err).To(Succeed())
				})

			})

			When("some test delays configured", func() {
				var (
					expectedDuration time.Duration
					startTime        time.Time
				)

				BeforeEach(func() {
					startTime = time.Now()
					expectedDuration = 100 * time.Millisecond
					testDelays = mocks.TestAccrualDelayMap{
						domain.OrderID(orderID): {GetOrderStatus: expectedDuration},
					}
				})

				It("should eventually succeed", func() {
					elapsed := time.Since(startTime)

					Expect(err).To(Succeed())

					Expect(elapsed).To(BeNumerically(">=", expectedDuration))
				})
			})
		})
	})

})
