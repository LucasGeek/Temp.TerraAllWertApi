package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type EnterpriseRepository interface {
	Create(ctx context.Context, enterprise *entities.Enterprise) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Enterprise, error)
	GetBySlug(ctx context.Context, slug string) (*entities.Enterprise, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Enterprise, error)
	Update(ctx context.Context, enterprise *entities.Enterprise) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByCity(ctx context.Context, city string, limit, offset int) ([]*entities.Enterprise, error)
	GetByStatus(ctx context.Context, status entities.EnterpriseStatus, limit, offset int) ([]*entities.Enterprise, error)
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Enterprise, error)
}