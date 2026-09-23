package portfolio

import (
	"context"

	"github.com/zentooling/golang-web-server/models"
	"gorm.io/gorm"
)

// UserRepository defines the behavior for managing user data
type HoldingsRepository interface {
	//Create(ctx context.Context, user *models.Portfolio) error
	//GetByID(ctx context.Context, id uint) (*models.Portfolio, error)
	GetAll(ctx context.Context) (*[]models.Portfolio, error)
	//Update(ctx context.Context, user *models.Portfolio) error
	//Delete(ctx context.Context, id uint) error
}

// gormHoldingsRepository implements the UserRepository interface
type gormHoldingsRepository struct {
	db *gorm.DB
}

// NewUserRepository is a constructor function that returns the interface
func NewHoldingsRepository(db *gorm.DB) HoldingsRepository {
	return &gormHoldingsRepository{db: db}
}

// // Create inserts a new user into the database
//
//	func (r *gormUserRepository) Create(ctx context.Context, user *User) error {
//		return r.db.WithContext(ctx).Create(user).Error
//	}
//
// GetByID finds a single user by their primary key
func (r *gormHoldingsRepository) GetAll(ctx context.Context) (*[]models.Portfolio, error) {
	var holdings []models.Portfolio
	// 1. Create a slice to hold all the users

	// 2. Find all records and save them into the slice
	result := r.db.Find(&holdings)

	// This runs the SQL equivalent of: SELECT * FROM users;

	if result.Error != nil {
		return nil, result.Error
	}
	return &holdings, nil
}

//
//// GetByID finds a single user by their primary key
//func (r *gormUserRepository) GetByID(ctx context.Context, id uint) (*User, error) {
//	var user User
//	err := r.db.WithContext(ctx).First(&user, id).Error
//	if err != nil {
//		return nil, err
//	}
//	return &user, nil
//}
//
//// Update saves changes made to an existing user
//func (r *gormUserRepository) Update(ctx context.Context, user *User) error {
//	return r.db.WithContext(ctx).Save(user).Error
//}
//
//// Delete marks a user record as deleted (soft delete)
//func (r *gormUserRepository) Delete(ctx context.Context, id uint) error {
//	return r.db.WithContext(ctx).Delete(&User{}, id).Error
//}
//
