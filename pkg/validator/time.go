package validator

import (
	"time"
)

type TimeValidator struct {
	Field    string
	RawValue string
	Layout   string
	Required bool
	parsed   time.Time
}

func (v *TimeValidator) Validate() error {
	if !v.Required && v.RawValue == "" {
		return nil
	}

	parsed, err := time.Parse(v.Layout, v.RawValue)
	if err != nil {
		return ValidationError{
			Fields: ErrorFields{
				v.Field: "wrong the time format",
			},
		}
	}

	v.parsed = parsed

	return nil
}

func (v *TimeValidator) Value() time.Time {
	return v.parsed
}

func NewTimeValidator(field, value, layout string, required bool) *TimeValidator {
	return &TimeValidator{
		Field:    field,
		RawValue: value,
		Layout:   layout,
		Required: required,
	}
}
