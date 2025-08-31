package interfaces

import (
	"context"

	"api/domain/entities"
)

type ApartmentRepository interface {
	// Basic CRUD operations
	Create(ctx context.Context, apartment *entities.Apartment) error
	GetByID(ctx context.Context, id string) (*entities.Apartment, error)
	GetByFloorID(ctx context.Context, floorID string) ([]*entities.Apartment, error)
	Update(ctx context.Context, apartment *entities.Apartment) error
	Delete(ctx context.Context, id string) error
	
	// Business operations
	GetWithImages(ctx context.Context, id string) (*entities.Apartment, error)
	Search(ctx context.Context, criteria *entities.ApartmentSearchCriteria) ([]*entities.Apartment, error)
	ExistsByFloorAndNumber(ctx context.Context, floorID, number string) (bool, error)
	GetAvailableCount(ctx context.Context) (int, error)
	GetByStatus(ctx context.Context, status entities.ApartmentStatus) ([]*entities.Apartment, error)
}