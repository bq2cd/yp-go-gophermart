package integration_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

func TestIntegration(t *testing.T) {
	testutil.MaybeEnableDebugLogging()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}
