package login

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	email2 "github.com/zentooling/golang-web-server/email"
	"github.com/zentooling/golang-web-server/models"
	"github.com/zentooling/golang-web-server/routes"
	"github.com/zentooling/golang-web-server/ulid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// Register renders the HTML content of the register page
func (svc Service) Register(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Register")
	c.HTML(http.StatusOK, "register.gohtml", pd)
}

// RegisterPost handles requests to register users and returns appropriate messages as HTML content
func (svc Service) RegisterPost(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	passwordError := pd.Trans("Your password must be 8 characters in length or longer")
	registerError := pd.Trans("Could not register, please make sure the details you have provided are correct and that you do not already have an existing account.")
	registerSuccess := pd.Trans("Thank you for registering. An activation email has been sent with steps describing how to activate your account.")
	pd.Title = pd.Trans("Register")
	password := c.PostForm("password")
	if len(password) < 8 {
		pd.AddMessage(routes.Error, passwordError)
		c.HTML(http.StatusBadRequest, "register.gohtml", pd)
		return
	}

	// The password is hashed as early as possible to make timing attacks that reveal registered users harder
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		pd.AddMessage(routes.Error, registerError)
		slog.Error("RegisterPost:GenerateFromPassword", "error", err)
		c.HTML(http.StatusInternalServerError, "register.gohtml", pd)
		return
	}

	email := c.PostForm("email")

	// Validate the email
	validate := validator.New()
	err = validate.Var(email, "required,email")

	if err != nil {
		pd.AddMessage(routes.Error, registerError)
		slog.Error("RegisterPost:Validate", "error", err)
		c.HTML(http.StatusInternalServerError, "register.gohtml", pd)
		return
	}

	user := models.User{Email: email}
	role := models.Role{}

	db := svc.env.GetDb()

	// retrieve the 'user' role
	res := db.Where("name='user'").First(&role)
	if (res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound)) || res.RowsAffected > 1 {
		pd.AddMessage(routes.Error, registerError)
		slog.Error("RegisterPost", "error", res.Error)
		c.HTML(http.StatusInternalServerError, "register.gohtml", pd)
		return
	}

	user.Roles = append(user.Roles, role)

	res = db.Where(&user).First(&user)
	if (res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound)) || res.RowsAffected > 0 {
		pd.AddMessage(routes.Error, registerError)
		slog.Error("RegisterPost", "error", res.Error)
		c.HTML(http.StatusInternalServerError, "register.gohtml", pd)
		return
	}

	user.Password = string(hashedPassword)

	res = db.Save(&user)
	if res.Error != nil || res.RowsAffected == 0 {
		pd.AddMessage(routes.Error, registerError)
		slog.Error("Register:SaveUser", "error", res.Error)
		c.HTML(http.StatusInternalServerError, "register.gohtml", pd)
		return
	}

	// Generate activation token and send activation email
	go svc.activationEmailHandler(user.ID, email, pd.Trans)

	pd.AddMessage(routes.Success, registerSuccess)

	c.HTML(http.StatusOK, "register.gohtml", pd)
}

func (svc Service) activationEmailHandler(userID uint, email string, trans func(string) string) {
	activationToken := models.Token{
		Value: ulid.Generate(),
		Type:  models.TokenUserActivation,
	}

	db := svc.env.GetDb()

	res := db.Where(&activationToken).First(&activationToken)
	if (res.Error != nil && res.Error != gorm.ErrRecordNotFound) || res.RowsAffected > 0 {
		// If the activation token already exists we try to generate it again
		svc.activationEmailHandler(userID, email, trans)
		return
	}

	activationToken.ModelID = int(userID)
	activationToken.ModelType = "User"
	activationToken.ExpiresAt = time.Now().Add(time.Minute * 10)

	res = db.Save(&activationToken)
	if res.Error != nil || res.RowsAffected == 0 {
		slog.Error("activationEmailHandler:Save", "error", res.Error)
		return
	}
	svc.sendActivationEmail(activationToken.Value, email, trans)
}

func (svc Service) sendActivationEmail(token string, email string, trans func(string) string) {
	cfg := svc.env.GetConfig()
	u, err := url.Parse(cfg.BaseURL)
	if err != nil {
		slog.Error("activationEmailHandler:sendActivationEmail", "error", err)
		return
	}

	u.Path = path.Join(u.Path, "/activate/", token)

	activationURL := u.String()

	emailService := email2.New(cfg)

	emailService.Send(email, trans("User Activation"), fmt.Sprintf(trans("Use the following link to activate your account. If this was not requested by you, please ignore this email.\n%s"), activationURL))
}
