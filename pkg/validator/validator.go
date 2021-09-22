package validator

import (
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"
)

func New() (*validator.Validate, error) {
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

	return validate, nil
}
