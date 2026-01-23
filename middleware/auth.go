// Package middleware defines all the middlewares for the application
package middleware

import (
	"log/slog"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// UserIDKey is the key used to set and get the user id in the context of the current request
const UserIDKey = "UserID"
const UserRoleKey = "UserRole"

// Auth middleware redirects to /login and aborts the current request if there is no authenticated user
func Auth() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(UserIDKey)
		if !exists {
			slog.Debug("UserIDKey does not exist in context", "keys", c.Keys)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}

	}
}

// Admin middleware requires user to be logged in and have "admin" role assigned
func Admin() gin.HandlerFunc {
	return func(c *gin.Context) {
		_, exists := c.Get(UserIDKey)
		if !exists {
			slog.Debug("UserIDKey does not exist in context", "keys", c.Keys)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}
		role, exists := c.Get(UserRoleKey)
		roleStr := role.(string)
		if !exists {
			slog.Debug("UserRoleKey does not exist in context", "keys", c.Keys)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}
		if !strings.Contains(roleStr, "admin") {
			slog.Debug("User does not have admin privileges", "role", c.Keys)
			c.Redirect(http.StatusTemporaryRedirect, "/login")
			c.Abort()
			return
		}
	}
}
