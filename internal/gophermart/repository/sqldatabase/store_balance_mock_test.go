package sqldatabase_test

import (
	"database/sql/driver"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
)

var _ = Describe("StoreBalance", func() {
	var storage *sqldatabase.Storage

	Context("mock database", func() {
		var (
			mock sqlmock.Sqlmock
			err  error
		)

		BeforeEach(func() {
			mockStorage := createMockStorage()

			storage = mockStorage.Storage
			mock = mockStorage.Mock
		})

		DescribeTableSubtree("getting user's balance",
			func(_actionFn func() (any, error)) {
				JustBeforeEach(func() {
					_, err = _actionFn()
				})

				whenMockDatabaseFailsToFindExistingUser(&mock, &err)

				When("database fails during search for user's balance", func() {
					BeforeEach(func() {
						mockQueryUserExists(mock)
						mock.
							ExpectQuery(mockPatternSelectBalanceByUserID).
							WillReturnError(ErrMock)
					})

					It("should return an error", func() {
						Expect(err).To(MatchError(ContainSubstring("cannot search balances: mock error")))
					})
				})
			},
			Entry("current value", func() (any, error) {
				return storage.GetCurrentValue(GinkgoT().Context(), domain.UserID(exampleUserLogin))
			}),
			Entry("withdrawn value", func() (any, error) {
				return storage.GetTotalAmountWithdrawn(GinkgoT().Context(), domain.UserID(exampleUserLogin))
			}),
		)

		Context("withdrawing funds", func() {
			var success bool

			JustBeforeEach(func() {
				success, err = storage.WithdrawFunds(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
					domain.OrderID(exampleOrderID),
					9.99,
				)

				Expect(success).To(BeFalse())
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			When("database fails to update user's balance", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mock.ExpectBegin()
					mockQueryBalanceExists(mock, 10.0, 0.0)
					mock.
						ExpectExec(mockPatternUpdateBalance).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot update balance: mock error")))
				})
			})

			When("database fails to get updated user's balance", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mock.ExpectBegin()
					mockQueryBalanceExists(mock, 10.0, 0.0)
					mock.
						ExpectExec(mockPatternUpdateBalance).
						WillReturnResult(driver.RowsAffected(1))
					mock.
						ExpectQuery(mockPatternSelectBalanceByUserID).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot search balances: mock error")))
				})
			})

			When("database fails to add withdrawal transaction", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mock.ExpectBegin()
					mockQueryBalanceExists(mock, 10.0, 0.0)
					mock.
						ExpectExec(mockPatternUpdateBalance).
						WillReturnResult(driver.RowsAffected(1))
					mockQueryBalanceExists(mock, 0.01, 9.99)
					mock.
						ExpectQuery(mockPatternInsertWithdrawal).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot add withdrawal transaction: mock error")))
				})
			})
		})

		Context("listing withdrawals", func() {
			var result []domain.WithdrawalTransaction

			JustBeforeEach(func() {
				result, err = storage.GetWithdrawalTransactions(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
				)

				Expect(result).To(BeEmpty())
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			When("database fail during search for withdrawal transactions", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mock.
						ExpectQuery(mockPatternSelectWithdrawalsByUserID).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot search withdrawals: mock error")))
				})
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func mockQueryBalanceExists(mock sqlmock.Sqlmock, current, withdrawn float64) {
	mock.
		ExpectQuery(mockPatternSelectBalanceByUserID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "current", "withdrawn", "user_id"}).
				AddRow(1, current, withdrawn, 1),
		)
}
