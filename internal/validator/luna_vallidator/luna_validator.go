package lunavallidator

import "strconv"

type lunaValidator struct {
	lastReport []string
}

func New() *lunaValidator {
	return &lunaValidator{}
}

func (v *lunaValidator) Validate(numberStr []byte) (bool, error) {
	if len(numberStr) == 0 {
		v.lastReport = []string{"empty luna number"}
		return false, nil
	}

	number, err := strconv.Atoi(string(numberStr))
	if err != nil {
		v.lastReport = []string{"invalid luna number"}
		return false, nil
	}

	ok := isValidLuna(number)
	if !ok {
		v.lastReport = []string{"invalid luna number"}
		return false, nil
	}

	v.lastReport = []string{"valid luna number"}
	return true, nil
}

func (v *lunaValidator) Report() []string {
	return v.lastReport
}

// Надеюсь, это работает
func isValidLuna(number int) bool {
	return (number%10+checksum(number/10))%10 == 0
}

func checksum(number int) int {
	var luhn int

	for i := 0; number > 0; i++ {
		cur := number % 10

		if i%2 == 0 {
			cur = cur * 2
			if cur > 9 {
				cur = cur%10 + cur/10
			}
		}

		luhn += cur
		number = number / 10
	}

	return luhn % 10
}
