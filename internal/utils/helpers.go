package utils

import (
	"io"

	"github.com/google/uuid"
)

type JSONEncoder interface {
	Encode() ([]byte, error)
}

type JSONDecoder interface {
	Decode(payload []byte) error
	DecodeFrom(r io.Reader) error
}

func IsValidUUID(id string) bool {
	_, err := uuid.Parse(id)
	return err == nil
}
