package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/test/spechelpers"
)

var _ = Describe("Postgres Storage", func() {
	var (
		databaseURI string
		err         error
	)

	BeforeEach(OncePerOrdered, func() {
		container := spechelpers.StartPostgresContainer(GinkgoT().Context())
		DeferCleanup(container.Terminate, GinkgoT().Context())

		databaseURI, err = container.ConnectionString(GinkgoT().Context())
		Expect(err).To(Succeed())
	})

	describeOrderedAPISpec(&databaseURI)
})
