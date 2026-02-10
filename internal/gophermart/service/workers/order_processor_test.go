package workers_test

import (
	"context"
	"log/slog"
	"math"
	"sync"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	accdomain "github.com/bq2cd/yp-go-gophermart/internal/accrual/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
	"github.com/bq2cd/yp-go-gophermart/pkg/option"
)

// Ensure [workers.OrderProcessor] implements [service.OrderProcessor].
var _ service.OrderProcessor = (*workers.OrderProcessor)(nil)

var _ = Describe("OrderProcessor", MustPassRepeatedly(5), func() {
	var (
		orderRepo             *fakes.TestOrderRepository
		accrualClient         *fakes.TestAccrualClient
		orderProcessorOptions []option.Option[workers.OrderProcessor]
		orderProcessor        *workers.OrderProcessor
		mu                    sync.Mutex
	)

	BeforeEach(func() {
		mu.Lock()
		defer mu.Unlock()

		orderRepo = fakes.NewTestOrderRepository()
		accrualClient = fakes.NewTestAccrualClient()
		orderProcessorOptions = []option.Option[workers.OrderProcessor]{}
	})

	JustBeforeEach(func() {
		orderProcessor = workers.NewOrderProcessor(
			orderRepo,
			orderRepo,
			accrualClient,
			orderProcessorOptions...,
		)
	})

	Describe("processing orders", func() {
		var (
			testCtx               *TestOrderProcessorContext
			runLeeway, runTimeout time.Duration
		)

		BeforeEach(func() {
			testCtx = NewTestOrderProcessorContext()
			runLeeway = 0
			runTimeout = 100 * time.Millisecond
			orderProcessorOptions = append(orderProcessorOptions,
				workers.WithOrderProcessorDelayConfig(
					workers.NewOrderEventDelayConfig(
						workers.WithOrderEventInitialDelay(10*time.Millisecond),
						workers.WithOrderEventMaxJitter(5*time.Millisecond),
						workers.WithOrderEventMaxDelay(1*time.Second),
						workers.WithOrderEventMinRetriesUntilBackOff(10),
					),
				),
			)
		})

		JustBeforeEach(func() {
			testCtx.SetupMocks(orderRepo, accrualClient)

			ctx, cancel := context.WithTimeout(GinkgoT().Context(), runTimeout)
			defer cancel()

			go orderProcessor.Run(ctx)

			time.Sleep(runLeeway)

			Eventually(orderProcessor.HasFinished).To(BeTrue(), "order processor should have finished by now")
		})

		Context("all orders are new", func() {
			BeforeEach(func() {
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
				runTimeout = 100 * time.Millisecond

				orderProcessorOptions = append(orderProcessorOptions,
					workers.WithOrderProcessorShutdownTimeout(50*time.Millisecond),
				)

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
						GetOrderStatus:     30 * time.Millisecond,
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

		Context("accrual server constantly overloaded", func() {
			DescribeTableSubtree("with number of workers",
				func(_numWorkers int) {
					BeforeEach(func() {
						runTimeout = 500 * time.Millisecond

						orderProcessorOptions = append(orderProcessorOptions,
							workers.WithOrderProcessorPerEventTimeout(600*time.Millisecond),
							workers.WithOrderProcessorShutdownTimeout(100*time.Millisecond),
							workers.WithOrderProcessorWorkerPoolSize(uint(_numWorkers)),
						)

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

						accrualError := fakes.NewTestErrorCustom(
							&accdomain.RateLimitExceededError{RetryAfter: 120 * time.Second},
							math.MaxUint,
						)

						testCtx.AccrualErrors = fakes.TestAccrualErrorMap{
							10_123: {GetOrderStatus: accrualError},
							10_456: {GetOrderStatus: accrualError},
							10_789: {GetOrderStatus: accrualError},
							20_123: {GetOrderStatus: accrualError},
							20_456: {GetOrderStatus: accrualError},
							20_789: {GetOrderStatus: accrualError},
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

						Expect(accrualClient.NumCalls().GetOrderStatus).To(BeNumerically("<=", _numWorkers))
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

		Context("order processor should sleep until next queue processing is triggered", func() {
			BeforeEach(func() {
				runTimeout = 500 * time.Millisecond
				runLeeway = 600 * time.Millisecond

				testCtx.OrderRepoData = fakes.TestOrderRepoData{
					Balances: fakes.TestBalanceMap{
						"user1": 0.0,
					},
					Orders: fakes.TestOrderMap{
						10_123: {UserID: "user1", Status: domain.OrderStatusNew},
					},
				}
				testCtx.AccrualData = fakes.TestAccrualData{
					10_123: {Status: accdomain.OrderStatusProcessed, Accrual: 7.77},
				}
				testCtx.AccrualErrors = fakes.TestAccrualErrorMap{
					10_123: {GetOrderStatus: fakes.NewTestError(1)},
				}
			})

			When("delays are too high", func() {
				BeforeEach(func() {
					orderProcessorOptions = append(orderProcessorOptions,
						workers.WithOrderProcessorDelayConfig(
							workers.NewOrderEventDelayConfig(
								workers.WithOrderEventInitialDelay(10*time.Minute),
								workers.WithOrderEventMaxJitter(5*time.Minute),
								workers.WithOrderEventMaxDelay(1*time.Hour),
								workers.WithOrderEventMinRetriesUntilBackOff(0),
							),
						),
					)
				})

				It("should wake up only once", func() {
					orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
						Balances: fakes.TestBalanceMap{
							"user1": 0.0,
						},
						Orders: fakes.TestOrderMap{
							10_123: {UserID: "user1", Status: domain.OrderStatusNew},
						},
					})

					Expect(orderProcessor.NumWakeups()).To(BeNumerically("==", 1))
				})
			})

			When("delays are moderate", func() {
				var maxJitter time.Duration

				BeforeEach(func() {
					maxJitter = 50 * time.Millisecond

					orderProcessorOptions = append(orderProcessorOptions,
						workers.WithOrderProcessorDelayConfig(
							workers.NewOrderEventDelayConfig(
								workers.WithOrderEventInitialDelay(50*time.Millisecond),
								workers.WithOrderEventMaxJitter(maxJitter),
								workers.WithOrderEventMaxDelay(1*time.Second),
								workers.WithOrderEventMinRetriesUntilBackOff(100),
							),
						),
					)
				})

				Context("a single order is inflight", func() {
					It("should wake up two times", func() {
						orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
							Balances: fakes.TestBalanceMap{
								"user1": 7.77,
							},
							Orders: fakes.TestOrderMap{
								10_123: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 7.77},
							},
						})

						Expect(accrualClient.NumCalls().GetOrderStatus).To(BeNumerically("==", 2))
						Expect(orderProcessor.NumWakeups()).To(BeNumerically("==", 2))
					})
				})

				Context("two orders are inflight on a single worker", func() {
					BeforeEach(func() {
						orderProcessorOptions = append(orderProcessorOptions,
							workers.WithOrderProcessorWorkerPoolSize(1),
						)

						testCtx.OrderRepoData.Merge(fakes.TestOrderRepoData{
							Balances: fakes.TestBalanceMap{
								"user2": 0.0,
							},
							Orders: fakes.TestOrderMap{
								20_456: {UserID: "user2", Status: domain.OrderStatusNew},
							},
						})
						testCtx.AccrualData.Merge(fakes.TestAccrualData{
							20_456: {Status: accdomain.OrderStatusProcessing},
						})

						go func() {
							time.Sleep(time.Duration(float64(runTimeout) / 2.0))

							mu.Lock()
							defer mu.Unlock()

							accrualClient.InjectData(fakes.TestAccrualData{
								20_456: {Status: accdomain.OrderStatusInvalid},
							})

							slog.Debug("test: injected accrual client data")
						}()
					})

					It("should wake up multiple times until all orders are processed", func() {
						orderRepo.GetData().ExpectEqual(fakes.TestOrderRepoData{
							Balances: fakes.TestBalanceMap{
								"user1": 7.77,
								"user2": 0.0,
							},
							Orders: fakes.TestOrderMap{
								10_123: {UserID: "user1", Status: domain.OrderStatusProcessed, Accrual: 7.77},
								20_456: {UserID: "user2", Status: domain.OrderStatusInvalid},
							},
						})

						expectedWakeups := int(float64(runTimeout) / float64(maxJitter) / 2.0)
						expectedMaxWakeups := expectedWakeups * 3

						Expect(accrualClient.NumCalls().GetOrderStatus).To(BeNumerically(">=", expectedWakeups+2))
						Expect(accrualClient.NumCalls().GetOrderStatus).To(BeNumerically("<=", expectedMaxWakeups))

						Expect(orderProcessor.NumWakeups()).To(BeNumerically(">=", expectedWakeups))
						Expect(orderProcessor.NumWakeups()).To(BeNumerically("<=", expectedMaxWakeups))
					})
				})
			})
		})
	})
})
