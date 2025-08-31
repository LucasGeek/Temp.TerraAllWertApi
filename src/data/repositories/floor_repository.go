package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type floorRepository struct {
	db *gorm.DB
}

func NewFloorRepository(db *gorm.DB) interfaces.FloorRepository {
	return &floorRepository{db: db}
}

func (r *floorRepository) Create(ctx context.Context, floor *entities.Floor) error {
	return r.db.WithContext(ctx).Create(floor).Error
}

func (r *floorRepository) GetByID(ctx context.Context, id string) (*entities.Floor, error) {
	var floor entities.Floor
	err := r.db.WithContext(ctx).Preload("Tower").First(&floor, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &floor, nil
}

func (r *floorRepository) GetByTowerID(ctx context.Context, towerID string) ([]*entities.Floor, error) {
	var floors []*entities.Floor
	err := r.db.WithContext(ctx).Where("tower_id = ?", towerID).Find(&floors).Error
	return floors, err
}

func (r *floorRepository) Update(ctx context.Context, floor *entities.Floor) error {
	return r.db.WithContext(ctx).Save(floor).Error
}

func (r *floorRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.Floor{}, "id = ?", id).Error
}

func (r *floorRepository) GetWithApartments(ctx context.Context, id string) (*entities.Floor, error) {
	var floor entities.Floor
	err := r.db.WithContext(ctx).
		Preload("Tower").
		Preload("Apartments").
		First(&floor, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &floor, nil
}

func (r *floorRepository) GetTotalApartments(ctx context.Context, id string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Apartment{}).Where("floor_id = ?", id).Count(&count).Error
	return int(count), err
}

func (r *floorRepository) ExistsByTowerAndNumber(ctx context.Context, towerID, number string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Floor{}).
		Where("tower_id = ? AND number = ?", towerID, number).
		Count(&count).Error
	return count > 0, err
}