package validator

import (
	"reflect"
	"strings"
	"time"

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

	err := validate.RegisterValidation("sql-date", func(fl validator.FieldLevel) bool {
		val := fl.Field().String()
		if val == "" {
			return true
		}

		_, err := time.Parse("2006-01-02", fl.Field().String())
		return err == nil
	})
	if err != nil {
		return nil, err
	}

	return validate, nil
}
