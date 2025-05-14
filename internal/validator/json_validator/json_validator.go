package jsonvalidator

import "github.com/xeipuuv/gojsonschema"

type Validator struct {
	schema       gojsonschema.JSONLoader
	latestResult *gojsonschema.Result
}

func New(schema string) *Validator {
	return &Validator{
		schema: gojsonschema.NewStringLoader(schema),
	}
}

func (v *Validator) Validate(data []byte) (bool, error) {
	document := gojsonschema.NewBytesLoader(data)

	result, err := gojsonschema.Validate(v.schema, document)
	if err != nil {
		return false, err
	}

	v.latestResult = result

	return result.Valid(), nil
}

func (v *Validator) Report() []string {
	if v.latestResult == nil {
		//хз что лучше вернуть. Пустую структуру или nil
		return []string{}
	}

	errors := make([]string, len(v.latestResult.Errors()))
	for i, err := range v.latestResult.Errors() {
		errors[i] = err.String()
	}

	return errors
}
