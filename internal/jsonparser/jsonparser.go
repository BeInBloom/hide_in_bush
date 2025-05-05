package jsonparser

import (
	"encoding/json"
	"errors"
	"fmt"
)

type validator[T any] interface {
	Validate(data T) (bool, error)
	Report() []string
}

type parser[T any] struct {
	validators []validator[T]
}

func New[T any](validators ...validator[T]) *parser[T] {
	return &parser[T]{
		validators: validators,
	}
}

func (p *parser[T]) Parse(data []byte) (T, error) {
	var result T

	if err := json.Unmarshal(data, &result); err != nil {
		return result, fmt.Errorf("failed to unmarshal JSON: %w", err)
	}

	var errs []error
	for _, validator := range p.validators {
		ok, err := validator.Validate(result)
		if err != nil {
			errs = append(errs, err)
		} else if !ok {
			errs = append(errs, fmt.Errorf("%s", validator.Report()))
		}
	}

	if len(errs) > 0 {
		return result, fmt.Errorf("validation errors: %w", errors.Join(errs...))
	}

	return result, nil
}
