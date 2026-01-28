package fakes_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	fakes "github.com/bq2cd/yp-go-gophermart/internal/test/fakes/gophermart/service/workers"
)

var _ = Describe("OrderRepository", func() {
	var (
		orderRepo  *fakes.TestOrderRepository
		testData   fakes.TestOrderRepoData
		testErrors fakes.TestOrderRepoErrorMap
		testDelays fakes.TestOrderRepoDelayMap
		userID     domain.UserID
		orderID    domain.OrderID
		err        error
		actionFn   func(context.Context)
		actionCtx  context.Context
	)

	BeforeEach(func() {
		orderRepo = fakes.NewTestOrderRepository()
		testData = fakes.NewTestOrderRepoData()
		testErrors = fakes.TestOrderRepoErrorMap{}
		testDelays = fakes.TestOrderRepoDelayMap{}
		userID = domain.UserID("user1")
		orderID = domain.OrderID(12345)
		actionCtx = GinkgoT().Context() //nolint:fatcontext
	})

	JustBeforeEach(func() {
		orderRepo.Setup(testData, testErrors, testDelays)

		actionFn(actionCtx)
	})

	whenOrderDoesNotExist := func(extraExpectations ...func()) {
		When("order does not exist", func() {
			It("should return ErrOrderNotFound", func() {
				Expect(err).To(MatchError(domain.ErrOrderNotFound))
				for _, expectFn := range extraExpectations {
					expectFn()
				}
			})
		})
	}

	whenOrderHasDifferentUserID := func(extraExpectations ...func()) {
		Context("order has different user ID", func() {
			BeforeEach(func() {
				userID = domain.UserID("user2")
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("user ID mismatch")))
				for _, expectFn := range extraExpectations {
					expectFn()
				}
			})
		})
	}

	whenActionContextExpired := func(extraExpectations ...func()) {
		Context("action context expired", func() {
			BeforeEach(func() {
				ctx, cancel := context.WithCancel(GinkgoT().Context())

				cancel()

				actionCtx = ctx //nolint:fatcontext
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("context canceled")))
				for _, expectFn := range extraExpectations {
					expectFn()
				}
			})
		})
	}

	Describe("GetProcessableOrdersPerUser", func() {
		var (
			actualOrders, expectedOrders map[domain.UserID][]domain.OrderID
		)

		BeforeEach(func() {
			actionFn = func(ctx context.Context) {
				actualOrders, err = orderRepo.GetProcessableOrdersPerUser(ctx)
			}
		})

		whenActionContextExpired(func() {
			Expect(actualOrders).To(BeEmpty())
		})

		When("users have some unprocessed orders", func() {
			BeforeEach(func() {
				testData = fakes.TestOrderRepoData{
					Orders: fakes.TestOrderMap{
						domain.OrderID(10_123): {UserID: "user1", Status: domain.OrderStatusNew},
						domain.OrderID(10_456): {UserID: "user1", Status: domain.OrderStatusProcessing},
						domain.OrderID(10_789): {UserID: "user1", Status: domain.OrderStatusProcessed},
						domain.OrderID(10_900): {UserID: "user1", Status: domain.OrderStatusInvalid},
						domain.OrderID(20_789): {UserID: "user2", Status: domain.OrderStatusProcessed},
						domain.OrderID(20_900): {UserID: "user2", Status: domain.OrderStatusInvalid},
					},
				}

				expectedOrders = map[domain.UserID][]domain.OrderID{
					"user1": {10_123, 10_456},
				}
			})
			It("should return only orders that need processing", func() {
				Expect(err).To(Succeed())

				for userID := range expectedOrders {
					Expect(actualOrders[userID]).To(ConsistOf(expectedOrders[userID]))
				}
			})
		})

		When("users have all orders processed", func() {
			BeforeEach(func() {
				testData = fakes.TestOrderRepoData{
					Orders: fakes.TestOrderMap{
						domain.OrderID(10_123): {UserID: "user1", Status: domain.OrderStatusInvalid},
						domain.OrderID(10_456): {UserID: "user1", Status: domain.OrderStatusProcessed},
						domain.OrderID(10_789): {UserID: "user1", Status: domain.OrderStatusProcessed},
						domain.OrderID(10_900): {UserID: "user1", Status: domain.OrderStatusInvalid},
						domain.OrderID(20_789): {UserID: "user2", Status: domain.OrderStatusProcessed},
						domain.OrderID(20_900): {UserID: "user2", Status: domain.OrderStatusInvalid},
					},
				}
			})
			It("should return empty result", func() {
				Expect(err).To(Succeed())
				Expect(actualOrders).To(BeEmpty())
			})
		})
	})

	Describe("GetOrderStatus", func() {
		var status, expectedStatus domain.OrderStatus

		BeforeEach(func() {
			actionFn = func(ctx context.Context) {
				status, err = orderRepo.GetOrderStatus(ctx, userID, orderID)
			}

			expectedStatus = domain.OrderStatusProcessing
		})

		whenOrderDoesNotExist(func() {
			Expect(status).To(Equal(domain.OrderStatusInvalid))
		})

		When("order exists", func() {
			expectError := func(msg string) {
				Expect(err).To(MatchError(ContainSubstring(msg)))
				Expect(status).To(Equal(domain.OrderStatusInvalid))
			}

			expectSuccess := func() {
				Expect(err).To(Succeed())
				Expect(status).To(Equal(expectedStatus))
			}

			BeforeEach(func() {
				testData = fakes.TestOrderRepoData{
					Orders: fakes.TestOrderMap{
						orderID: {UserID: userID, Status: expectedStatus},
					},
				}
			})

			whenOrderHasDifferentUserID(func() {
				Expect(status).To(Equal(domain.OrderStatusInvalid))
			})

			whenActionContextExpired(func() {
				Expect(status).To(Equal(domain.OrderStatusInvalid))
			})

			Context("no test errors configured", func() {
				It("should return order status", func() {
					expectSuccess()
				})
			})

			Context("some test errors configured", func() {
				BeforeEach(func() {
					testErrors = fakes.TestOrderRepoErrorMap{
						orderID: {GetOrderStatus: fakes.NewTestError(3, 0, 1)},
					}
				})

				It("should return configured errors", func() {
					expectError("test error 5")

					actionFn(actionCtx)
					expectError("test error 4")

					actionFn(actionCtx)
					expectError("test error 3")

					actionFn(actionCtx)
					expectSuccess()

					actionFn(actionCtx)
					expectError("test error 1")

					actionFn(actionCtx)
					expectSuccess()
				})
			})

			Context("some test delays configured", func() {
				var (
					expectedDuration time.Duration
					startTime        time.Time
				)

				BeforeEach(func() {
					startTime = time.Now()
					expectedDuration = 100 * time.Millisecond
					testDelays = fakes.TestOrderRepoDelayMap{
						orderID: {GetOrderStatus: expectedDuration},
					}
				})

				It("should eventually succeed", func() {
					elapsed := time.Since(startTime)

					expectSuccess()

					Expect(elapsed).To(BeNumerically(">=", expectedDuration))
				})
			})
		})
	})

	Context("updating orders", func() {
		var (
			initialOrderStatus, targetOrderStatus   domain.OrderStatus
			initialOrderAccrual, targetOrderAccrual float64
		)

		BeforeEach(func() {
			initialOrderAccrual = 0
			targetOrderAccrual = 0
		})

		initTestData := func(extraFn func() fakes.TestOrderRepoData) {
			BeforeEach(func() {
				testData.Merge(fakes.TestOrderRepoData{
					Orders: fakes.TestOrderMap{
						orderID: {UserID: userID, Status: initialOrderStatus, Accrual: initialOrderAccrual},
					},
				})
				if extraFn != nil {
					testData.Merge(extraFn())
				}
			})
		}

		expectOrderProperties := func(status domain.OrderStatus, accrual float64) {
			Expect(orderRepo.GetData().Orders[orderID].Status).To(Equal(status))
			Expect(orderRepo.GetData().Orders[orderID].Accrual).To(Equal(accrual))
		}

		expectError := func(msg string, status domain.OrderStatus, accrual float64) {
			Expect(err).To(MatchError(ContainSubstring(msg)))
			expectOrderProperties(status, accrual)
		}

		expectSuccess := func(status domain.OrderStatus, accrual float64) {
			Expect(err).To(Succeed())
			expectOrderProperties(status, accrual)
		}

		whenOrderAlreadyInTargetState := func() {
			Context("order already in target state", func() {
				BeforeEach(func() {
					testData.Merge(fakes.TestOrderRepoData{
						Orders: fakes.TestOrderMap{
							orderID: {UserID: userID, Status: targetOrderStatus, Accrual: targetOrderAccrual},
						},
					})
				})

				It("should succeed", func() {
					expectSuccess(targetOrderStatus, targetOrderAccrual)
				})
			})
		}

		whenNoTestErrorsConfigured := func() {
			Context("no test errors configured", func() {
				It("should succeed", func() {
					expectSuccess(targetOrderStatus, targetOrderAccrual)
				})
			})
		}

		whenTestErrorsAreConfigured := func(
			setupErr func(fakes.TestError) fakes.TestOrderRepoError,
			extraSuccessExpectations ...func(),
		) {
			Context("some test errors configured", func() {
				BeforeEach(func() {
					testErrors = fakes.TestOrderRepoErrorMap{
						orderID: setupErr(fakes.NewTestError(2)),
					}
				})

				It("should return configured errors", func() {
					expectError("test error 2", initialOrderStatus, initialOrderAccrual)

					actionFn(actionCtx)

					expectError("test error 1", initialOrderStatus, initialOrderAccrual)

					actionFn(actionCtx)

					expectSuccess(targetOrderStatus, targetOrderAccrual)
					for _, expectFn := range extraSuccessExpectations {
						expectFn()
					}
				})
			})
		}

		whenTestDelaysAreConfigured := func(
			setupDelay func(time.Duration) fakes.TestOrderRepoDelay,
			extraSuccessExpectations ...func(),
		) {
			Context("some test delays configured", func() {
				var (
					expectedDuration time.Duration
					startTime        time.Time
				)

				BeforeEach(func() {
					startTime = time.Now()
					expectedDuration = 100 * time.Millisecond
					testDelays = fakes.TestOrderRepoDelayMap{
						orderID: setupDelay(expectedDuration),
					}
				})

				It("should eventually succeed", func() {
					elapsed := time.Since(startTime)

					expectSuccess(targetOrderStatus, targetOrderAccrual)
					for _, expectFn := range extraSuccessExpectations {
						expectFn()
					}

					Expect(elapsed).To(BeNumerically(">=", expectedDuration))
				})
			})
		}

		Describe("MarkOrderInvalid", func() {
			BeforeEach(func() {
				actionFn = func(ctx context.Context) {
					err = orderRepo.MarkOrderInvalid(ctx, userID, orderID)
				}
				initialOrderStatus = domain.OrderStatusProcessing
				targetOrderStatus = domain.OrderStatusInvalid
			})

			whenOrderDoesNotExist()

			When("order exists", func() {
				initTestData(nil)

				whenOrderHasDifferentUserID()

				whenActionContextExpired()

				Context("order is in processed state", func() {
					BeforeEach(func() {
						initialOrderAccrual = 0.22

						testData.Merge(fakes.TestOrderRepoData{
							Orders: fakes.TestOrderMap{
								orderID: {
									UserID:  userID,
									Status:  domain.OrderStatusProcessed,
									Accrual: initialOrderAccrual,
								},
							},
						})
					})

					It("should return an error", func() {
						expectError("cannot mark processed order", domain.OrderStatusProcessed, initialOrderAccrual)
					})
				})

				whenOrderAlreadyInTargetState()

				whenNoTestErrorsConfigured()

				whenTestErrorsAreConfigured(
					func(testErr fakes.TestError) fakes.TestOrderRepoError {
						return fakes.TestOrderRepoError{MarkOrderInvalid: testErr}
					},
				)

				whenTestDelaysAreConfigured(
					func(delay time.Duration) fakes.TestOrderRepoDelay {
						return fakes.TestOrderRepoDelay{MarkOrderInvalid: delay}
					},
				)
			})

		})

		Describe("MarkOrderProcessing", func() {
			BeforeEach(func() {
				actionFn = func(ctx context.Context) {
					err = orderRepo.MarkOrderProcessing(ctx, userID, orderID)
				}
				initialOrderStatus = domain.OrderStatusNew
				targetOrderStatus = domain.OrderStatusProcessing
			})

			whenOrderDoesNotExist()

			When("order exists", func() {
				initTestData(nil)

				whenOrderHasDifferentUserID()

				whenActionContextExpired()

				DescribeTableSubtree("order is in processed or invalid state",
					func(status domain.OrderStatus) {
						BeforeEach(func() {
							testData.Merge(fakes.TestOrderRepoData{
								Orders: fakes.TestOrderMap{
									orderID: {UserID: userID, Status: status, Accrual: initialOrderAccrual},
								},
							})
						})

						It("should return an error", func() {
							expectError("cannot mark invalid or processed order", status, initialOrderAccrual)
						})
					},
					Entry("processed", domain.OrderStatusProcessed),
					Entry("invalid", domain.OrderStatusInvalid),
				)

				whenOrderAlreadyInTargetState()

				whenNoTestErrorsConfigured()

				whenTestErrorsAreConfigured(
					func(testErr fakes.TestError) fakes.TestOrderRepoError {
						return fakes.TestOrderRepoError{MarkOrderProcessing: testErr}
					},
				)

				whenTestDelaysAreConfigured(
					func(delay time.Duration) fakes.TestOrderRepoDelay {
						return fakes.TestOrderRepoDelay{MarkOrderProcessing: delay}
					},
				)
			})

		})

		Describe("MarkOrderProcessed", func() {
			var amount, initialBalance float64

			BeforeEach(func() {
				actionFn = func(ctx context.Context) {
					err = orderRepo.MarkOrderProcessed(ctx, userID, orderID, amount)
				}
				amount = 3.33
				initialBalance = 10.07
				initialOrderStatus = domain.OrderStatusNew
				targetOrderStatus = domain.OrderStatusProcessed
				targetOrderAccrual = amount
			})

			whenOrderDoesNotExist()

			When("order exists", func() {
				initTestData(func() fakes.TestOrderRepoData {
					return fakes.TestOrderRepoData{
						Balances: fakes.TestBalanceMap{userID: initialBalance},
					}
				})

				whenOrderHasDifferentUserID()

				whenActionContextExpired()

				Context("order is in invalid state", func() {
					BeforeEach(func() {
						testData.Merge(fakes.TestOrderRepoData{
							Orders: fakes.TestOrderMap{
								orderID: {UserID: userID, Status: domain.OrderStatusInvalid},
							},
						})
					})

					It("should return an error", func() {
						expectError("cannot mark invalid order", domain.OrderStatusInvalid, 0)
					})
				})

				Context("order already in processed state", func() {
					BeforeEach(func() {
						initialOrderAccrual = 1.11

						testData.Merge(fakes.TestOrderRepoData{
							Orders: fakes.TestOrderMap{
								orderID: {UserID: userID, Status: targetOrderStatus, Accrual: initialOrderAccrual},
							},
						})
					})

					It("should succeed", func() {
						expectSuccess(targetOrderStatus, initialOrderAccrual)
						Expect(orderRepo.GetData().Balances[userID]).To(Equal(initialBalance))
					})
				})

				DescribeTableSubtree("no test errors configured",
					func(status domain.OrderStatus) {
						BeforeEach(func() {
							testData.Merge(fakes.TestOrderRepoData{
								Orders: fakes.TestOrderMap{
									orderID: {UserID: userID, Status: status},
								},
							})
						})

						It("should succeed", func() {
							expectSuccess(targetOrderStatus, targetOrderAccrual)
							Expect(orderRepo.GetData().Balances[userID]).To(Equal(initialBalance + targetOrderAccrual))
						})
					},
					Entry("new -> processed", domain.OrderStatusNew),
					Entry("processing -> processed", domain.OrderStatusProcessing),
				)

				whenTestErrorsAreConfigured(
					func(testErr fakes.TestError) fakes.TestOrderRepoError {
						return fakes.TestOrderRepoError{MarkOrderProcessed: testErr}
					},
					func() {
						Expect(orderRepo.GetData().Balances[userID]).To(Equal(initialBalance + targetOrderAccrual))
					},
				)

				whenTestDelaysAreConfigured(
					func(delay time.Duration) fakes.TestOrderRepoDelay {
						return fakes.TestOrderRepoDelay{MarkOrderProcessed: delay}
					},
					func() {
						Expect(orderRepo.GetData().Balances[userID]).To(Equal(initialBalance + targetOrderAccrual))
					},
				)
			})
		})
	})
})
