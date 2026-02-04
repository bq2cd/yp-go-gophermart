package integration_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/test/spechelpers"
)

var _ = Describe("Sqlite Storage", func() {
	var databaseURI string

	BeforeEach(OncePerOrdered, func() {
		tempDir := spechelpers.CreateLocalTempDir()
		DeferCleanup(tempDir.Cleanup)

		databaseURI = fmt.Sprintf(
			"sqlite:%s?_foreign_keys=1&_journal_mode=WAL&_txlock=immediate",
			tempDir.PathTo("integration.db"),
		)
	})

	describeOrderedAPISpec(&databaseURI)
})
