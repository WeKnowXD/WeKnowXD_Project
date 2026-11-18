package auth

import (
	"crypto/md5"
	"encoding/hex"
)

func HashPassword(password string) string {
	passwordBytes := []byte(password)
	passwordHash := md5.Sum(passwordBytes)
	passwordHashString := hex.EncodeToString(passwordHash[:])

	return passwordHashString
}

func VerifyPassword(storedHash string, password string) bool {
	passwordHash := HashPassword(password)
	return storedHash == passwordHash
}
