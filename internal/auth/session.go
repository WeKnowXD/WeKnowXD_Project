package auth

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

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

// ActiveUsers returns how many distinct logged-in users were seen within window.
func ActiveUsers(window time.Duration) int {
	now := time.Now()
	cutoff := now.Add(-window)

	SessionsMutex.RLock()
	defer SessionsMutex.RUnlock()

	seen := make(map[int]bool)
	for _, s := range Sessions {
		if now.After(s.ExpiresAt) || s.LastSeen.Before(cutoff) {
			continue
		}
		seen[s.UserID] = true
	}
	return len(seen)
}
