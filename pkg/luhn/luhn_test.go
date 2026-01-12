package luhn_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/bq2cd/yp-go-gophermart/pkg/luhn"
)

var _ = Describe("Luhn", func() {

	testCases(
		"valid numbers",
		func(number int) {
			expectResult(number, true)
		},
		"should return nil",
		17893729974,
		49927398716,
		79927398713,
		4539148803436467,
		6011111111111117,
		378282246310005,
		4111111111111111,
		5555555555554444,
		371449635398431,
		6011000990139424,
		4532900000000000,
		4222222222222220,
	)

	testCases(
		"invalid numbers",
		func(number int) {
			expectResult(number, false)
		},
		"should return error",
		123,
		4539148803436468,
		6011111111111110,
		1234567890123456,
		4111111111111112,
		5555555555554445,
		371449635398432,
		6011000990139423,
		4532900000000006,
		4532876512345678,
	)
})

func testCases(tableDescription string, validatorFn func(int), entryDescription string, numbers ...int) {
	args := make([]any, 0, len(numbers))

	args = append(args, validatorFn)
	args = append(args, func(num int) string {
		return fmt.Sprintf("%d %s", num, entryDescription)
	})

	for _, number := range numbers {
		args = append(args, Entry(nil, number))
	}

	DescribeTable(tableDescription, args...)
}

func expectResult(number int, expectValid bool) {
	err := luhn.Validate(uint64(number))

	if expectValid {
		Expect(err).NotTo(HaveOccurred())
	} else {
		Expect(err).To(HaveOccurred())
	}
}
