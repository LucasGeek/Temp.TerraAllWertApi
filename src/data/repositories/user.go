package repositories

import (
	"context"

	"gorm.io/gorm"
)

type UserRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{
		db: db,
	}
}

func (r *UserRepository) HasUsers(ctx context.Context) (bool, error) {
	// Placeholder implementation
	// var count int64
	// if err := r.db.WithContext(ctx).Model(&User{}).Count(&count).Error; err != nil {
	//     return false, err
	// }
	// return count > 0, nil
	return false, nil
}