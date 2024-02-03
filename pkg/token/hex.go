package token

import (
	"crypto/rand"
	"fmt"
)

type Hex struct{}

func (h Hex) Token(size int) (string, error) {
	buf := make([]byte, size)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("fill the buffer random values: %w", err)
	}

	return fmt.Sprintf("%x", buf), nil
}
