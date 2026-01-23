package login

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/infra"
	"github.com/zentooling/golang-web-server/middleware"
	"github.com/zentooling/golang-web-server/models"
	"github.com/zentooling/golang-web-server/routes"
	"github.com/zentooling/golang-web-server/ulid"
	"golang.org/x/crypto/bcrypt"
)

type Service struct {
	env infra.ILair
}

func NewService(env infra.ILair) *Service {
	return &Service{env: env}
}

// Login renders the HTML of the login page
func (svc Service) Login(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Login")
	c.HTML(http.StatusOK, "login.gohtml", pd)
}

// LoginPost handles login requests and returns the appropriate HTML and messages
func (svc Service) LoginPost(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Login")

	loginError := pd.Trans("Could not login, please make sure that you have typed in the correct email and password. If you have forgotten your password, please click the forgot password link below.")

	db := svc.env.GetDb()

	email := c.PostForm("email")

	// load user and associated roles

	//res := db.Model(&models.User{}).Preload("Roles").Find(&user)

	user := models.User{Email: email}
	res := db.Preload("Roles").Where(&user).First(&user)

	//res := db.Where(&user).First(&user)
	if res.Error != nil {
		pd.AddMessage(routes.Error, loginError)
		slog.Error("LoginPost", "error", res.Error)
		c.HTML(http.StatusInternalServerError, "login.gohtml", pd)
		return
	}

	if res.RowsAffected == 0 {
		pd.AddMessage(routes.Error, loginError)
		c.HTML(http.StatusBadRequest, "login.gohtml", pd)
		return
	}

	if user.ActivatedAt == nil {
		pd.AddMessage(routes.Error, pd.Trans("Account is not activated yet."))
		c.HTML(http.StatusBadRequest, "login.gohtml", pd)
		return
	}

	if len(user.Roles) == 0 {
		pd.AddMessage(routes.Error, pd.Trans("Account does not contain role attributes."))
		c.HTML(http.StatusBadRequest, "login.gohtml", pd)
		return

	}

	password := c.PostForm("password")
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		pd.AddMessage(routes.Error, loginError)
		c.HTML(http.StatusBadRequest, "login.gohtml", pd)
		return
	}

	// Generate a ulid for the current session
	sessionIdentifier := ulid.Generate()

	ses := models.Session{
		Identifier: sessionIdentifier,
	}

	// Session is valid for 1 hour
	ses.ExpiresAt = time.Now().Add(time.Hour)
	ses.UserID = user.ID
	var roles []string
	for _, role := range user.Roles {
		roles = append(roles, role.Name)
	}
	ses.Role = strings.Join(roles, ",") // comma seperated list of roles

	slog.Debug("LoginPost", "session", ses)

	res = db.Save(&ses)
	if res.Error != nil {
		pd.AddMessage(routes.Error, loginError)
		slog.Error("LoginPost", "error", res.Error)
		c.HTML(http.StatusInternalServerError, "login.gohtml", pd)
		return
	}

	session := middleware.DefaultSessionWithOptions(c)
	session.Set(middleware.SessionIDKey, sessionIdentifier)
	// Safari strictness requires the following

	err = session.Save()
	if err != nil {
		pd.AddMessage(routes.Error, loginError)
		slog.Error("LoginPost", "error", err)
		c.HTML(http.StatusInternalServerError, "login.gohtml", pd)
		return
	}

	//c.Redirect(http.StatusTemporaryRedirect, "/admin")
	c.Redirect(http.StatusMovedPermanently, "/")
}
