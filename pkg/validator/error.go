package validator

import (
	"strings"
)

type ErrorFields map[string]string

func (ef ErrorFields) String() string {
	b := strings.Builder{}
	b.WriteByte('[')

	i := 0
	for field, value := range ef {
		if i > 0 {
			b.WriteString(", ")
		}

		i++
		b.WriteString(field + ":" + value)
	}

	b.WriteByte(']')
	return b.String()
}

type Error struct {
	Fields ErrorFields
}

func (e Error) Error() string {
	return e.Fields.String()
}
