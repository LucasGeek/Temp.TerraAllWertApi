package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type UserRepository interface {
	Create(ctx context.Context, user *entities.User) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.User, error)
	GetByEmail(ctx context.Context, email string) (*entities.User, error)
	GetByEnterpriseID(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*entities.User, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.User, error)
	Update(ctx context.Context, user *entities.User) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateLastLogin(ctx context.Context, userID uuid.UUID) error
	GetByRole(ctx context.Context, role entities.UserRole, limit, offset int) ([]*entities.User, error)
	GetActiveUsers(ctx context.Context, limit, offset int) ([]*entities.User, error)
	UpdatePassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
}