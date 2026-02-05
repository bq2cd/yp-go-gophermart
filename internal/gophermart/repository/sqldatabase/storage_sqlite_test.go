package sqldatabase_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"gorm.io/gorm/logger"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/sqldatabase"
	"github.com/bq2cd/yp-go-gophermart/internal/test/spechelpers"
)

type TempStorage struct {
	*sqldatabase.Storage

	Cleanup func()
}

func createTempStorage() *TempStorage {
	GinkgoHelper()

	tempDir := spechelpers.CreateLocalTempDir()

	storage, err := sqldatabase.NewStorage("sqlite:" + tempDir.PathTo("gophermart.db") + "?_foreign_keys=on")
	Expect(err).To(Succeed())

	err = storage.AutoMigrate()
	Expect(err).To(Succeed())

	return &TempStorage{
		Storage: storage.WithLogLevel(logger.Silent),
		Cleanup: tempDir.Cleanup,
	}
}

/////////////////////////////////////////////////////////////////////////////////
