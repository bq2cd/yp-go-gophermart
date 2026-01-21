package service_test

import (
	"errors"
	"slices"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"go.uber.org/mock/gomock"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/handler"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	mocks "github.com/bq2cd/yp-go-gophermart/internal/test/mocks/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

// Ensure [service.BalanceManager] implements [handler.BalanceService].
var _ handler.BalanceService = (*service.BalanceManager)(nil)

var _ = Describe("BalanceManager", func() {
	var (
		balanceRepo *mocks.MockBalanceRepository
		balanceMgr  *service.BalanceManager
		userID      domain.UserID
		err         error
	)

	BeforeEach(func() {
		ctrl := gomock.NewController(GinkgoT())

		balanceRepo = mocks.NewMockBalanceRepository(ctrl)
		balanceMgr = service.NewBalanceManager(balanceRepo)
		userID = exampleUserID
	})

	Describe("getting current value of user's balance", func() {
		var (
			balance  float64
			mockCall *mocks.MockBalanceRepositoryGetCurrentValueCall
		)

		BeforeEach(func() {
			mockCall = balanceRepo.EXPECT().
				GetCurrentValue(testutil.MockCtx(), userID)
		})

		JustBeforeEach(func() {
			balance, err = balanceMgr.GetBalance(GinkgoT().Context(), userID)
		})

		When("balance repository is healthy", func() {
			var expectedBalance float64

			BeforeEach(func() {
				expectedBalance = 20.25

				mockCall.Return(expectedBalance, nil)
			})

			It("should return user's balance", func() {
				Expect(err).To(Succeed())
				Expect(balance).To(Equal(expectedBalance))
			})
		})

		When("balance repository fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("balance is unavailable")

				mockCall.Return(0, expectedErr)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(balance).To(BeZero())
			})
		})

	})

	Describe("getting total amount of withdrawals", func() {
		var (
			amount   float64
			mockCall *mocks.MockBalanceRepositoryGetTotalAmountWithdrawnCall
		)

		BeforeEach(func() {
			mockCall = balanceRepo.EXPECT().
				GetTotalAmountWithdrawn(testutil.MockCtx(), userID)
		})

		JustBeforeEach(func() {
			amount, err = balanceMgr.GetTotalAmountWithdrawn(GinkgoT().Context(), userID)
		})

		Context("balance repository is healthy", func() {
			JustBeforeEach(func() {
				Expect(err).To(Succeed())
			})

			When("user has made no withdrawals", func() {
				BeforeEach(func() {
					mockCall.Return(0, nil)
				})

				It("should return zero", func() {
					Expect(amount).To(BeZero())
				})
			})

			When("user has made some withdrawals", func() {
				BeforeEach(func() {
					mockCall.Return(20.0, nil)
				})

				It("should return total amount", func() {
					Expect(amount).To(Equal(20.0))
				})
			})
		})

		When("balance repository fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("withrawals transactions are unavailable")

				mockCall.Return(0, expectedErr)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(amount).To(BeZero())
			})
		})
	})

	Describe("getting all withdrawal transactions", func() {
		var (
			transactions []domain.WithdrawalTransaction
			mockCall     *mocks.MockBalanceRepositoryGetWithdrawalTransactionsCall
		)

		BeforeEach(func() {
			mockCall = balanceRepo.EXPECT().
				GetWithdrawalTransactions(testutil.MockCtx(), userID)
		})

		JustBeforeEach(func() {
			transactions, err = balanceMgr.GetWithdrawalTransactions(GinkgoT().Context(), userID)
		})

		Context("balance repository is healthy", func() {
			JustBeforeEach(func() {
				Expect(err).To(Succeed())
			})

			When("user has made no withdrawals", func() {
				BeforeEach(func() {
					mockCall.Return(nil, nil)
				})

				It("should return empty array", func() {
					Expect(transactions).To(BeEmpty())
				})
			})

			When("user has made some withdrawals", func() {
				var expectedTransactions []domain.WithdrawalTransaction

				BeforeEach(func() {
					mockTransactions := getExampleWithdrawalTransactionsUnsorted()

					expectedTransactions = slices.Clone(mockTransactions)
					domain.SortByTimestampFromNewestToOldest(expectedTransactions)
					Expect(expectedTransactions).NotTo(Equal(mockTransactions))

					mockCall.Return(mockTransactions, nil)
				})

				It("should return sorted array from the newest to the oldest", func() {
					Expect(transactions).To(Equal(expectedTransactions))
				})
			})
		})

		When("balance repository fails", func() {
			var expectedErr error

			BeforeEach(func() {
				expectedErr = errors.New("withrawals transactions are unavailable")

				mockCall.Return(nil, expectedErr)
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(expectedErr))
				Expect(transactions).To(BeEmpty())
			})
		})
	})

	Describe("paying for another order from user's balance", func() {
		var (
			orderID domain.OrderID
			amount  float64
		)

		BeforeEach(func() {
			orderID = exampleOrderID
		})

		JustBeforeEach(func() {
			err = balanceMgr.PayForOrderFromBalance(GinkgoT().Context(), userID, orderID, amount)
		})

		Context("amount is positive", func() {
			var (
				mockCall    *mocks.MockBalanceRepositoryWithdrawFundsCall
				expectedErr error
			)

			BeforeEach(func() {
				amount = 5.23

				mockCall = balanceRepo.EXPECT().
					WithdrawFunds(testutil.MockCtx(), userID, amount)
			})

			When("there are enough funds", func() {
				BeforeEach(func() {
					mockCall.Return(true, nil)
				})

				It("should succeed", func() {
					Expect(err).To(Succeed())
				})
			})

			When("there are not enough funds", func() {
				BeforeEach(func() {
					mockCall.Return(false, nil)
				})

				It("should return ErrBalanceNotEnoughFunds", func() {
					Expect(err).To(MatchError(domain.ErrBalanceNotEnoughFunds))
				})
			})

			When("balance repository fails", func() {
				BeforeEach(func() {
					expectedErr = errors.New("balance not available")

					mockCall.Return(false, expectedErr)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(expectedErr))
				})
			})
		})

		DescribeTableSubtree("amount is not positive",
			func(_amount float64) {
				BeforeEach(func() {
					amount = _amount
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(service.ErrAmountIsNotPositive))
				})
			},
			Entry("amount is negative", -7.77),
			Entry("amount is zero", 0.0),
		)
	})
})

func getExampleWithdrawalTransactionsUnsorted() []domain.WithdrawalTransaction {
	return []domain.WithdrawalTransaction{
		{
			OrderID:     123,
			Amount:      1.23,
			ProcessedAt: time.Date(2025, 11, 5, 10, 11, 12, 0, time.UTC),
		},
		{
			OrderID:     456,
			Amount:      4.56,
			ProcessedAt: time.Date(2025, 11, 7, 9, 10, 11, 0, time.UTC),
		},
		{
			OrderID:     789,
			Amount:      7.89,
			ProcessedAt: time.Date(2025, 11, 15, 8, 9, 10, 0, time.UTC),
		},
		{
			OrderID:     1230,
			Amount:      12.3,
			ProcessedAt: time.Date(2025, 11, 17, 7, 8, 9, 0, time.UTC),
		},
		{
			OrderID:     4560,
			Amount:      45.6,
			ProcessedAt: time.Date(2025, 11, 3, 12, 13, 14, 0, time.UTC),
		},
		{
			OrderID:     7890,
			Amount:      78.9,
			ProcessedAt: time.Date(2025, 11, 1, 17, 16, 15, 0, time.UTC),
		},
	}
}
