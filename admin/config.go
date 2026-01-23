package admin

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/zentooling/golang-web-server/infra"
	"github.com/zentooling/golang-web-server/routes"
)

type ConfigPageData struct {
	routes.PageData
	Config   *infra.Config
	LogLevel string
}

func (svc Service) pageData(c *gin.Context) *ConfigPageData {
	return &ConfigPageData{
		PageData: routes.DefaultPageData(c, svc.env.GetBundle(), svc.env.GetConfig().CacheParameter),
		Config:   svc.env.GetConfig(),
		LogLevel: svc.env.GetLoggingLevel().Level().String(),
	}
}
func (svc Service) ConfigRouteHandler(c *gin.Context) {
	pd := svc.pageData(c)
	pd.Title = pd.Trans("Configuration")

	currentLevel := infra.LairInstance().GetLoggingLevel()
	slog.Error("ConfigRouteHandler", "current logging Level", currentLevel)

	c.HTML(http.StatusOK, "config.gohtml", pd)
}

// LoggingRouteHandlerPost handles change in logging submitted by user
func (svc Service) LoggingRouteHandlerPost(c *gin.Context) {

	levelStr := c.PostForm("log-select")

	level, err := infra.StringToLevel(levelStr)

	// update logging level
	if err == nil {
		svc.env.GetLoggingLevel().Set(level)
	}

	slog.Error("set logging level to " + levelStr)
	slog.Warn("set logging level to " + levelStr)
	slog.Info("set logging level to " + levelStr)
	slog.Debug("set logging level to " + levelStr)

	// read updated state
	pd := svc.pageData(c)
	pd.Title = pd.Trans("Configuration")
	pd.AddMessage(routes.Success, pd.Trans("Logging level set to "+levelStr))

	c.HTML(http.StatusOK, "config.gohtml", pd)
}

func (svc Service) ConfigRouteHandlerPost(c *gin.Context) {

	pd := svc.pageData(c)
	pd.Title = pd.Trans("Configuration")
	// this points to config struct so changes are 'persistent'
	prevCfg := svc.env.GetConfig()

	newValue := c.PostForm("base_url")
	if newValue != prevCfg.BaseURL {
		slog.Info("BaseUrl", "newValue", newValue)
		prevCfg.BaseURL = newValue
		pd.AddMessage(routes.Success, pd.Trans("Base URL changed"))
	}
	newValue = c.PostForm("smtp_host")
	if newValue != prevCfg.SMTPHost {
		slog.Info("SmtpHost", "newValue", newValue)
		prevCfg.SMTPHost = newValue
		pd.AddMessage(routes.Success, pd.Trans("SMTP host changed"))
	}
	newValue = c.PostForm("smtp_port")
	if newValue != prevCfg.SMTPPort {
		slog.Info("SMTPPort", "newValue", newValue)
		prevCfg.SMTPPort = newValue
		pd.AddMessage(routes.Success, pd.Trans("SMTP port changed"))
	}
	newValue = c.PostForm("smtp_sender")
	if newValue != prevCfg.SMTPSender {
		slog.Info("SMTPSender", "newValue", newValue)
		prevCfg.SMTPSender = newValue
		pd.AddMessage(routes.Success, pd.Trans("SMTP sender changed"))
	}
	newValue = c.PostForm("smtp_username")
	if newValue != prevCfg.SMTPUsername {
		slog.Info("SMTPUsername", "newValue", newValue)
		prevCfg.SMTPUsername = newValue
		pd.AddMessage(routes.Success, pd.Trans("SMTP username changed"))
	}
	newValue = c.PostForm("smtp_password")
	if newValue != prevCfg.SMTPPassword {
		slog.Info("SMTPPassword", "newValue", newValue)
		prevCfg.SMTPPassword = newValue
		pd.AddMessage(routes.Success, pd.Trans("SMTP password changed"))
	}
	newValue = c.PostForm("request_per_minute")
	newValueInt, err := strconv.Atoi(newValue)

	if err != nil {
		pd.AddMessage(routes.Error, "Can't convert to integer: "+newValue)
	} else if newValueInt != prevCfg.RequestsPerMinute {
		slog.Info("RequestsPerMinute", "newValue", newValue)
		prevCfg.RequestsPerMinute = newValueInt
		pd.AddMessage(routes.Success, pd.Trans("RequestsPerMinute changed"))
	}
	newValue = c.PostForm("cache_parameter")
	if newValue != prevCfg.CacheParameter {
		slog.Info("CacheParameter", "newValue", newValue)
		prevCfg.CacheParameter = newValue
		pd.AddMessage(routes.Success, pd.Trans("Cache parameter changed"))
	}
	newValue = c.PostForm("cache_max_age")
	// reuse from above
	newValueInt, err = strconv.Atoi(newValue)
	if err != nil {
		pd.AddMessage(routes.Error, "Can't convert to integer: "+newValue)
	} else if newValueInt != prevCfg.CacheMaxAge {
		slog.Info("CacheMaxAge", "newValue", newValue)
		prevCfg.CacheMaxAge = newValueInt
		pd.AddMessage(routes.Success, pd.Trans("Cache max age changed"))
	}

	// refresh with new state
	// pd.Config = infra.LairInstance().GetConfig()

	c.HTML(http.StatusOK, "config.gohtml", pd)
}
