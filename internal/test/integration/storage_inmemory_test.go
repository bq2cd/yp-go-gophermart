package integration_test

import (
	. "github.com/onsi/ginkgo/v2"
)

var _ = Describe("In-Memory Storage", func() {
	describeOrderedAPISpec("")
})
