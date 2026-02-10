package sqldatabase_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service"
	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/service/workers"
)

// Ensure [sqldatabase.Storage] implements all necessary interfaces.
var (
	_ service.UserRepository    = (*sqldatabase.Storage)(nil)
	_ service.BalanceRepository = (*sqldatabase.Storage)(nil)
	_ service.OrderRepository   = (*sqldatabase.Storage)(nil)
	_ workers.OrderRepository   = (*sqldatabase.Storage)(nil)
	_ workers.OrderQueue        = (*sqldatabase.Storage)(nil)
)

/////////////////////////////////////////////////////////////////////////////////

var _ = Describe("Storage", func() {
	var storage *sqldatabase.Storage

	Context("creating new storage from URI", func() {
		var (
			databaseURI string
			err         error
		)

		JustBeforeEach(func() {
			storage, err = sqldatabase.NewStorage(databaseURI)
		})

		When("database url contains unsupported driver", func() {
			BeforeEach(func() {
				databaseURI = "mysql://host:3306"
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(sqldatabase.ErrUnsupportedDatabaseDriver))
				Expect(storage).To(BeNil())
			})
		})

		When("database url is invalid", func() {
			BeforeEach(func() {
				databaseURI = ":::"
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("cannot parse database uri")))
				Expect(storage).To(BeNil())
			})
		})

		When("database url cannot be opened", func() {
			BeforeEach(func() {
				databaseURI = "sqlite://non-existent-file"
			})

			It("should return an error", func() {
				Expect(err).To(MatchError(ContainSubstring("cannot open database")))
				Expect(storage).To(BeNil())
			})
		})
	})
})
