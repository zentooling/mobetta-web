package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// NoRoute handles rendering of the 404 page
func (svc Service) NoRoute(c *gin.Context) {
	pd := DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter)
	pd.Title = pd.Trans("404 Not Found")
	c.HTML(http.StatusOK, "404.gohtml", pd)
}
