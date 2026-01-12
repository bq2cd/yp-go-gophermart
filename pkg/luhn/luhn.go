package luhn

const (
	doubleCoefficient = 2
	overflowValue     = 9
	numberSystemBase  = 10
)

// Validate calculates the checksum of a provided number using Luhn's algorithm.
// It returns an error if the resulting checksum does not match the number.
func Validate(number uint64) error {
	expectedCheckDigit, actualCheckDigit := calculateCheckDigit(number)

	if actualCheckDigit != expectedCheckDigit {
		return ErrCheckDigitMismatch
	}

	return nil
}

// numberToDigitsReversed returns "reversed" slice of digits, e.g.
// for '123' it will return [3, 2, 1].
// This is exactly what we need for luhn's algorithm.
func numberToDigitsReversed(number uint64) []uint64 {
	digits := make([]uint64, 0)

	for number > 0 {
		n, r := number/numberSystemBase, number%numberSystemBase
		digits = append(digits, r)
		number = n
	}

	return digits
}

func calculateCheckDigit(number uint64) (uint64, uint64) {
	var checkSum uint64

	digits := numberToDigitsReversed(number)

	expectedCheckDigit := digits[0]

	for i := 1; i < len(digits); i++ {
		checkSum += maybeDoubleDigit(digits[i], i+1)
	}

	actualCheckDigit := (numberSystemBase - checkSum%numberSystemBase) % numberSystemBase

	return expectedCheckDigit, actualCheckDigit
}

func maybeDoubleDigit(digit uint64, position int) uint64 {
	if position%2 != 0 {
		return digit
	}

	return doubleDigit(digit)
}

func doubleDigit(digit uint64) uint64 {
	result := digit * doubleCoefficient
	if result > overflowValue {
		result -= overflowValue
	}

	return result
}
