package workers_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

// Ensure [workers.OrderProcessor] implements [service.OrderProcessor].
var _ service.OrderProcessor = (*workers.OrderProcessor)(nil)

var _ = Describe("OrderProcessor", MustPassRepeatedly(5), func() {
	testutil.MaybeEnableDebugLogging()

	var (
		orderRepo             *fakes.TestOrderRepository
		accrualClient         *fakes.TestAccrualClient
		orderQueue            workers.OrderQueue
		orderProcessorOptions []option.Option[workers.OrderProcessor]
		orderProcessor        *workers.OrderProcessor
	)

	BeforeEach(func() {
		orderRepo = fakes.NewTestOrderRepository()
		accrualClient = fakes.NewTestAccrualClient()
		orderQueue = workers.NewOrderQueue()
		orderProcessorOptions = []option.Option[workers.OrderProcessor]{}
	})

	JustBeforeEach(func() {
		orderProcessor = workers.NewOrderProcessor(
			orderRepo,
			orderQueue,
			accrualClient,
			orderProcessorOptions...,
		)
	})

	Describe("enqueuing orders", func() {
		Context("single order", func() {
			var (
				userID                  domain.UserID
				orderID                 domain.OrderID
				accepted                bool
				shutdownBeforeEnqueuing bool
			)

			BeforeEach(func() {
				userID = domain.UserID("user1")
				orderID = domain.OrderID(12334)
			})

			JustBeforeEach(func() {
				if shutdownBeforeEnqueuing {
					ensureOrderProcessorIsClosed(orderProcessor)
				}

				accepted = orderProcessor.EnqueueOrder(userID, orderID)
			})

			When("order processor is running", func() {
				It("should be accepted", func() {
					Expect(accepted).To(BeTrue())
				})
			})

			When("order processor has been shutdown", func() {
				BeforeEach(func() {
					shutdownBeforeEnqueuing = true
				})

				It("should be rejected", func() {
					Expect(accepted).To(BeFalse())
				})
			})
		})
	})

	Describe("processing orders", func() {
		var (
			testCtx    *TestOrderProcessorContext
			runTimeout time.Duration
		)

		BeforeEach(func() {
			testCtx = NewTestOrderProcessorContext()
			runTimeout = 100 * time.Millisecond
			orderProcessorOptions = append(orderProcessorOptions,
				workers.WithOrderProcessorDelayConfig(
					workers.NewOrderEventDelayConfig(
						workers.WithOrderEventInitialDelay(10*time.Millisecond),
						workers.WithOrderEventMaxJitter(5*time.Millisecond),
						workers.WithOrderEventMaxDelay(1*time.Second),
					),
				),
			)
		})

		JustBeforeEach(func() {
			testCtx.SetupMocks(orderRepo, accrualClient)
			testCtx.EnqueueOrders(orderProcessor)

			ctx, cancel := context.WithTimeout(GinkgoT().Context(), runTimeout)
			defer cancel()

			go orderProcessor.Run(ctx)

			Eventually(orderProcessor.HasFinished).To(BeTrue(), "order processor should have finished by now")
		})

		Context("all orders are new", func() {
			BeforeEach(func() {
				testCtx.InputOrders = []domain.OrderID{10_123, 10_789, 20_123, 20_456, 20_789, 30_456, 40_123, 99_999}
				testCtx.OrderRepoData = fakes.TestOrderRepoData{
					Balances: fakes.TestBalanceMap{
						"user1": 0.0,
						"user2": 9.99,
					},
					Orders: fakes.TestOrderMap{
						10_123: {UserID: "user1", Status: domain.OrderStatusNew},
						10_456: {UserID: "user1", Status: domain.OrderStatusNew},
						10_789: {UserID: "user1", Status: domain.OrderStatusNew},
						20_123: {UserID: "user2", Status: domain.OrderStatusNew},
						20_456: {UserID: "user2", Status: domain.OrderStatusNew},
						20_789: {UserID: "user2", Status: domain.OrderStatusNew},
						30_123: {UserID: "user3", Status: domain.OrderStatusNew},
						30_456: {UserID: "user3", Status: domain.OrderStatusNew},
						30_789: {UserID: "user3", Status: domain.OrderStatusNew},
						40_123: {UserID: "user4", Status: domain.OrderStatusNew},
					},
				}
				testCtx.AccrualData = fakes.TestAccrualData{
					10_123: {Status: accdomain.OrderStatusRegistered},
					10_789: {Status: accdomain.OrderStatusProcessed, Accrual: 7.77},
					20_123: {Status: accdomain.OrderStatusProcessing},
					20_456: {Status: accdomain.OrderStatusProcessed, Accrual: 0.02},
					20_789: {Status: accdomain.OrderStatusInvalid},
					30_456: {Status: accdomain.OrderStatusRegistered},
				}
				testCtx.OrderRepoErrors = fakes.TestOrderRepoErrorMap{
					10_789: {GetOrderStatus: fakes.NewTestError(2)},
					20_456: {MarkOrderProcessing: fakes.NewTestError(1)},
					30_456: {MarkOrderProcessing: fakes.NewTestError(0, 1)},
				}
				testCtx.AccrualErrors = fakes.TestAccrualErrorMap{
					10_123: {GetOrderStatus: fakes.NewTestError(3)},
				}
			})

			It("should process only orders uploaded to accrual server", func() {
				orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
					Balances: fakes.TestBalanceMap{
						"user1": 7.77,
						"user2": 10.01,
					},
					Orders: fakes.TestOrderMap{
						10_123: {UserID: "user1", Status: domain.OrderStatusProcessing},
						10_456: {UserID: "user1", Status: domain.OrderStatusNew},
						10_789: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 7.77},
						20_123: {UserID: "user2", Status: domain.OrderStatusProcessing},
						20_456: {UserID: "user2", Status: domain.OrderStatusProcessed, Accrual: 0.02},
						20_789: {UserID: "user2", Status: domain.OrderStatusInvalid},
						30_123: {UserID: "user3", Status: domain.OrderStatusNew},
						30_456: {UserID: "user3", Status: domain.OrderStatusProcessing},
						30_789: {UserID: "user3", Status: domain.OrderStatusNew},
						40_123: {UserID: "user4", Status: domain.OrderStatusNew},
					},
				})
			})
		})

		Context("picking up previous state", func() {
			BeforeEach(func() {
				testCtx.InputOrders = []domain.OrderID{
					99_998,
					10_123,
					10_456,
					20_123,
					20_456,
					30_123,
					40_123,
					50_789,
					99_999,
				}
				testCtx.OrderRepoData = fakes.TestOrderRepoData{
					Balances: fakes.TestBalanceMap{
						"user1": 5.0,
						"user2": 10.0,
						"user4": 20.0,
					},
					Orders: fakes.TestOrderMap{
						10_123: {UserID: "user1", Status: domain.OrderStatusInvalid},
						10_456: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 12.2},
						20_123: {UserID: "user2", Status: domain.OrderStatusNew},
						20_456: {UserID: "user2", Status: domain.OrderStatusNew},
						30_123: {UserID: "user3", Status: domain.OrderStatusNew},
						40_123: {UserID: "user4", Status: domain.OrderStatusProcessing},
						40_789: {UserID: "user4", Status: domain.OrderStatusNew},
						50_789: {UserID: "user5", Status: domain.OrderStatus(-99)},
					},
				}
				testCtx.AccrualData = fakes.TestAccrualData{
					10_123: {Status: accdomain.OrderStatusRegistered},
					10_456: {Status: accdomain.OrderStatusProcessing},
					20_123: {Status: accdomain.OrderStatusProcessed, Accrual: 0.0},
					20_456: {Status: accdomain.OrderStatusProcessed, Accrual: 5.55},
					30_123: {Status: accdomain.OrderStatus(-99)},
					40_123: {Status: accdomain.OrderStatusProcessed, Accrual: 17.5},
				}
				testCtx.OrderRepoErrors = fakes.TestOrderRepoErrorMap{
					20_456: {MarkOrderProcessed: fakes.NewTestError(2)},
					30_123: {MarkOrderInvalid: fakes.NewTestError(2)},
				}
				testCtx.AccrualErrors = fakes.TestAccrualErrorMap{
					40_123: {GetOrderStatus: fakes.NewTestError(2)},
				}
			})

			It("should process only orders that are eligible for processing", func() {
				orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
					Balances: fakes.TestBalanceMap{
						"user1": 5.0,
						"user2": 15.55,
						"user4": 37.5,
					},
					Orders: fakes.TestOrderMap{
						10_123: {UserID: "user1", Status: domain.OrderStatusInvalid},
						10_456: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 12.2},
						20_123: {UserID: "user2", Status: domain.OrderStatusProcessed, Accrual: 0.0},
						20_456: {UserID: "user2", Status: domain.OrderStatusProcessed, Accrual: 5.55},
						30_123: {UserID: "user3", Status: domain.OrderStatusInvalid},
						40_123: {UserID: "user4", Status: domain.OrderStatusProcessed, Accrual: 17.5},
						40_789: {UserID: "user4", Status: domain.OrderStatusNew},
						50_789: {UserID: "user5", Status: domain.OrderStatus(-99)},
					},
				})
			})

		})

		Context("hitting delays during processing", func() {
			BeforeEach(func() {
				runTimeout = 80 * time.Millisecond

				orderProcessorOptions = append(orderProcessorOptions,
					workers.WithOrderProcessorShutdownTimeout(50*time.Millisecond),
				)

				testCtx.InputOrders = []domain.OrderID{
					10_123,
					20_456,
				}
				testCtx.OrderRepoData = fakes.TestOrderRepoData{
					Balances: fakes.TestBalanceMap{
						"user1": 5.0,
						"user2": 0.0,
					},
					Orders: fakes.TestOrderMap{
						10_123: {UserID: "user1", Status: domain.OrderStatusNew},
						20_456: {UserID: "user1", Status: domain.OrderStatusNew},
					},
				}
				testCtx.AccrualData = fakes.TestAccrualData{
					10_123: {Status: accdomain.OrderStatusProcessed, Accrual: 8.21},
					20_456: {Status: accdomain.OrderStatusProcessed, Accrual: 4.44},
				}
				testCtx.OrderRepoDelays = fakes.TestOrderRepoDelayMap{
					10_123: {
						GetOrderStatus:     20 * time.Millisecond,
						MarkOrderProcessed: 30 * time.Millisecond,
					},
				}
				testCtx.AccrualDelays = fakes.TestAccrualDelayMap{
					10_123: {GetOrderStatus: 50 * time.Millisecond},
				}
			})

			When("order processor has single worker", func() {
				BeforeEach(func() {
					orderProcessorOptions = append(orderProcessorOptions,
						workers.WithOrderProcessorWorkerPoolSize(1),
					)
				})

				It("should process only single order", func() {
					orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
						Balances: fakes.TestBalanceMap{
							"user1": 13.21,
							"user2": 0.0,
						},
						Orders: fakes.TestOrderMap{
							10_123: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 8.21},
							20_456: {UserID: "user1", Status: domain.OrderStatusNew},
						},
					})
				})
			})

			When("order processor has two workers", func() {
				BeforeEach(func() {
					orderProcessorOptions = append(orderProcessorOptions,
						workers.WithOrderProcessorWorkerPoolSize(2),
					)
				})

				It("should process two orders", func() {
					orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
						Balances: fakes.TestBalanceMap{
							"user1": 17.65,
							"user2": 0.0,
						},
						Orders: fakes.TestOrderMap{
							10_123: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 8.21},
							20_456: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 4.44},
						},
					})
				})
			})
		})

		Context("consistently hitting per-event timeouts during processing", func() {
			DescribeTableSubtree("with number of workers",
				func(_numWorkers int) {
					BeforeEach(func() {
						runTimeout = 500 * time.Millisecond

						orderProcessorOptions = append(orderProcessorOptions,
							workers.WithOrderProcessorPerEventTimeout(20*time.Millisecond),
							workers.WithOrderProcessorShutdownTimeout(100*time.Millisecond),
							workers.WithOrderProcessorWorkerPoolSize(uint(_numWorkers)),
						)

						testCtx.InputOrders = []domain.OrderID{
							10_123,
							10_456,
							10_789,
							20_123,
							20_456,
							20_789,
						}
						testCtx.OrderRepoData = fakes.TestOrderRepoData{
							Balances: fakes.TestBalanceMap{
								"user1": 5.0,
							},
							Orders: fakes.TestOrderMap{
								10_123: {UserID: "user1", Status: domain.OrderStatusNew},
								10_456: {UserID: "user1", Status: domain.OrderStatusNew},
								10_789: {UserID: "user1", Status: domain.OrderStatusNew},
								20_123: {UserID: "user2", Status: domain.OrderStatusNew},
								20_456: {UserID: "user2", Status: domain.OrderStatusNew},
								20_789: {UserID: "user2", Status: domain.OrderStatusNew},
							},
						}
						testCtx.AccrualData = fakes.TestAccrualData{
							10_123: {Status: accdomain.OrderStatusProcessed, Accrual: 8.21},
							10_456: {Status: accdomain.OrderStatusProcessed, Accrual: 4.44},
							10_789: {Status: accdomain.OrderStatusRegistered},
							20_123: {Status: accdomain.OrderStatusRegistered},
							20_456: {Status: accdomain.OrderStatusRegistered},
							20_789: {Status: accdomain.OrderStatusRegistered},
						}
						testCtx.OrderRepoDelays = fakes.TestOrderRepoDelayMap{
							10_123: {GetOrderStatus: 50 * time.Millisecond},
							10_456: {GetOrderStatus: 50 * time.Millisecond},
							10_789: {GetOrderStatus: 50 * time.Millisecond},
						}
						testCtx.AccrualDelays = fakes.TestAccrualDelayMap{
							20_123: {GetOrderStatus: 50 * time.Millisecond},
							20_456: {GetOrderStatus: 50 * time.Millisecond},
							20_789: {GetOrderStatus: 50 * time.Millisecond},
						}
					})

					It("should not process any of the orders", func() {
						orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
							Balances: fakes.TestBalanceMap{
								"user1": 5.0,
							},
							Orders: fakes.TestOrderMap{
								10_123: {UserID: "user1", Status: domain.OrderStatusNew},
								10_456: {UserID: "user1", Status: domain.OrderStatusNew},
								10_789: {UserID: "user1", Status: domain.OrderStatusNew},
								20_123: {UserID: "user2", Status: domain.OrderStatusNew},
								20_456: {UserID: "user2", Status: domain.OrderStatusNew},
								20_789: {UserID: "user2", Status: domain.OrderStatusNew},
							},
						})
					})
				},
				Entry(nil, 1),
				Entry(nil, 2),
				Entry(nil, 3),
				Entry(nil, 4),
				Entry(nil, 5),
				Entry(nil, 6),
			)

		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func ensureOrderProcessorIsClosed(
	orderProcessor *workers.OrderProcessor,
) {
	GinkgoHelper()

	ctx, cancel := context.WithCancel(GinkgoT().Context())

	go orderProcessor.Run(ctx)

	cancel()

	Eventually(orderProcessor.IsClosed).To(BeTrue())
}
