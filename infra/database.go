package infra

import (
	"fmt"
	"log/slog"
	"time"

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
