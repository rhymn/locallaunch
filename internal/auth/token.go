package auth

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
)

const tokenLength = 16

func Generate() (string, error) {
	b := make([]byte, tokenLength)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generating token: %w", err)
	}
	return hex.EncodeToString(b), nil
}

func Validate(token, expected string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(expected)) == 1
}
