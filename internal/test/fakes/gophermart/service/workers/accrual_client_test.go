package fakes_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

var _ = Describe("AccrualClient", func() {
	var (
		accrualClient *fakes.TestAccrualClient
		testData      fakes.TestAccrualData
		testErrors    fakes.TestAccrualErrorMap
		testDelays    fakes.TestAccrualDelayMap
	)

	BeforeEach(func() {
		accrualClient = fakes.NewTestAccrualClient()
		testData = fakes.TestAccrualData{}
		testErrors = fakes.TestAccrualErrorMap{}
		testDelays = fakes.TestAccrualDelayMap{}
	})

	Describe("GetOrderStatus", func() {
		var (
			orderID              accdomain.OrderID
			order, expectedOrder accdomain.Order
			err                  error
			actionFn             func(context.Context)
			actionCtx            context.Context
			expectedNumCalls     uint
		)

		BeforeEach(func() {
			actionFn = func(ctx context.Context) {
				order, err = accrualClient.GetOrderStatus(ctx, orderID)
			}
			actionCtx = GinkgoT().Context() //nolint:fatcontext
			orderID = accdomain.OrderID(1234)
			expectedOrder = accdomain.Order{}
			expectedNumCalls = 1
		})

		JustBeforeEach(func() {
			accrualClient.Setup(testData, testErrors, testDelays)

			actionFn(actionCtx)
		})

		JustAfterEach(func() {
			Expect(order).To(Equal(expectedOrder))
			Expect(accrualClient.NumCalls().GetOrderStatus).To(Equal(expectedNumCalls))
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

				testData.Merge(fakes.TestAccrualData{
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
					testErrors.Merge(fakes.TestAccrualErrorMap{
						domain.OrderID(orderID): {GetOrderStatus: fakes.NewTestError(2)},
					})
					expectedNumCalls = 3
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
					testDelays = fakes.TestAccrualDelayMap{
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
