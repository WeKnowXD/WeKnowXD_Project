package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"

	"github.com/WeKnowXD/WeKnowXD_Project/internal/models"
)

var (
	Sessions      map[string]models.SessionData
	SessionsMutex sync.RWMutex
)

func Init() {
	Sessions = make(map[string]models.SessionData)
}

func GenerateSessionToken() (string, error) {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
