package sqldatabase_test

import (
	"errors"

	"github.com/DATA-DOG/go-sqlmock"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/driver/postgres"
	"gorm.io/gorm/logger"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
)

//nolint:lll
const (
	mockPatternSelectUserByLogin          = `SELECT . FROM "users" WHERE "login" = .+`
	mockPatternSelectUserByID             = `SELECT . FROM "users" WHERE "users"\."id" = .+`
	mockPatternSelectUsersByIDs           = `SELECT . FROM "users" WHERE "users"\."id" IN [(][^)]+[)]`
	mockPatternSelectOrderByID            = `SELECT . FROM "orders" WHERE "id" = .+`
	mockPatternSelectOrdersByUserID       = `SELECT . FROM "orders" WHERE "user_id" = .+ ORDER BY "created_at" DESC`
	mockPatternSelectOrdersByStatuses     = `SELECT . FROM "orders" WHERE "status" IN [(][^)]+[)]`
	mockPatternSelectAccrualByOrderID     = `SELECT . FROM "accruals" WHERE "order_id" = .+`
	mockPatternSelectAccrualsByOrderIDs   = `SELECT . FROM "accruals" WHERE "order_id" IN [(][^)]+[)]`
	mockPatternSelectBalanceByUserID      = `SELECT . FROM "balances" WHERE "user_id" = .+`
	mockPatternSelectWithdrawalsByUserID  = `SELECT . FROM "withdrawals" WHERE "user_id" = .+ ORDER BY "created_at" DESC`
	mockPatternSelectProcessableOrderByID = `SELECT . FROM "processable_orders" WHERE "order_id" = .+`
	mockPatternUpdateOrderStatus          = `UPDATE "orders" SET "status"=.+ WHERE "id" = .+`
	mockPatternUpdateAccrualAmount        = `UPDATE "accruals" SET "amount"=.+ WHERE "order_id" = .+`
	mockPatternUpdateBalance              = `UPDATE "balances" SET "current"="current" [+] .+,"withdrawn"="withdrawn" [+] .+ WHERE "user_id" = .+`
	mockPatternInsertUser                 = `INSERT INTO "users" [(]"login","password_hash"[)] VALUES [(][^)]+[)] RETURNING "id"`
	mockPatternInsertBalance              = `INSERT INTO "balances" [(]"user_id","current","withdrawn"[)] VALUES [(][^)]+[)]`
	mockPatternInsertOrder                = `INSERT INTO "orders" [(]"id","status","created_at","user_id"[)] VALUES [(][^)]+[)]`
	mockPatternInsertWithdrawal           = `INSERT INTO "withdrawals" [(]"amount","created_at","next_order_id","user_id"[)] VALUES [(][^)]+[)] RETURNING "id"`
	mockPatternInsertProcessableOrder     = `INSERT INTO "processable_orders" [(]"order_id","process_after","retries"[)] VALUES [(][^)]+[)]`
	mockPatternDeleteProcessableOrder     = `DELETE FROM "processable_orders" WHERE "order_id" = .+`
	mockPatternForUpdate                  = `\s*FOR UPDATE`
)

var (
	ErrMock = errors.New("mock error")
)

type MockStorage struct {
	*sqldatabase.Storage

	Mock sqlmock.Sqlmock
}

func createMockStorage() *MockStorage {
	GinkgoHelper()

	db, mock, err := sqlmock.New()
	Expect(err).To(Succeed())

	driver := postgres.New(postgres.Config{
		Conn: db,
	})

	storage, err := sqldatabase.NewStorageFromDriver(driver)
	Expect(err).To(Succeed())

	return &MockStorage{
		Storage: storage.WithLogLevel(logger.Silent),
		Mock:    mock,
	}
}
