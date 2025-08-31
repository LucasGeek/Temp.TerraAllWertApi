package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type imagePinRepository struct {
	db *gorm.DB
}

func NewImagePinRepository(db *gorm.DB) interfaces.ImagePinRepository {
	return &imagePinRepository{db: db}
}

func (r *imagePinRepository) Create(ctx context.Context, pin *entities.ImagePin) error {
	return r.db.WithContext(ctx).Create(pin).Error
}

func (r *imagePinRepository) GetByID(ctx context.Context, id string) (*entities.ImagePin, error) {
	var pin entities.ImagePin
	err := r.db.WithContext(ctx).
		Preload("GalleryImage").
		Preload("Apartment").
		Preload("Apartment.Floor").
		Preload("Apartment.Floor.Tower").
		First(&pin, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &pin, nil
}

func (r *imagePinRepository) GetByGalleryImageID(ctx context.Context, galleryImageID string) ([]*entities.ImagePin, error) {
	var pins []*entities.ImagePin
	err := r.db.WithContext(ctx).
		Preload("Apartment").
		Preload("Apartment.Floor").
		Preload("Apartment.Floor.Tower").
		Where("gallery_image_id = ?", galleryImageID).
		Find(&pins).Error
	return pins, err
}

func (r *imagePinRepository) Update(ctx context.Context, pin *entities.ImagePin) error {
	return r.db.WithContext(ctx).Save(pin).Error
}

func (r *imagePinRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.ImagePin{}, "id = ?", id).Error
}