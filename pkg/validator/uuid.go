package validator

import "github.com/google/uuid"

type uuidValidator struct {
	field, value string
}

func isValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}

func (v uuidValidator) Validate() ErrorFields {
	errFields := ErrorFields{}

	if !isValidUUID(v.value) {
		errFields[v.field] = "must be uuid"
	}

	return errFields
}

func UUIDValidator(field, value string) Validator {
	return uuidValidator{
		field: field,
		value: value,
	}
}
