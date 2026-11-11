package models

import "time"

type SessionData struct {
	UserID    int
	Username  string
	ExpiresAt time.Time
}

type User struct {
	Id       int
	Username string
	Email    string
	Password string
}
