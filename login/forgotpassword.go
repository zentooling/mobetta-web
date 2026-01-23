package login

import (
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/gin-gonic/gin"
	email2 "github.com/zentooling/golang-web-server/email"
	"github.com/zentooling/golang-web-server/models"
	"github.com/zentooling/golang-web-server/routes"
	"github.com/zentooling/golang-web-server/ulid"
	"gorm.io/gorm"
)

// ForgotPassword renders the HTML page where a password request can be initiated
func (svc Service) ForgotPassword(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Forgot Password")
	c.HTML(http.StatusOK, "forgotpassword.gohtml", pd)
}

// ForgotPasswordPost handles the POST request which requests a password reset and then renders the HTML page with the appropriate message
func (svc Service) ForgotPasswordPost(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Forgot Password")

	db := svc.env.GetDb()

	email := c.PostForm("email")
	user := models.User{Email: email}
	res := db.Where(&user).First(&user)
	if res.Error == nil && user.ActivatedAt != nil {
		go svc.forgotPasswordEmailHandler(user.ID, email, pd.Trans)
	}

	pd.AddMessage(routes.Success, pd.Trans("An email with instructions to reset password has been sent"))
	// We always return a positive response here to prevent user enumeration
	c.HTML(http.StatusOK, "forgotpassword.gohtml", pd)
}

func (svc Service) forgotPasswordEmailHandler(userID uint, email string, trans func(string) string) {
	forgotPasswordToken := models.Token{
		Value: ulid.Generate(),
		Type:  models.TokenPasswordReset,
	}

	db := svc.env.GetDb()

	res := db.Where(&forgotPasswordToken).First(&forgotPasswordToken)
	if (res.Error != nil && res.Error != gorm.ErrRecordNotFound) || res.RowsAffected > 0 {
		// If the forgot password token already exists we try to generate it again
		svc.forgotPasswordEmailHandler(userID, email, trans)
		return
	}

	forgotPasswordToken.ModelID = int(userID)
	forgotPasswordToken.ModelType = "User"
	// The token will expire 10 minutes after it was created
	forgotPasswordToken.ExpiresAt = time.Now().Add(time.Minute * 10)

	res = db.Save(&forgotPasswordToken)
	if res.Error != nil || res.RowsAffected == 0 {
		slog.Error("sendForgetPasswordEmail", "error", res.Error)
		return
	}
	svc.sendForgotPasswordEmail(forgotPasswordToken.Value, email, trans)
}

func (svc Service) sendForgotPasswordEmail(token string, email string, trans func(string) string) {
	conf := svc.env.GetConfig()
	u, err := url.Parse(conf.BaseURL)
	if err != nil {
		slog.Error("sendForgetPasswordEmail", "error", err)
		return
	}

	u.Path = path.Join(u.Path, "/user/password/reset/", token)

	resetPasswordURL := u.String()

	emailService := email2.New(conf)

	emailService.Send(email, trans("Password Reset"), fmt.Sprintf(trans("Use the following link to reset your password. If this was not requested by you, please ignore this email.\n%s"), resetPasswordURL))
}
