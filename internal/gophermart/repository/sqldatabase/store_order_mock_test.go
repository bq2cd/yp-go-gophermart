package sqldatabase_test

import (
	"database/sql/driver"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
)

var _ = Describe("StoreOrder", func() {
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

		Context("creating an order", func() {
			var (
				created bool
				ownerID domain.UserID
				orderID uint
			)

			BeforeEach(func() {
				orderID = exampleOrderID
			})

			JustBeforeEach(func() {
				created, ownerID, err = storage.CreateOrder(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
					domain.OrderID(orderID),
				)

				Expect(created).To(BeFalse(), "order should not be created")
				Expect(ownerID).To(Equal(domain.UserIDEmptyValue))
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			whenMockDatabaseUserNotFound(&mock, &err)

			whenMockDatabaseFailsToFindExistingOrder(&mock, &err)

			When("database fails during creation of the order", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mockQueryOrderNotFound(mock)
					mock.ExpectBegin()
					mock.
						ExpectExec(mockPatternInsertOrder).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot create order: mock error")))
				})
			})

			When("database fails during creation of the processable order", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mockQueryOrderNotFound(mock)
					mock.ExpectBegin()
					mock.
						ExpectExec(mockPatternInsertOrder).
						WillReturnResult(driver.RowsAffected(1))
					mock.
						ExpectExec(mockPatternInsertProcessableOrder).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot create processable order: mock error")))
				})
			})
		})

		Context("listing orders", func() {
			var (
				result []domain.Order
				err    error
			)

			JustBeforeEach(func() {
				result, err = storage.GetOrders(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
				)

				Expect(result).To(BeEmpty())
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			When("database fails during listing of the orders", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mock.
						ExpectQuery(mockPatternSelectOrdersByUserID).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot list orders: mock error")))
				})
			})
		})

		Context("listing accruals", func() {
			var (
				result map[domain.OrderID]float64
				err    error
			)

			JustBeforeEach(func() {
				result, err = storage.GetOrderAccruals(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
					[]domain.OrderID{123, 456, 789},
				)

				Expect(result).To(BeEmpty())
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			When("database fails during search for accruals", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mock.
						ExpectQuery(mockPatternSelectAccrualsByOrderIDs).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot search accruals: mock error")))
				})
			})
		})

		Context("getting order status", func() {
			var (
				orderID uint
				result  domain.OrderStatus
				err     error
			)

			BeforeEach(func() {
				orderID = exampleOrderID
			})

			JustBeforeEach(func() {
				result, err = storage.GetOrderStatus(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
					domain.OrderID(orderID),
				)

				Expect(result).To(Equal(domain.OrderStatusInvalid))
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			whenMockDatabaseUserNotFound(&mock, &err)

			whenMockDatabaseFailsToFindExistingOrder(&mock, &err)
		})

		Context("setting order status", func() {
			var (
				orderID uint
				err     error
			)

			BeforeEach(func() {
				orderID = exampleOrderID
			})

			JustBeforeEach(func() {
				err = storage.MarkOrderProcessed(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
					domain.OrderID(orderID),
					5.25,
				)
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			whenMockDatabaseUserNotFound(&mock, &err)

			whenMockDatabaseFailsToFindExistingOrder(&mock, &err)

			Context("inside transaction", func() {
				BeforeEach(func() {
					mockQueryUserExists(mock)
					mockQueryOrderExists(mock, exampleUserLogin, orderID)
					mock.ExpectBegin()
				})

				When("database fails during setting order status", func() {
					BeforeEach(func() {
						mock.
							ExpectExec(mockPatternUpdateOrderStatus).
							WillReturnError(ErrMock)
					})

					It("should return an error", func() {
						Expect(err).To(MatchError(ContainSubstring("cannot update order status: mock error")))
					})
				})

				When("database fails during setting accrual points", func() {
					BeforeEach(func() {
						mock.
							ExpectExec(mockPatternUpdateOrderStatus).
							WillReturnResult(driver.RowsAffected(1))
					})

					When("database fails during search for existing accrual points", func() {
						BeforeEach(func() {
							mock.
								ExpectQuery(mockPatternSelectAccrualByOrderID + mockPatternForUpdate).
								WillReturnError(ErrMock)
						})

						It("should return an error", func() {
							Expect(err).To(MatchError(ContainSubstring("cannot search accruals: mock error")))
						})
					})

					When("database fails during updating accrual points", func() {
						BeforeEach(func() {
							mock.
								ExpectQuery(mockPatternSelectAccrualByOrderID + mockPatternForUpdate).
								WillReturnRows(sqlmock.NewRows([]string{"order_id", "amount"}).
									AddRow(orderID, 1.0),
								)
							mock.
								ExpectExec(mockPatternUpdateAccrualAmount).
								WillReturnError(ErrMock)
						})

						It("should return an error", func() {
							Expect(err).To(MatchError(ContainSubstring("cannot update accrual: mock error")))
						})
					})
				})

				When("database fails during updating user's balance", func() {
					BeforeEach(func() {
						mock.
							ExpectExec(mockPatternUpdateOrderStatus).
							WillReturnResult(driver.RowsAffected(1))
						mock.
							ExpectQuery(mockPatternSelectAccrualByOrderID + mockPatternForUpdate).
							WillReturnRows(sqlmock.NewRows([]string{"order_id", "amount"}).
								AddRow(orderID, 1.0),
							)
						mock.
							ExpectExec(mockPatternUpdateAccrualAmount).
							WillReturnResult(driver.RowsAffected(1))
					})

					When("database fails during search for user's balance", func() {
						BeforeEach(func() {
							mock.
								ExpectQuery(mockPatternSelectBalanceByUserID + mockPatternForUpdate).
								WillReturnError(ErrMock)
						})

						It("should return an error", func() {
							Expect(err).To(MatchError(ContainSubstring("cannot search balances: mock error")))
						})
					})

					When("database fails during updating user's balance", func() {
						BeforeEach(func() {
							mock.
								ExpectQuery(mockPatternSelectBalanceByUserID + mockPatternForUpdate).
								WillReturnRows(sqlmock.NewRows([]string{"id", "current", "user_id"}).
									AddRow(1, 2.0, 1),
								)
							mock.
								ExpectExec(mockPatternUpdateBalance).
								WillReturnError(ErrMock)
						})

						It("should return an error", func() {
							Expect(err).To(MatchError(ContainSubstring("cannot update balance: mock error")))
						})
					})
				})

				When("database fails during removing processable order", func() {
					BeforeEach(func() {
						mock.
							ExpectExec(mockPatternUpdateOrderStatus).
							WillReturnResult(driver.RowsAffected(1))
						mock.
							ExpectQuery(mockPatternSelectAccrualByOrderID + mockPatternForUpdate).
							WillReturnRows(sqlmock.NewRows([]string{"order_id", "amount"}).
								AddRow(orderID, 1.0),
							)
						mock.
							ExpectExec(mockPatternUpdateAccrualAmount).
							WillReturnResult(driver.RowsAffected(1))
						mock.
							ExpectQuery(mockPatternSelectBalanceByUserID + mockPatternForUpdate).
							WillReturnRows(sqlmock.NewRows([]string{"id", "current", "user_id"}).
								AddRow(1, 2.0, 1),
							)
						mock.
							ExpectExec(mockPatternUpdateBalance).
							WillReturnResult(driver.RowsAffected(1))
						mock.
							ExpectQuery(mockPatternSelectBalanceByUserID).
							WillReturnRows(sqlmock.NewRows([]string{"id", "current", "user_id"}).
								AddRow(1, 3.0, 1),
							)
					})

					When("database fails during search for a processable order", func() {
						BeforeEach(func() {
							mock.
								ExpectQuery(mockPatternSelectProcessableOrderByID + mockPatternForUpdate).
								WillReturnError(ErrMock)
						})

						It("should return an error", func() {
							Expect(err).To(MatchError(ContainSubstring("cannot search processable orders: mock error")))
						})
					})

					When("database fails during removing of a processable order", func() {
						BeforeEach(func() {
							mock.
								ExpectQuery(mockPatternSelectProcessableOrderByID + mockPatternForUpdate).
								WillReturnRows(sqlmock.NewRows([]string{"order_id", "process_after", "retries", "user_id"}).
									AddRow(1, time.Now(), 5, 1),
								)
							mock.
								ExpectExec(mockPatternDeleteProcessableOrder).
								WillReturnError(ErrMock)
						})

						It("should return an error", func() {
							Expect(err).To(MatchError(ContainSubstring("cannot remove processable order: mock error")))
						})
					})
				})
			})
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func whenMockDatabaseFailsToFindExistingOrder(mockPtr *sqlmock.Sqlmock, errPtr *error) {
	When("database fails during search for existing order", func() {
		BeforeEach(func() {
			mockQueryUserExists(*mockPtr)
			(*mockPtr).
				ExpectQuery(mockPatternSelectOrderByID).
				WillReturnError(ErrMock)
		})

		It("should return an error", func() {
			Expect(*errPtr).To(MatchError(ContainSubstring("cannot query orders: mock error")))
		})
	})
}

func mockQueryOrderExists(mock sqlmock.Sqlmock, login string, orderID uint) {
	mock.
		ExpectQuery(mockPatternSelectOrderByID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "status", "user_id"}).
				AddRow(orderID, domain.OrderStatusNew.Int(), 1),
		)
	mock.
		ExpectQuery(mockPatternSelectUserByID).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login"}).
				AddRow(1, login),
		)
}

func mockQueryOrderNotFound(mock sqlmock.Sqlmock) {
	mock.
		ExpectQuery(mockPatternSelectOrderByID).
		WillReturnRows(
			sqlmock.NewRows([]string{}),
		)
}
