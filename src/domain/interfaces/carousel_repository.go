package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type MenuCarouselRepository interface {
	Create(ctx context.Context, carousel *entities.MenuCarousel) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.MenuCarousel, error)
	GetByMenuID(ctx context.Context, menuID uuid.UUID) (*entities.MenuCarousel, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.MenuCarousel, error)
	Update(ctx context.Context, carousel *entities.MenuCarousel) error
	Delete(ctx context.Context, id uuid.UUID) error
}

type CarouselItemRepository interface {
	Create(ctx context.Context, item *entities.CarouselItem) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.CarouselItem, error)
	GetByMenuCarouselID(ctx context.Context, menuCarouselID uuid.UUID, limit, offset int) ([]*entities.CarouselItem, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.CarouselItem, error)
	Update(ctx context.Context, item *entities.CarouselItem) error
	Delete(ctx context.Context, id uuid.UUID) error
	UpdatePosition(ctx context.Context, itemID uuid.UUID, position int) error
	GetActiveItems(ctx context.Context, menuCarouselID uuid.UUID, limit, offset int) ([]*entities.CarouselItem, error)
	GetByItemType(ctx context.Context, menuCarouselID uuid.UUID, itemType entities.CarouselItemType, limit, offset int) ([]*entities.CarouselItem, error)
}

type CarouselTextOverlayRepository interface {
	Create(ctx context.Context, overlay *entities.CarouselTextOverlay) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.CarouselTextOverlay, error)
	GetByCarouselItemID(ctx context.Context, carouselItemID uuid.UUID, limit, offset int) ([]*entities.CarouselTextOverlay, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.CarouselTextOverlay, error)
	Update(ctx context.Context, overlay *entities.CarouselTextOverlay) error
	Delete(ctx context.Context, id uuid.UUID) error
}