package sqldatabase_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/generated"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase/models"
)

var _ = Describe("StoreBalance", func() {
	var storage *sqldatabase.Storage

	Context("sqlite database", func() {
		var (
			user  models.User
			query gorm.Interface[models.Balance]
		)

		BeforeEach(OncePerOrdered, func() {
			tempStorage := createTempStorage()
			DeferCleanup(tempStorage.Cleanup)

			storage = tempStorage.Storage
			user = ensureUserExists(storage, exampleUserLogin)
		})

		BeforeEach(func() {
			query = sqldatabase.Query[models.Balance](storage)
		})

		Context("user's balance", func() {
			var (
				expected, actual float64
				err              error
				targetLogin      string
			)

			BeforeEach(OncePerOrdered, func() {
				expected = 2.34
			})

			BeforeEach(func() {
				targetLogin = user.Login
			})

			DescribeTableSubtree("getting balance",
				func(_setupFn func(), _actionFn func() (float64, error)) {
					BeforeEach(func() {
						_setupFn()
					})

					JustBeforeEach(func() {
						actual, err = _actionFn()
					})

					When("user exists", func() {
						It("should return correct value", func() {
							Expect(err).To(Succeed())
							Expect(actual).To(Equal(expected))
						})
					})

					When("user does not exist", func() {
						BeforeEach(func() {
							targetLogin = exampleUserNonExistentLogin
						})

						It("should return zero", func() {
							Expect(err).To(Succeed())
							Expect(actual).To(BeZero())
						})
					})
				},
				Entry("current",
					func() {
						ensureBalanceExists(storage, user, expected, 0)
					},
					func() (float64, error) {
						return storage.GetCurrentValue(GinkgoT().Context(), domain.UserID(targetLogin))
					}),
				Entry("withdrawn",
					func() {
						ensureBalanceExists(storage, user, 0, expected)
					},
					func() (float64, error) {
						return storage.GetTotalAmountWithdrawn(GinkgoT().Context(), domain.UserID(targetLogin))
					}),
			)

		})

		Context("withdrawing funds", Ordered, func() {
			var (
				initialCurrent, initialWithdrawn   float64
				expectedCurrent, expectedWithdrawn float64
				amount                             float64
				success                            bool
				err                                error
			)

			BeforeAll(func() {
				initialCurrent = 20.26
				initialWithdrawn = 7.55

				ensureBalanceExists(storage, user, initialCurrent, initialWithdrawn)
			})

			validateBalanceData := func() {
				GinkgoHelper()

				balance, err := query.Where(generated.Balance.UserID.Eq(user.ID)).First(GinkgoT().Context())
				Expect(err).To(Succeed())
				Expect(balance.Current).To(BeNumerically("~", expectedCurrent, 1e-12))
				Expect(balance.Withdrawn).To(BeNumerically("~", expectedWithdrawn, 1e-12))
			}

			When("existing user performs a withdrawal", func() {
				type Want struct {
					Current   float64
					Withdrawn float64
				}
				type Match struct {
					Success OmegaMatcher
					Error   OmegaMatcher
				}

				DescribeTableSubtree(
					"amount",
					func(orderID uint, _amount float64, want Want, match Match) {
						BeforeEach(func() {
							amount = _amount
							expectedCurrent = want.Current
							expectedWithdrawn = want.Withdrawn
						})

						JustBeforeEach(func() {
							success, err = storage.WithdrawFunds(
								GinkgoT().Context(),
								domain.UserID(user.Login),
								domain.OrderID(orderID),
								amount,
							)
						})

						It("should finish with valid balance", func() {
							Expect(err).To(match.Error)
							Expect(success).To(match.Success)

							validateBalanceData()
						})
					},
					Entry(
						"once OK",
						exampleOrderID+1,
						5.25,
						Want{Current: 15.01, Withdrawn: 12.8},
						Match{Success: BeTrue(), Error: Succeed()},
					),
					Entry(
						"twice OK",
						exampleOrderID+2,
						7.41,
						Want{Current: 7.6, Withdrawn: 20.21},
						Match{Success: BeTrue(), Error: Succeed()},
					),
					Entry(
						"thrice OK",
						exampleOrderID+3,
						3.5,
						Want{Current: 4.1, Withdrawn: 23.71},
						Match{Success: BeTrue(), Error: Succeed()},
					),
					Entry(
						"oops, duplicate order",
						exampleOrderID+3,
						2.2,
						Want{Current: 4.1, Withdrawn: 23.71},
						Match{Success: BeFalse(), Error: MatchError(ContainSubstring("UNIQUE constraint failed"))},
					),
					Entry(
						"oops, not enough funds",
						exampleOrderID+4,
						10.1,
						Want{Current: 4.1, Withdrawn: 23.71},
						Match{Success: BeFalse(), Error: MatchError(domain.ErrBalanceNotEnoughFunds)},
					),
					Entry(
						"oops, not enough funds, 1e-4",
						exampleOrderID+5,
						4.1001,
						Want{Current: 4.1, Withdrawn: 23.71},
						Match{Success: BeFalse(), Error: MatchError(domain.ErrBalanceNotEnoughFunds)},
					),
					Entry(
						"oops, not enough funds, 1e-8",
						exampleOrderID+6,
						4.10000001,
						Want{Current: 4.1, Withdrawn: 23.71},
						Match{Success: BeFalse(), Error: MatchError(domain.ErrBalanceNotEnoughFunds)},
					),
				)
			})

			When("existing user lists withdrawal transactions", func() {
				It("should return transactions from the newest to the oldest", func() {
					transactions, err := storage.GetWithdrawalTransactions(
						GinkgoT().Context(),
						domain.UserID(user.Login),
					)
					Expect(err).To(Succeed())
					Expect(transactions).To(HaveLen(3))
					Expect(transactions[0].OrderID.Uint()).To(Equal(exampleOrderID + 3))
					Expect(transactions[0].Amount).To(BeNumerically("~", 3.5, 1e-12))
					Expect(transactions[1].OrderID.Uint()).To(Equal(exampleOrderID + 2))
					Expect(transactions[1].Amount).To(BeNumerically("~", 7.41, 1e-12))
					Expect(transactions[2].OrderID.Uint()).To(Equal(exampleOrderID + 1))
					Expect(transactions[2].Amount).To(BeNumerically("~", 5.25, 1e-12))
				})
			})

			Context("non-existent user", func() {
				var targetLogin string

				BeforeEach(func() {
					targetLogin = exampleUserNonExistentLogin
					amount = 5.55
				})

				When("performing a withdrawal", func() {
					It("should return ErrUserNotFound", func() {
						success, err = storage.WithdrawFunds(
							GinkgoT().Context(),
							domain.UserID(targetLogin),
							domain.OrderID(exampleOrderID),
							amount,
						)
						Expect(err).To(MatchError(domain.ErrUserNotFound))
						Expect(success).To(BeFalse())
					})
				})

				When("listing withdrawal transactions", func() {
					It("should return empty array", func() {
						transactions, err := storage.GetWithdrawalTransactions(
							GinkgoT().Context(),
							domain.UserID(targetLogin),
						)
						Expect(err).To(Succeed())
						Expect(transactions).To(BeEmpty())
					})
				})
			})
		})

	})
})

func ensureBalanceExists(storage *sqldatabase.Storage, user models.User, current, withdrawn float64) {
	GinkgoHelper()

	balances, err := sqldatabase.Query[models.Balance](storage).
		Where(generated.Balance.UserID.Eq(user.ID)).
		Find(GinkgoT().Context())
	Expect(err).To(Succeed())
	Expect(balances).To(HaveLen(1))

	balance := balances[0]
	balance.Current = current
	balance.Withdrawn = withdrawn

	_, err = sqldatabase.Query[models.Balance](storage.Debug()).
		Updates(GinkgoT().Context(), balance)
	Expect(err).To(Succeed())
}
