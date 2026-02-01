package sqldatabase_test

import (
	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/domain"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
)

var _ = Describe("StoreUser", func() {
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

		Context("creating a user", func() {
			var (
				created bool
			)

			JustBeforeEach(func() {
				created, err = storage.CreateUser(
					GinkgoT().Context(),
					domain.UserID(exampleUserLogin),
					domain.PasswordHash(generateRandomBytes()),
				)

				Expect(created).To(BeFalse(), "user should not be created")
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)

			When("database fails when starting a transaction", func() {
				BeforeEach(func() {
					mockQueryUserNotFound(mock)
					mock.ExpectBegin().WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot execute transaction: mock error")))
				})
			})

			When("database fails during creation of the user", func() {
				BeforeEach(func() {
					mockQueryUserNotFound(mock)
					mock.ExpectBegin()
					mock.
						ExpectQuery(mockPatternInsertUser).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot create user: mock error")))
				})
			})

			When("database fails during creation of user's balance", func() {
				BeforeEach(func() {
					mockQueryUserNotFound(mock)
					mock.ExpectBegin()
					mock.
						ExpectQuery(mockPatternInsertUser).
						WillReturnRows(sqlmock.NewRows([]string{}))
					mock.
						ExpectQuery(mockPatternInsertBalance).
						WillReturnError(ErrMock)
				})

				It("should return an error", func() {
					Expect(err).To(MatchError(ContainSubstring("cannot create balance: mock error")))
				})
			})
		})

		Context("getting password hash", func() {
			var (
				hash domain.PasswordHash
			)

			JustBeforeEach(func() {
				hash, err = storage.GetPasswordHash(GinkgoT().Context(), domain.UserID(exampleUserLogin))

				Expect(hash).To(BeNil())
			})

			whenMockDatabaseFailsToFindExistingUser(&mock, &err)
		})
	})
})

/////////////////////////////////////////////////////////////////////////////////

func whenMockDatabaseFailsToFindExistingUser(mockPtr *sqlmock.Sqlmock, errPtr *error) {
	When("database fails during search for existing user", func() {
		BeforeEach(func() {
			(*mockPtr).
				ExpectQuery(mockPatternSelectUserByLogin).
				WillReturnError(ErrMock)
		})

		It("should return an error", func() {
			Expect(*errPtr).To(MatchError(ContainSubstring("cannot search users: mock error")))
		})
	})
}

func whenMockDatabaseUserNotFound(mockPtr *sqlmock.Sqlmock, errPtr *error) {
	When("user does not exist", func() {
		BeforeEach(func() {
			mockQueryUserNotFound(*mockPtr)
		})

		It("should return ErrUserNotFound", func() {
			Expect(*errPtr).To(MatchError(domain.ErrUserNotFound))
		})
	})
}

func mockQueryUserExists(mock sqlmock.Sqlmock) {
	mock.
		ExpectQuery(mockPatternSelectUserByLogin).
		WillReturnRows(
			sqlmock.NewRows([]string{"id", "login"}).
				AddRow(1, exampleUserLogin),
		)
}

func mockQueryUserNotFound(mock sqlmock.Sqlmock) {
	mock.
		ExpectQuery(mockPatternSelectUserByLogin).
		WillReturnRows(sqlmock.NewRows([]string{}))
}
