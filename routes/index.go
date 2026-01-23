package routes

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Index renders the HTML of the index page
func (svc Service) Index(c *gin.Context) {
	cfg := svc.env
	pd := DefaultPageData(c, cfg.GetBundle(), cfg.GetConfig().CacheParameter)
	pd.Title = pd.Trans("Home")
	c.HTML(http.StatusOK, "index.gohtml", pd)
}
