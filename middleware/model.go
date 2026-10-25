package middleware

import (
	"time"

	"gorm.io/gorm"
)

// Session holds information about user sessions and when they expire
type Session struct {
	gorm.Model
	Identifier string
	UserID     uint
	Role       string
	ExpiresAt  time.Time
}

// Token holds tokens typically used for user activation and password resets
type Token struct {
	gorm.Model
	Value     string
	Type      string
	ModelID   int
	ModelType string
	ExpiresAt time.Time
}

// Website holds information about different websites
type Website struct {
	gorm.Model
	Title       string
	Description string
	URL         string
}

// HasExpired is a helper function that checks if the current time is after the session expire datetime
func (s Session) HasExpired() bool {
	return s.ExpiresAt.Before(time.Now())
}
