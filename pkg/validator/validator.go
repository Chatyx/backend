package validator

import (
	"errors"
	"fmt"

	"github.com/go-playground/validator/v10"
)

type Validate struct {
	validate *validator.Validate
}

func NewValidate(v *validator.Validate) Validate {
	return Validate{validate: v}
}

func (v Validate) Struct(val any) error {
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

func (v Validate) Var(val any, tag string) error {
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
