package login

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/models"
	"github.com/zentooling/golang-web-server/routes"
)

// ResendActivation renders the HTML page used to request a new activation email
func (svc Service) ResendActivation(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Resend Activation Email")
	c.HTML(http.StatusOK, "resendactivation.gohtml", pd)
}

// ResendActivationPost handles the post request for requesting a new activation email
func (svc Service) ResendActivationPost(c *gin.Context) {
	pd := routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	db := svc.env.GetDb()
	pd.Title = pd.Trans("Resend Activation Email")
	email := c.PostForm("email")
	user := models.User{Email: email}
	res := db.Where(&user).First(&user)
	if res.Error == nil && user.ActivatedAt == nil {
		activationToken := models.Token{
			Type:    models.TokenUserActivation,
			ModelID: int(user.ID),
		}

		res = db.Where(&activationToken).First(&activationToken)
		if res.Error == nil {
			// If the activation token exists we simply send an email
			go svc.sendActivationEmail(activationToken.Value, user.Email, pd.Trans)
		} else {
			// If there is no token then we need to generate a new token
			go svc.activationEmailHandler(user.ID, user.Email, pd.Trans)
		}
	} else {
		slog.Error("ResendActivationPost", "error", res.Error)
	}

	// We always return a positive response here to prevent user enumeration and other attacks
	pd.AddMessage(routes.Success, pd.Trans("A new activation email has been sent if the account exists and is not already activated. Please remember to check your spam inbox in case the email is not showing in your inbox."))
	c.HTML(http.StatusOK, "resendactivation.gohtml", pd)
}
