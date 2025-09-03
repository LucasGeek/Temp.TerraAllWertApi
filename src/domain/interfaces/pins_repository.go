package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type MenuPinsRepository interface {
	Create(ctx context.Context, pins *entities.MenuPins) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.MenuPins, error)
	GetByMenuID(ctx context.Context, menuID uuid.UUID) (*entities.MenuPins, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.MenuPins, error)
	Update(ctx context.Context, pins *entities.MenuPins) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type PinMarkerRepository interface {
	Create(ctx context.Context, marker *entities.PinMarker) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.PinMarker, error)
	GetByMenuPinID(ctx context.Context, menuPinID uuid.UUID, limit, offset int) ([]*entities.PinMarker, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.PinMarker, error)
	Update(ctx context.Context, marker *entities.PinMarker) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetVisibleMarkers(ctx context.Context, menuPinID uuid.UUID, limit, offset int) ([]*entities.PinMarker, error)
	GetByActionType(ctx context.Context, menuPinID uuid.UUID, actionType entities.PinAction, limit, offset int) ([]*entities.PinMarker, error)
	GetByPosition(ctx context.Context, menuPinID uuid.UUID, minX, maxX, minY, maxY float64, limit, offset int) ([]*entities.PinMarker, error)
}

type PinMarkerImageRepository interface {
	Create(ctx context.Context, image *entities.PinMarkerImage) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.PinMarkerImage, error)
	GetByPinMarkerID(ctx context.Context, pinMarkerID uuid.UUID, limit, offset int) ([]*entities.PinMarkerImage, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.PinMarkerImage, error)
	Update(ctx context.Context, image *entities.PinMarkerImage) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePosition(ctx context.Context, imageID uuid.UUID, position int) error
}