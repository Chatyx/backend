package validator

type chainValidator []Validator

func (v chainValidator) Validate() ErrorFields {
	for _, validator := range v {
		errFields := validator.Validate()
		if len(errFields) != 0 {
			return errFields
		}
	}

	return ErrorFields{}
}

func ChainValidator(validators ...Validator) Validator {
	return chainValidator(validators)
}
