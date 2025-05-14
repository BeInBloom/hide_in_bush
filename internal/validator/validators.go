package validator

import (
	"errors"
)

var (
	ErrNoValidatorsProvided = errors.New("no validators provided")
)

type ValidatorInput interface {
	~[]byte | ~int | string
}

type Validator[T ValidatorInput] interface {
	Validate(data T) (bool, error)
}

type (
	compositeValidator[T ValidatorInput] struct {
		validators []Validator[T]
		failures   []string
	}

	anyValidator[T ValidatorInput] struct {
		compositeValidator[T]
	}

	allValidator[T ValidatorInput] struct {
		compositeValidator[T]
	}

	NotValidator[T ValidatorInput] struct {
		validator Validator[T]
		failure   []string
	}
)

func NewAny[T ValidatorInput](validators ...Validator[T]) *anyValidator[T] {
	return &anyValidator[T]{
		compositeValidator[T]{
			validators: validators,
		},
	}
}

func NewAll[T ValidatorInput](validators ...Validator[T]) *allValidator[T] {
	return &allValidator[T]{
		compositeValidator[T]{
			validators: validators,
		},
	}
}

func NewNot[T ValidatorInput](v Validator[T]) *NotValidator[T] {
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
	}

	return false, nil
}

func (v *allValidator[T]) Validate(data T) (bool, error) {
	v.failures = nil

	if len(v.validators) == 0 {
		return false, ErrNoValidatorsProvided
	}

	success := true

	for _, validator := range v.validators {
		if ok, err := validator.Validate(data); err != nil {
			return false, err
		} else if !ok {
			return false, nil
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

	return false, nil
}

func (v *NotValidator[T]) Report() []string {
	if v.failure == nil {
		return []string{}
	}

	return v.failure
}
