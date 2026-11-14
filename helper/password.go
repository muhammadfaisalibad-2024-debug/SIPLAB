package helper

import (
	"golang.org/x/crypto/bcrypt"
)

// HashPassword mengubah password polos menjadi hash bcrypt dengan DefaultCost.
func HashPassword(plain string) (string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashed), nil
}

// CheckPassword membandingkan password polos dengan hash bcrypt secara konstan.
func CheckPassword(hashed, plain string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hashed), []byte(plain)) == nil
}
