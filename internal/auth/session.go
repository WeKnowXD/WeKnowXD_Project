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

func GenerateSessionToken() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}
