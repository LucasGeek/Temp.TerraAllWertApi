package interfaces

import (
	"context"

	"api/domain/entities"
)

type FloorRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, floor *entities.Floor) error
	GetByID(ctx context.Context, id string) (*entities.Floor, error)
	GetByTowerID(ctx context.Context, towerID string) ([]*entities.Floor, error)
	Update(ctx context.Context, floor *entities.Floor) error
	Delete(ctx context.Context, id string) error
	
	// Business operations
	GetWithApartments(ctx context.Context, id string) (*entities.Floor, error)
	GetTotalApartments(ctx context.Context, id string) (int, error)
	ExistsByTowerAndNumber(ctx context.Context, towerID, number string) (bool, error)
}