package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type SuiteSearchFilters struct {
	MinBedrooms   *int
	MaxBedrooms   *int
	MinArea       *float64
	MaxArea       *float64
	MinPrice      *float64
	MaxPrice      *float64
	Status        *entities.SuiteStatus
	SunPosition   *entities.SunPosition
	FloorID       *uuid.UUID
	TowerID       *uuid.UUID
	MinSuites     *int
	MaxSuites     *int
	MinBathrooms  *int
	MaxBathrooms  *int
	ParkingSpaces *int
}

type SuiteRepository interface {
	Create(ctx context.Context, suite *entities.Suite) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Suite, error)
	GetByFloorID(ctx context.Context, floorID uuid.UUID, limit, offset int) ([]*entities.Suite, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Suite, error)
	Update(ctx context.Context, suite *entities.Suite) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdateStatus(ctx context.Context, suiteID uuid.UUID, status entities.SuiteStatus) error
	GetByStatus(ctx context.Context, status entities.SuiteStatus, limit, offset int) ([]*entities.Suite, error)
	Search(ctx context.Context, filters SuiteSearchFilters, limit, offset int) ([]*entities.Suite, error)
	GetByTowerID(ctx context.Context, towerID uuid.UUID, limit, offset int) ([]*entities.Suite, error)
	GetAvailableSuites(ctx context.Context, limit, offset int) ([]*entities.Suite, error)
	GetSuitesByPriceRange(ctx context.Context, minPrice, maxPrice float64, limit, offset int) ([]*entities.Suite, error)
}