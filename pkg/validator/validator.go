package validator

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

type Validator struct {
	validate *validator.Validate
}

func NewValidator() Validator {
	validate := validator.New()
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		tagValue := field.Tag.Get("json")
		if tagValue == "" {
			return field.Name
		}

		fieldName := strings.SplitN(tagValue, ",", 2)[0]
		if fieldName == "-" {
			return field.Name
		}

		return fieldName
	})

	return Validator{validate: validate}
}

func (v Validator) Struct(val any) error {
	if err := v.validate.Struct(val); err != nil {
		vErrs := validator.ValidationErrors{}

		if errors.As(err, &vErrs) {
			fields := make(ErrorFields, len(vErrs))
			for _, vErr := range vErrs {
				fullFieldName := strings.Join(
					strings.Split(vErr.Namespace(), ".")[1:],
					".")
				fields[fullFieldName] = fmt.Sprintf("failed on the '%s' tag", vErr.Tag())
			}

			return Error{Fields: fields}
		}

		return fmt.Errorf("validate struct: %w", err)
	}

	return nil
}

func (v Validator) Var(val any, key, tag string) error {
	if err := v.validate.Var(val, tag); err != nil {
		vErrs := validator.ValidationErrors{}

		if errors.As(err, &vErrs) {
			vErr := vErrs[0]
			return Error{
				Fields: ErrorFields{
					key: fmt.Sprintf("failed on the '%s' tag", vErr.Tag()),
				},
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
