package mocks_test

import (
	"context"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers/mocks"
)

var _ = Describe("OrderRepository", func() {
	var (
		orderRepo  *mocks.TestOrderRepository
		testData   mocks.TestOrderRepoData
		testErrors mocks.TestOrderRepoErrorMap
		testDelays mocks.TestOrderRepoDelayMap
		userID     domain.UserID
		orderID    domain.OrderID
		err        error
		actionFn   func(context.Context)
		actionCtx  context.Context
	)

	BeforeEach(func() {
		orderRepo = mocks.NewTestOrderRepository()
		testData = mocks.NewTestOrderRepoData()
		testErrors = mocks.TestOrderRepoErrorMap{}
		testDelays = mocks.TestOrderRepoDelayMap{}
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
				testData = mocks.TestOrderRepoData{
					Orders: mocks.TestOrderMap{
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
					testErrors = mocks.TestOrderRepoErrorMap{
						orderID: {GetOrderStatus: mocks.NewTestError(3, 0, 1)},
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
					testDelays = mocks.TestOrderRepoDelayMap{
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

		initTestData := func(extraFn func() mocks.TestOrderRepoData) {
			BeforeEach(func() {
				testData.Merge(mocks.TestOrderRepoData{
					Orders: mocks.TestOrderMap{
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
					testData.Merge(mocks.TestOrderRepoData{
						Orders: mocks.TestOrderMap{
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
			setupErr func(mocks.TestError) mocks.TestOrderRepoError,
			extraSuccessExpectations ...func(),
		) {
			Context("some test errors configured", func() {
				BeforeEach(func() {
					testErrors = mocks.TestOrderRepoErrorMap{
						orderID: setupErr(mocks.NewTestError(2)),
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
			setupDelay func(time.Duration) mocks.TestOrderRepoDelay,
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
					testDelays = mocks.TestOrderRepoDelayMap{
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

						testData.Merge(mocks.TestOrderRepoData{
							Orders: mocks.TestOrderMap{
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
					func(testErr mocks.TestError) mocks.TestOrderRepoError {
						return mocks.TestOrderRepoError{MarkOrderInvalid: testErr}
					},
				)

				whenTestDelaysAreConfigured(
					func(delay time.Duration) mocks.TestOrderRepoDelay {
						return mocks.TestOrderRepoDelay{MarkOrderInvalid: delay}
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
							testData.Merge(mocks.TestOrderRepoData{
								Orders: mocks.TestOrderMap{
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
					func(testErr mocks.TestError) mocks.TestOrderRepoError {
						return mocks.TestOrderRepoError{MarkOrderProcessing: testErr}
					},
				)

				whenTestDelaysAreConfigured(
					func(delay time.Duration) mocks.TestOrderRepoDelay {
						return mocks.TestOrderRepoDelay{MarkOrderProcessing: delay}
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
				initTestData(func() mocks.TestOrderRepoData {
					return mocks.TestOrderRepoData{
						Balances: mocks.TestBalanceMap{userID: initialBalance},
					}
				})

				whenOrderHasDifferentUserID()

				whenActionContextExpired()

				Context("order is in invalid state", func() {
					BeforeEach(func() {
						testData.Merge(mocks.TestOrderRepoData{
							Orders: mocks.TestOrderMap{
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

						testData.Merge(mocks.TestOrderRepoData{
							Orders: mocks.TestOrderMap{
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
							testData.Merge(mocks.TestOrderRepoData{
								Orders: mocks.TestOrderMap{
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
					func(testErr mocks.TestError) mocks.TestOrderRepoError {
						return mocks.TestOrderRepoError{MarkOrderProcessed: testErr}
					},
					func() {
						Expect(orderRepo.GetData().Balances[userID]).To(Equal(initialBalance + targetOrderAccrual))
					},
				)

				whenTestDelaysAreConfigured(
					func(delay time.Duration) mocks.TestOrderRepoDelay {
						return mocks.TestOrderRepoDelay{MarkOrderProcessed: delay}
					},
					func() {
						Expect(orderRepo.GetData().Balances[userID]).To(Equal(initialBalance + targetOrderAccrual))
					},
				)
			})
		})
	})

})
