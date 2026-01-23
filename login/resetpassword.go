package login

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/models"
	"github.com/zentooling/golang-web-server/routes"
	"golang.org/x/crypto/bcrypt"
)

// ResetPasswordPageData defines additional data needed to render the reset password page
type ResetPasswordPageData struct {
	routes.PageData
	Token string
}

// ResetPassword renders the HTML page for resetting the users password
func (svc Service) ResetPassword(c *gin.Context) {
	token := c.Param("token")
	pdPre := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pdPre.Title = pdPre.Trans("Reset Password")
	pd := ResetPasswordPageData{
		PageData: pdPre,
		Token:    token,
	}
	c.HTML(http.StatusOK, "resetpassword.gohtml", pd)
}

// ResetPasswordPost handles post request used to reset users passwords
func (svc Service) ResetPasswordPost(c *gin.Context) {
	pdPre := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	passwordError := pdPre.Trans("Your password must be 8 characters in length or longer")
	resetError := pdPre.Trans("Could not reset password, please try again")

	token := c.Param("token")
	pdPre.Title = pdPre.Trans("Reset Password")
	pd := ResetPasswordPageData{
		PageData: pdPre,
		Token:    token,
	}
	password := c.PostForm("password")

	if len(password) < 8 {
		pd.AddMessage(routes.Error, passwordError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	forgotPasswordToken := models.Token{
		Value: token,
		Type:  models.TokenPasswordReset,
	}

	db := svc.env.GetDb()

	res := db.Where(&forgotPasswordToken).First(&forgotPasswordToken)
	if res.Error != nil {
		pd.AddMessage(routes.Error, resetError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	if forgotPasswordToken.HasExpired() {
		pd.AddMessage(routes.Error, resetError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	user := models.User{}
	user.ID = uint(forgotPasswordToken.ModelID)
	res = db.Where(&user).First(&user)
	if res.Error != nil {
		pd.AddMessage(routes.Error, resetError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)

	if err != nil {
		slog.Error("ResetPasswordPost", "error", err)
		pd.AddMessage(routes.Error, resetError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	user.Password = string(hashedPassword)

	res = db.Save(&user)
	if res.Error != nil {
		pd.AddMessage(routes.Error, resetError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	res = db.Delete(&forgotPasswordToken)
	if res.Error != nil {
		pd.AddMessage(routes.Error, resetError)
		c.HTML(http.StatusBadRequest, "resetpassword.gohtml", pd)
		return
	}

	pd.AddMessage(routes.Success, pd.Trans("Your password has been reset successfully."))

	c.HTML(http.StatusOK, "resetpassword.gohtml", pd)
}
