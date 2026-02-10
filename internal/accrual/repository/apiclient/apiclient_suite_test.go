package apiclient_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestApiclient(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Apiclient Suite")
}
