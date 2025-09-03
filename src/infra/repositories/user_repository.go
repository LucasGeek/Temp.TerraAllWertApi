package repositories

import (
	"context"

	"terra-allwert/domain/entities"
	"terra-allwert/domain/interfaces"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// UserRepository implements the user repository interface
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) interfaces.UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (r *UserRepository) Create(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

// Delete deletes a user
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&entities.User{}, id).Error
}

// GetByEnterpriseID gets users by enterprise ID
func (r *UserRepository) GetByEnterpriseID(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).Where("enterprise_id = ?", enterpriseID).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// GetAll gets all users with pagination
func (r *UserRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// UpdateLastLogin updates user's last login time
func (r *UserRepository) UpdateLastLogin(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", userID).Update("last_login_at", "NOW()").Error
}

// GetByRole gets users by role
func (r *UserRepository) GetByRole(ctx context.Context, role entities.UserRole, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).Where("role = ?", role).Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// GetActiveUsers gets active users (not deleted)
func (r *UserRepository) GetActiveUsers(ctx context.Context, limit, offset int) ([]*entities.User, error) {
	var users []*entities.User
	err := r.db.WithContext(ctx).Where("deleted_at IS NULL").Limit(limit).Offset(offset).Find(&users).Error
	return users, err
}

// UpdatePassword updates user password
func (r *UserRepository) UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	return r.db.WithContext(ctx).Model(&entities.User{}).Where("id = ?", userID).Update("password_hash", passwordHash).Error
}