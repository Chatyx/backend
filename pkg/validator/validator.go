package validator

import "github.com/go-playground/validator/v10"

type ErrorFields map[string]string

type Validator interface {
	Validate() ErrorFields
}

var validate *validator.Validate

func SetValidate(v *validator.Validate) {
	validate = v
}
