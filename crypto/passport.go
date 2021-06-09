package crypto

import (
	"crypto/rand"
	"encoding/hex"
	"io"

	"golang.org/x/crypto/scrypt"
)

const (
	SALT_PASS_DEFAULT string = "yqs04UUMewGQZd0g"
)

// GenerateSalt generates a random salt
func GenerateSalt() string {
	saltBytes := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, saltBytes); err != nil {
		return SALT_PASS_DEFAULT
	}
	salt := make([]byte, 32)
	hex.Encode(salt, saltBytes)
	return string(salt)
}

// HashPassword hashes a string
func HashPassword(password, salt string) (string, error) {
	hashedPasswordBytes, err := scrypt.Key([]byte(password), []byte(salt), 16384, 8, 1, 32)
	if err != nil {
		return password, err
	}
	hashedPassword := make([]byte, 64)
	hex.Encode(hashedPassword, hashedPasswordBytes)
	return string(hashedPassword), nil
}

// VerifyPassword verifies if password matches the users password
func VerifyPassword(passHash, passPlain, salt string) bool {
	hashedPassword, err := HashPassword(passPlain, salt)
	if err != nil {
		return false
	}
	return passHash == hashedPassword
}
