package validator

import (
	"errors"
)

var (
	ErrNoValidatorsProvided = errors.New("no validators provided")
)

type validator[T any] interface {
	Validate(data T) (bool, error)
	Report() []string
}

type (
	compositeValidator[T any] struct {
		validators []validator[T]
		failures   []string
	}

	anyValidator[T any] struct {
		compositeValidator[T]
	}

	allValidator[T any] struct {
		compositeValidator[T]
	}

	NotValidator[T any] struct {
		validator validator[T]
		failure   []string
	}
)

func NewAny[T any](validators ...validator[T]) *anyValidator[T] {
	return &anyValidator[T]{
		compositeValidator[T]{
			validators: validators,
		},
	}
}

func NewAll[T any](validators ...validator[T]) *allValidator[T] {
	return &allValidator[T]{
		compositeValidator[T]{
			validators: validators,
		},
	}
}

func NewNot[T any](v validator[T]) *NotValidator[T] {
	return &NotValidator[T]{
		validator: v,
	}
}

func (v *anyValidator[T]) Validate(data T) (bool, error) {
	v.failures = nil

	if len(v.validators) == 0 {
		return false, ErrNoValidatorsProvided
	}

	for _, validator := range v.validators {
		if ok, _ := validator.Validate(data); ok {
			return true, nil
		}
		v.failures = append(v.failures, validator.Report()...)
	}

	return false, nil
}

func (v *anyValidator[T]) Report() []string {
	if len(v.failures) == 0 {
		return []string{}
	}

	return v.failures
}

func (v *allValidator[T]) Validate(data T) (bool, error) {
	v.failures = nil

	if len(v.validators) == 0 {
		return false, ErrNoValidatorsProvided
	}

	success := true

	for _, validator := range v.validators {
		if ok, err := validator.Validate(data); !ok || err != nil {
			success = false
			if err != nil {
				v.failures = append(v.failures, err.Error())
			} else {
				v.failures = append(v.failures, validator.Report()...)
			}
		}
	}

	return success, nil
}

func (v *allValidator[T]) Report() []string {
	if v.failures == nil {
		return []string{}
	}

	return v.failures
}

func (v *NotValidator[T]) Validate(data T) (bool, error) {
	v.failure = nil

	ok, err := v.validator.Validate(data)
	if err != nil {
		return false, err
	}

	if !ok {
		return true, nil
	}

	v.failure = append(v.failure, v.validator.Report()...)
	return false, nil
}

func (v *NotValidator[T]) Report() []string {
	if v.failure == nil {
		return []string{}
	}

	return v.failure
}
