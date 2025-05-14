package lunavallidator

import "strconv"

type lunaValidator struct{}

func New() *lunaValidator {
	return &lunaValidator{}
}

func (v *lunaValidator) Validate(numberStr []byte) (bool, error) {
	if len(numberStr) == 0 {
		return false, nil
	}

	number, err := strconv.Atoi(string(numberStr))
	if err != nil {
		return false, nil
	}

	ok := isValidLuna(number)
	if !ok {
		return false, nil
	}

	return true, nil
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
