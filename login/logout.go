package login

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/middleware"
)

// Logout deletes the current user session and redirects the user to the index page
func (svc Service) Logout(c *gin.Context) {
	session := middleware.DefaultSessionWithOptions(c)

	session.Delete(middleware.SessionIDKey)
	err := session.Save()
	if err != nil {
		slog.Error("Logout", "error", err)
	}

	c.Redirect(http.StatusTemporaryRedirect, "/")
}
