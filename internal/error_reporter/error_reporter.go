package errorreporter

import "errors"

type formatter interface {
	Format(string) string
}

// Мне просто было интересно, как анврапать ошибки
type errorReporter struct {
	formatter formatter
}

func New(formatter formatter) *errorReporter {
	return &errorReporter{
		formatter: formatter,
	}
}

func (e *errorReporter) Report(err error) []string {
	errStrings := e.getStrings(err)

	if e.formatter != nil {
		return e.formatedErrString(errStrings)
	}

	return errStrings
}

func (e *errorReporter) getStrings(err error) []string {
	if err == nil {
		return []string{}
	}

	var errStrings []string

	errStrings = append(errStrings, err.Error())

	type joinedErrors interface {
		Unwrap() []error
	}

	// Здесь возможна бесконечная рекурсия
	// Если ошибки как-то криво вложатся. Циклические зависимости
	// Нужно будет проверять, анпакал ли я эту ошибку
	// То есть заменить рекурсию на динамическое программирование
	if joined, ok := err.(joinedErrors); ok {
		for _, joinedErr := range joined.Unwrap() {
			if joinedErr != nil {
				joinedStrings := e.getStrings(joinedErr)
				errStrings = append(errStrings, joinedStrings...)
			}
		}
	} else {
		unwrapped := errors.Unwrap(err)
		if unwrapped != nil {
			unwrappedStrings := e.getStrings(unwrapped)
			errStrings = append(errStrings, unwrappedStrings...)
		}
	}

	return errStrings
}

func (e *errorReporter) formatedErrString(errs []string) []string {
	formattedErr := make([]string, len(errs))

	for i, err := range errs {
		formattedErr[i] = e.formatter.Format(err)
	}

	return formattedErr
}
