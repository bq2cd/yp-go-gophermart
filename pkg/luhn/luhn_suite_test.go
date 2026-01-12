package luhn_test

import (
	"testing"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestLuhn(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Luhn Suite")
}
