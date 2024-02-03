package hasher

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type BCrypt struct{}

func (h BCrypt) Hash(s string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(s), bcrypt.MinCost)
	if err != nil {
		return "", fmt.Errorf("generate hash from password: %v", err)
	}

	return string(hash), nil
}

func (h BCrypt) CompareHashAndPassword(hash, password string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	return err == nil
}
