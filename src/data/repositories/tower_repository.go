package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type towerRepository struct {
	db *gorm.DB
}

func NewTowerRepository(db *gorm.DB) interfaces.TowerRepository {
	return &towerRepository{db: db}
}

func (r *towerRepository) Create(ctx context.Context, tower *entities.Tower) error {
	return r.db.WithContext(ctx).Create(tower).Error
}

func (r *towerRepository) GetByID(ctx context.Context, id string) (*entities.Tower, error) {
	var tower entities.Tower
	err := r.db.WithContext(ctx).First(&tower, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tower, nil
}

func (r *towerRepository) GetAll(ctx context.Context) ([]*entities.Tower, error) {
	var towers []*entities.Tower
	err := r.db.WithContext(ctx).Find(&towers).Error
	return towers, err
}

func (r *towerRepository) Update(ctx context.Context, tower *entities.Tower) error {
	return r.db.WithContext(ctx).Save(tower).Error
}

func (r *towerRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.Tower{}, "id = ?", id).Error
}

func (r *towerRepository) GetWithFloors(ctx context.Context, id string) (*entities.Tower, error) {
	var tower entities.Tower
	err := r.db.WithContext(ctx).Preload("Floors").First(&tower, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &tower, nil
}

func (r *towerRepository) GetTotalApartments(ctx context.Context, id string) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Table("apartments").
		Joins("JOIN floors ON apartments.floor_id = floors.id").
		Where("floors.tower_id = ?", id).
		Count(&count).Error
	return int(count), err
}

func (r *towerRepository) ExistsByName(ctx context.Context, name string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Model(&entities.Tower{}).Where("name = ?", name).Count(&count).Error
	return count > 0, err
}