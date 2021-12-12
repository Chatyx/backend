package validator

import "github.com/go-playground/validator/v10"

type Validator interface {
	Validate() error
}

var validate *validator.Validate

func SetValidate(v *validator.Validate) {
	validate = v
}
