package integration_test

import (
	. "github.com/onsi/ginkgo/v2"

	"github.com/bq2cd/yp-go-gophermart/internal/gophermart/repository/inmemory"
	"github.com/bq2cd/yp-go-gophermart/internal/test/integration"
)

var _ = Describe("In-Memory Storage", func() {
	describeOrderedAPISpec(func() integration.Storage {
		return inmemory.NewStorage()
	})
})
