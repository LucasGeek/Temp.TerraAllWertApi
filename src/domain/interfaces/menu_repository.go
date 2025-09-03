package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type MenuRepository interface {
	Create(ctx context.Context, menu *entities.Menu) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Menu, error)
	GetByEnterpriseID(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*entities.Menu, error)
	GetBySlug(ctx context.Context, enterpriseID uuid.UUID, slug string) (*entities.Menu, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Menu, error)
	Update(ctx context.Context, menu *entities.Menu) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetChildren(ctx context.Context, parentID uuid.UUID, limit, offset int) ([]*entities.Menu, error)
	GetRootMenus(ctx context.Context, enterpriseID uuid.UUID, limit, offset int) ([]*entities.Menu, error)
	GetByScreenType(ctx context.Context, enterpriseID uuid.UUID, screenType entities.ScreenType, limit, offset int) ([]*entities.Menu, error)
	UpdatePosition(ctx context.Context, menuID uuid.UUID, position int) error
	GetMenuHierarchy(ctx context.Context, enterpriseID uuid.UUID) ([]*entities.Menu, error)
}

type TowerRepository interface {
	Create(ctx context.Context, tower *entities.Tower) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Tower, error)
	GetByMenuFloorPlanID(ctx context.Context, menuFloorPlanID uuid.UUID, limit, offset int) ([]*entities.Tower, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Tower, error)
	Update(ctx context.Context, tower *entities.Tower) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePosition(ctx context.Context, towerID uuid.UUID, position int) error
}

type FloorRepository interface {
	Create(ctx context.Context, floor *entities.Floor) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.Floor, error)
	GetByTowerID(ctx context.Context, towerID uuid.UUID, limit, offset int) ([]*entities.Floor, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Floor, error)
	Update(ctx context.Context, floor *entities.Floor) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByFloorNumber(ctx context.Context, towerID uuid.UUID, floorNumber int) (*entities.Floor, error)
}