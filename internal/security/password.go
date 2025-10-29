package security

import (
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

func ComparePasswords(hashEncoded string, password string, salt string) error {
	hash, err := base64.StdEncoding.DecodeString(hashEncoded)
	if err != nil {
		return err
	}

	return bcrypt.CompareHashAndPassword(hash, []byte(password+salt))
}

func HashPassword(password string, salt string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password+salt), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return base64.StdEncoding.EncodeToString(hash), nil
}
