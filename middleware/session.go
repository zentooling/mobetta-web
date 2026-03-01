package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/models"
	"gorm.io/gorm"
)

// SessionIDKey is the key used to set and get the session id in the context of the current request
const SessionIDKey = "SessionID"

// Session middleware checks for an active session and sets the UserIDKey to the context of the current request if found
func Session(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		session := DefaultSessionWithOptions(c)
		sessionIdentifierInterface := session.Get(SessionIDKey)

		if sessionIdentifier, ok := sessionIdentifierInterface.(string); ok {
			ses := models.Session{
				Identifier: sessionIdentifier,
			}
			res := db.Where(&ses).First(&ses)
			if res.Error == nil && !ses.HasExpired() {
				c.Set(UserIDKey, ses.UserID)
				c.Set(UserRoleKey, ses.Role)
			} else {
				slog.Error("Session", "error", res.Error)
			}
		}
		c.Next()
	}
}

func DefaultSessionWithOptions(c *gin.Context) sessions.Session {
	session := sessions.Default(c)
	// safari strictness requires the SameSite option below
	session.Options(sessions.Options{
		SameSite: http.SameSiteStrictMode,
		MaxAge:   180, // 3 minutes
	})
	return session
}
