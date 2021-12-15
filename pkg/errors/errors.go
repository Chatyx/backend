package errors

import (
	"errors"
	"fmt"
)

type ContextFields map[string]interface{}

type ContextError struct {
	Err     error
	Message string
	Fields  ContextFields
}

func WrapInContextError(err error, message string, fields ContextFields) error {
	return ContextError{
		Err:     err,
		Message: message,
		Fields:  fields,
	}
}

func (e ContextError) Unwrap() error {
	return e.Err
}

func (e ContextError) Error() string {
	return fmt.Sprintf("%s: %v", e.Message, e.Err)
}

func UnwrapContextFields(err error) ContextFields {
	fields := ContextFields{}

	for ; err != nil; err = errors.Unwrap(err) {
		ctxErr, ok := err.(ContextError)
		if !ok {
			continue
		}

		for key, value := range ctxErr.Fields {
			fields[key] = value
		}
	}

	return fields
}
