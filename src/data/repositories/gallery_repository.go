package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type galleryRepository struct {
	db *gorm.DB
}

func NewGalleryRepository(db *gorm.DB) interfaces.GalleryRepository {
	return &galleryRepository{db: db}
}

func (r *galleryRepository) Create(ctx context.Context, image *entities.GalleryImage) error {
	return r.db.WithContext(ctx).Create(image).Error
}

func (r *galleryRepository) GetByID(ctx context.Context, id string) (*entities.GalleryImage, error) {
	var image entities.GalleryImage
	err := r.db.WithContext(ctx).
		Preload("Pins").
		Preload("Pins.Apartment").
		First(&image, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &image, nil
}

func (r *galleryRepository) GetByRoute(ctx context.Context, route string) ([]*entities.GalleryImage, error) {
	var images []*entities.GalleryImage
	err := r.db.WithContext(ctx).
		Where("route = ?", route).
		Order("display_order ASC").
		Find(&images).Error
	return images, err
}

func (r *galleryRepository) GetAllRoutes(ctx context.Context) ([]string, error) {
	var routes []string
	err := r.db.WithContext(ctx).
		Model(&entities.GalleryImage{}).
		Distinct("route").
		Pluck("route", &routes).Error
	return routes, err
}

func (r *galleryRepository) Update(ctx context.Context, image *entities.GalleryImage) error {
	return r.db.WithContext(ctx).Save(image).Error
}

func (r *galleryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.GalleryImage{}, "id = ?", id).Error
}

func (r *galleryRepository) ReorderImages(ctx context.Context, route string, imageIDs []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		for i, imageID := range imageIDs {
			if err := tx.Model(&entities.GalleryImage{}).
				Where("id = ? AND route = ?", imageID, route).
				Update("display_order", i).Error; err != nil {
				return err
			}
		}
		return nil
	})
}