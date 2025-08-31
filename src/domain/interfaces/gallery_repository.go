package interfaces

import (
	"context"

	"api/domain/entities"
)

type GalleryRepository interface {
	Create(ctx context.Context, image *entities.GalleryImage) error
	GetByID(ctx context.Context, id string) (*entities.GalleryImage, error)
	GetByRoute(ctx context.Context, route string) ([]*entities.GalleryImage, error)
	GetAllRoutes(ctx context.Context) ([]string, error)
	Update(ctx context.Context, image *entities.GalleryImage) error
	Delete(ctx context.Context, id string) error
	ReorderImages(ctx context.Context, route string, imageIDs []string) error
}

type ImagePinRepository interface {
	Create(ctx context.Context, pin *entities.ImagePin) error
	GetByID(ctx context.Context, id string) (*entities.ImagePin, error)
	GetByGalleryImageID(ctx context.Context, galleryImageID string) ([]*entities.ImagePin, error)
	Update(ctx context.Context, pin *entities.ImagePin) error
	Delete(ctx context.Context, id string) error
}

type ApartmentImageRepository interface {
	Create(ctx context.Context, image *entities.ApartmentImage) error
	GetByID(ctx context.Context, id string) (*entities.ApartmentImage, error)
	GetByApartmentID(ctx context.Context, apartmentID string) ([]*entities.ApartmentImage, error)
	Update(ctx context.Context, image *entities.ApartmentImage) error
	Delete(ctx context.Context, id string) error
	ReorderImages(ctx context.Context, apartmentID string, imageIDs []string) error
}