package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator(v *validator.Validate) Validator {
	return Validator{validate: v}
}

func (v Validator) Struct(val any) error {
	if err := v.validate.Struct(val); err != nil {
		vErrs := validator.ValidationErrors{}

		if errors.As(err, &vErrs) {
			fields := make(ErrorFields, len(vErrs))
			for _, vErr := range vErrs {
				fields[vErr.Field()] = vErr.Error()
			}

			return Error{Fields: fields}
		}

		return fmt.Errorf("validate struct: %w", err)
	}

	return nil
}

func (v Validator) Var(val any, tag string) error {
	if err := v.validate.Var(val, tag); err != nil {
		vErrs := validator.ValidationErrors{}

		if errors.As(err, &vErrs) {
			vErr := vErrs[0]
			return Error{
				Fields: ErrorFields{vErr.Field(): vErr.Error()},
			}
		}

		return fmt.Errorf("validate variable: %w", err)
	}

	return nil
}

func MergeResults(errs ...error) error {
	fields := make(ErrorFields)

	for _, err := range errs {
		if err == nil {
			continue
		}

		ve := Error{}
		if !errors.As(err, &ve) {
			return err
		}

		for k, v := range ve.Fields {
			fields[k] = v
		}
	}

	if len(fields) != 0 {
		return Error{Fields: fields}
	}
	return nil
}
