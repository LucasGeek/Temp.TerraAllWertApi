package interfaces

import (
	"context"

	"api/domain/entities"
)

type TowerRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, tower *entities.Tower) error
	GetByID(ctx context.Context, id string) (*entities.Tower, error)
	GetAll(ctx context.Context) ([]*entities.Tower, error)
	Update(ctx context.Context, tower *entities.Tower) error
	Delete(ctx context.Context, id string) error
	
	// Business operations
	GetWithFloors(ctx context.Context, id string) (*entities.Tower, error)
	GetTotalApartments(ctx context.Context, id string) (int, error)
	ExistsByName(ctx context.Context, name string) (bool, error)
}