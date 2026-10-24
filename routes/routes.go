// Package routes defines all the handling functions for all the routes
package routes

import (
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	// "github.com/zentooling/golang-web-server/database"

	"github.com/zentooling/golang-web-server/middleware"
)

// PageData holds the default data needed for HTML pages to render
type PageData struct {
	Title           string
	Messages        []Message
	IsAuthenticated bool
	IsAdminRole     bool
	CacheParameter  string
	Trans           func(s string) string
}

// Define an enum using iota
type MessageType int

const (
	Error MessageType = iota
	Warning
	Info
	Success
)

// Map for string representation
var messageTypeNames = map[MessageType]string{
	Error:   "error",
	Warning: "warning",
	Info:    "info",
	Success: "success",
}

func (pd *PageData) AddMessage(msgType MessageType, content string) {
	pd.Messages = append(pd.Messages, Message{
		Type:    msgType,
		Content: content,
	})
}

func (c MessageType) String() string {
	if name, ok := messageTypeNames[c]; ok {
		return name
	}
	return "Unknown"
}

// Message holds a message which can be rendered as responses on HTML pages in messages.gohtml
type Message struct {
	Type    MessageType // success, warning, error, etc.
	Content string
}

// isAuthenticated checks if the current user is authenticated or not
func isAuthenticated(c *gin.Context) bool {
	_, exists := c.Get(middleware.UserIDKey)
	return exists
}

func isAdminRole(c *gin.Context) bool {
	role, ret := c.Get(middleware.UserRoleKey)
	if !ret { // does not exist
		return false
	}
	// ret is true
	if role.(string) != "admin" {
		ret = false
	}

	return ret
}

func getUserId(c *gin.Context) uint {
	id, exists := c.Get(middleware.UserIDKey)
	if exists {
		return id.(uint)
	}
	return 0
}

func DefaultPageData(c *gin.Context) PageData {
	// TODO make this a singleton
	bundle := LoadLanguageBundles()
	langService := NewLangService(c, bundle)
	return PageData{
		Title:           "Home",
		Messages:        nil,
		IsAuthenticated: isAuthenticated(c),
		IsAdminRole:     isAdminRole(c),
		CacheParameter:  "jtFkffdefsdfcMVx", // TODO make this dynamic?
		Trans:           langService.Trans,
	}
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{
		db: db,
	}
}
