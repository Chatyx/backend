package validator

import "strconv"

type IntValidator struct {
	Field    string
	RawValue string
	Required bool
	parsed   int
}

func (v *IntValidator) Validate() error {
	if !v.Required && v.RawValue == "" {
		return nil
	}

	parsed, err := strconv.Atoi(v.RawValue)
	if err != nil {
		return ValidationError{
			Fields: ErrorFields{
				v.Field: "must be integer",
			},
		}
	}

	v.parsed = parsed

	return nil
}

func (v *IntValidator) Value() int {
	return v.parsed
}

func NewIntValidator(field, value string, required bool) *IntValidator {
	return &IntValidator{
		Field:    field,
		RawValue: value,
		Required: required,
	}
}
