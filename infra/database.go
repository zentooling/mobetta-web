package infra

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/zentooling/golang-web-server/models"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func ConnectToDatabase(c *Config) (*gorm.DB, error) {
	db, err := connectLoop(c, 0)
	// side effect set global for easy access
	return db, err
}

func connectLoop(c *Config, count int) (*gorm.DB, error) {
	db, err := attemptConnection(c)
	if err != nil {
		if count > 300 {
			return db, fmt.Errorf("could not connect to database after 300 seconds")
		}
		time.Sleep(1 * time.Second)
		return connectLoop(c, count+1)
	}
	return db, err
}

func attemptConnection(c *Config) (*gorm.DB, error) {
	var db *gorm.DB
	var err error
	switch c.Database {
	case "sqlite":
		// In-memory sqlite if no database name is specified
		// dsn := "file::memory:?cache=shared"
		dsn := "/tmp/test.db"
		if c.DatabaseName != "" {
			dsn = fmt.Sprintf("%s.db", c.DatabaseName)
		}
		slog.Info("db init", "dsn", dsn)

		db, err = gorm.Open(sqlite.Open(dsn), &gorm.Config{})

	case "mysql":
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", c.DatabaseUsername, c.DatabasePassword, c.DatabaseHost, c.DatabasePort, c.DatabaseName)
		db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	case "postgres":
		dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable TimeZone=UTC", c.DatabaseHost, c.DatabaseUsername, c.DatabasePassword, c.DatabaseName, c.DatabasePort)
		db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	default:
		db, err = nil, fmt.Errorf("no database specified: %s", c.Database)
	}

	return db, err
}

func MigrateDatabase(db *gorm.DB) error {
	err := db.AutoMigrate(&models.User{},
		&models.Role{},
		&models.Token{},
		&models.Session{},
		&models.Website{},
		// user defined tables
		&models.Portfolio{})
	seed(db)
	return err
}

func seed(db *gorm.DB) {

	roles := []models.Role{
		{
			Name:        "admin",
			Description: "Administrative privileges",
		},
		{
			Name:        "user",
			Description: "Regular user privileges",
		},
	}

	for _, role := range roles {
		res := db.Where(&role).First(&role)
		// If no record exists we insert
		if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
			db.Save(&role)
		}
	}

	// create admin account
	adminRole := models.Role{}
	res := db.Where("name='admin'").First(&adminRole)
	if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
		slog.Error("unable to find admin role")
	}
	// make the account active as of now
	tm := time.Now()

	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("password"), bcrypt.DefaultCost)
	userName := "admin@admin.org"
	adminUser := models.User{
		Email:       userName,
		Password:    string(hashedPassword),
		ActivatedAt: &tm,
	}
	adminUser.Roles = append(adminUser.Roles, adminRole)

	// hard delete (w/ Unscoped) previous admin users
	db.Unscoped().Where(&models.User{Email: userName}).Delete(&models.User{})

	// save new one
	db.Save(&adminUser)
	// We seed some websites for our search results
	websites := []models.Website{
		{
			Title:       "A Tour of Go",
			Description: "A Tour of Go has several interactive examples of how Go which you can learn from. There is a menu available if you would like to skip to different sections.",
			URL:         "https://go.dev/tour/welcome/1",
		},
		{
			Title:       "Go by Example",
			Description: "As described on the website: Go by Example is a hands-on introduction to Go using annotated example programs. I have used this site many times as a reference when I need to look something up.",
			URL:         "https://gobyexample.com/",
		},
		{
			Title:       "Go.dev",
			Description: "Learn how to install Go on your machine and read the documentation on the Go website.",
			URL:         "https://go.dev/learn/",
		},
		{
			Title:       "Zen tooling on Github",
			Description: "I am the creator of Golang Base Project. This is my Github profile.",
			URL:         "https://github.com/zentooling",
		},
		{
			Title:       "GORM",
			Description: "The fantastic ORM library for Golang.",
			URL:         "https://gorm.io/",
		},
		{
			Title:       "Bootstrap",
			Description: "Quickly design and customize responsive mobile-first sites with Bootstrap, the world’s most popular front-end open source toolkit, featuring Sass variables and mixins, responsive grid system, extensive prebuilt components, and powerful JavaScript plugins.",
			URL:         "https://getbootstrap.com/",
		},
		{
			Title:       "Gin Web Framework",
			Description: "Gin is a HTTP web framework written in Go (Golang). It features a Martini-like API with much better performance -- up to 40 times faster. If you need smashing performance, get yourself some Gin.",
			URL:         "https://github.com/gin-gonic/gin",
		},
	}

	for _, website := range websites {
		res := db.Where(&website).First(&website)
		// If no record exists we insert
		if res.Error != nil && errors.Is(res.Error, gorm.ErrRecordNotFound) {
			db.Save(&website)
		}
	}
}
