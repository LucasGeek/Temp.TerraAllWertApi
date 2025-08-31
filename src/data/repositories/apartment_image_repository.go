package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type apartmentImageRepository struct {
	db *gorm.DB
}

func NewApartmentImageRepository(db *gorm.DB) interfaces.ApartmentImageRepository {
	return &apartmentImageRepository{db: db}
}

func (r *apartmentImageRepository) Create(ctx context.Context, image *entities.ApartmentImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *apartmentImageRepository) GetByID(ctx context.Context, id string) (*entities.ApartmentImage, error) {
	var image entities.ApartmentImage
	err := r.db.WithContext(ctx).
		Preload("Apartment").
		Preload("Apartment.Floor").
		Preload("Apartment.Floor.Tower").
		First(&image, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *apartmentImageRepository) GetByApartmentID(ctx context.Context, apartmentID string) ([]*entities.ApartmentImage, error) {
	var images []*entities.ApartmentImage
	err := r.db.WithContext(ctx).
		Where("apartment_id = ?", apartmentID).
		Order("order ASC").
		Find(&images).Error
	return images, err
}

func (r *apartmentImageRepository) Update(ctx context.Context, image *entities.ApartmentImage) error {
	return r.db.WithContext(ctx).Save(image).Error
}

func (r *apartmentImageRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.ApartmentImage{}, "id = ?", id).Error
}

func (r *apartmentImageRepository) ReorderImages(ctx context.Context, apartmentID string, imageIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, imageID := range imageIDs {
			if err := tx.Model(&entities.ApartmentImage{}).
				Where("id = ? AND apartment_id = ?", imageID, apartmentID).
				Update("order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}