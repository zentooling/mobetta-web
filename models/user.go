// Package models defines all the database models for the application
package models

import (
	"time"

	"gorm.io/gorm"
)

// User holds information relating to users that use the application
type User struct {
	gorm.Model
	Email       string
	Password    string
	ActivatedAt *time.Time
	Roles       []Role  `gorm:"many2many:user_roles;"` // Many-to-many relationship with Role
	Tokens      []Token `gorm:"polymorphic:Model;"`
	Sessions    []Session
}

// Role represents a user role (user,admin,etc)
type Role struct {
	gorm.Model
	Name        string `gorm:"uniqueIndex;not null"`
	Description string
	Users       []User `gorm:"many2many:user_roles;"` // Many-to-many relationship with User
}
