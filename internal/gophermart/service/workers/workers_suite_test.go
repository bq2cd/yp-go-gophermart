package workers_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/internal/test/testutil"
)

func TestWorkers(t *testing.T) {
	testutil.MaybeEnableDebugLogging()

	RegisterFailHandler(Fail)
	RunSpecs(t, "Workers Suite")
}
