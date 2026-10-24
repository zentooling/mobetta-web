package config

import (
	"log/slog"
	"os"
	"strconv"

	"github.com/gorilla/securecookie"
	"github.com/zentooling/golang-web-server/text"
)

var GlobalConfig *Config

func Initialize() *Config {
	var c Config
	c.LogLevel = "DEBUG"
	if os.Getenv("LOG_LEVEL") != "" {
		c.LogLevel = os.Getenv("LOG_LEVEL")
	}
	c.Port = "8080"
	if os.Getenv("PORT") != "" {
		c.Port = os.Getenv("PORT")
	}

	c.BaseURL = "https://golangbase.com/"
	if os.Getenv("BASE_URL") != "" {
		c.BaseURL = os.Getenv("BASE_URL")
	}

	// A random secret will be generated when the application starts if no secret is provided. It is highly recommended providing a secret.
	c.CookieSecret = string(securecookie.GenerateRandomKey(64))
	if os.Getenv("COOKIE_SECRET") != "" {
		c.CookieSecret = os.Getenv("COOKIE_SECRET")
	}

	c.Database = "sqlite"
	if os.Getenv("DATABASE") != "" {
		c.Database = os.Getenv("DATABASE")
	}
	if os.Getenv("DATABASE_NAME") != "" {
		c.DatabaseName = os.Getenv("DATABASE_NAME")
	}
	if os.Getenv("DATABASE_HOST") != "" {
		c.DatabaseHost = os.Getenv("DATABASE_HOST")
	}
	if os.Getenv("DATABASE_PORT") != "" {
		c.DatabasePort = os.Getenv("DATABASE_PORT")
	}
	if os.Getenv("DATABASE_USERNAME") != "" {
		c.DatabaseUsername = os.Getenv("DATABASE_USERNAME")
	}
	if os.Getenv("DATABASE_PASSWORD") != "" {
		c.DatabasePassword = os.Getenv("DATABASE_PASSWORD")
	}

	if os.Getenv("SMTP_USERNAME") != "" {
		c.SMTPUsername = os.Getenv("SMTP_USERNAME")
	}
	if os.Getenv("SMTP_SENDER") != "" {
		c.SMTPSender = os.Getenv("SMTP_SENDER")
	}
	if os.Getenv("SMTP_PASSWORD") != "" {
		c.SMTPPassword = os.Getenv("SMTP_PASSWORD")
	}
	if os.Getenv("SMTP_HOST") != "" {
		c.SMTPHost = os.Getenv("SMTP_HOST")
	}
	if os.Getenv("SMTP_PORT") != "" {
		c.SMTPPort = os.Getenv("SMTP_PORT")
	}
	if os.Getenv("PORTFOLIO_XLS_LOC") != "" {
		c.PortfolioXlsLoc = os.Getenv("PORTFOLIO_XLS_LOC")
	}

	c.RequestsPerMinute = 5
	if os.Getenv("REQUESTS_PER_MINUTE") != "" {
		i, err := strconv.Atoi(os.Getenv("REQUESTS_PER_MINUTE"))
		if err != nil {
			slog.Warn("Env:REQUESTS_PER_MINUTE", "error", err)
		}
		c.RequestsPerMinute = i
	}

	// CacheParameter is added to the end of static file urls to prevent caching old versions
	c.CacheParameter = text.RandomString(10)
	if os.Getenv("CACHE_PARAMETER") != "" {
		c.CacheParameter = os.Getenv("CACHE_PARAMETER")
	}

	// CacheMaxAge is how many seconds to cache static assets, 1 year by default
	c.CacheMaxAge = 31536000
	if os.Getenv("CACHE_MAX_AGE") != "" {
		i, err := strconv.Atoi(os.Getenv("CACHE_MAX_AGE"))
		if err != nil {
			slog.Warn("Env:CACHE_MAX_AGE", "error", err)
		}
		c.CacheMaxAge = i
	}

	GlobalConfig = &c
	return &c
}
