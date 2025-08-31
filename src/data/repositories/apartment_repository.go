package repositories

import (
	"context"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

type apartmentRepository struct {
	db *gorm.DB
}

func NewApartmentRepository(db *gorm.DB) interfaces.ApartmentRepository {
	return &apartmentRepository{db: db}
}

func (r *apartmentRepository) Create(ctx context.Context, apartment *entities.Apartment) error {
	return r.db.WithContext(ctx).Create(apartment).Error
}

func (r *apartmentRepository) GetByID(ctx context.Context, id string) (*entities.Apartment, error) {
	var apartment entities.Apartment
	err := r.db.WithContext(ctx).
		Preload("Floor").
		Preload("Floor.Tower").
		First(&apartment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &apartment, nil
}

func (r *apartmentRepository) GetByFloorID(ctx context.Context, floorID string) ([]*entities.Apartment, error) {
	var apartments []*entities.Apartment
	err := r.db.WithContext(ctx).
		Preload("Floor").
		Where("floor_id = ?", floorID).
		Find(&apartments).Error
	return apartments, err
}

func (r *apartmentRepository) GetByTowerID(ctx context.Context, towerID string) ([]*entities.Apartment, error) {
	var apartments []*entities.Apartment
	err := r.db.WithContext(ctx).
		Preload("Floor").
		Joins("JOIN floors ON apartments.floor_id = floors.id").
		Where("floors.tower_id = ?", towerID).
		Find(&apartments).Error
	return apartments, err
}

func (r *apartmentRepository) Update(ctx context.Context, apartment *entities.Apartment) error {
	return r.db.WithContext(ctx).Save(apartment).Error
}

func (r *apartmentRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.Apartment{}, "id = ?", id).Error
}

func (r *apartmentRepository) GetWithImages(ctx context.Context, id string) (*entities.Apartment, error) {
	var apartment entities.Apartment
	err := r.db.WithContext(ctx).
		Preload("Floor").
		Preload("Floor.Tower").
		Preload("Images").
		First(&apartment, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &apartment, nil
}

func (r *apartmentRepository) Search(ctx context.Context, criteria *entities.ApartmentSearchCriteria) ([]*entities.Apartment, error) {
	query := r.db.WithContext(ctx).
		Preload("Floor").
		Preload("Floor.Tower")

	if criteria.Number != nil {
		query = query.Where("number ILIKE ?", "%"+*criteria.Number+"%")
	}
	if criteria.Suites != nil {
		query = query.Where("suites = ?", *criteria.Suites)
	}
	if criteria.Bedrooms != nil {
		query = query.Where("bedrooms = ?", *criteria.Bedrooms)
	}
	if criteria.ParkingSpots != nil {
		query = query.Where("parking_spots = ?", *criteria.ParkingSpots)
	}
	if criteria.SolarPosition != nil {
		query = query.Where("solar_position ILIKE ?", "%"+*criteria.SolarPosition+"%")
	}
	if criteria.FloorID != nil {
		query = query.Where("floor_id = ?", *criteria.FloorID)
	}
	if criteria.TowerID != nil {
		query = query.Joins("JOIN floors ON apartments.floor_id = floors.id").
			Where("floors.tower_id = ?", *criteria.TowerID)
	}
	if criteria.PriceMin != nil {
		query = query.Where("price >= ?", *criteria.PriceMin)
	}
	if criteria.PriceMax != nil {
		query = query.Where("price <= ?", *criteria.PriceMax)
	}
	if criteria.Status != nil {
		query = query.Where("status = ?", *criteria.Status)
	}
	if criteria.Available != nil {
		query = query.Where("available = ?", *criteria.Available)
	}

	if criteria.Limit != nil && *criteria.Limit > 0 {
		query = query.Limit(*criteria.Limit)
	}
	if criteria.Offset != nil && *criteria.Offset > 0 {
		query = query.Offset(*criteria.Offset)
	}

	var apartments []*entities.Apartment
	err := query.Find(&apartments).Error
	return apartments, err
}

func (r *apartmentRepository) ExistsByFloorAndNumber(ctx context.Context, floorID, number string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Apartment{}).
		Where("floor_id = ? AND number = ?", floorID, number).
		Count(&count).Error
	return count > 0, err
}

func (r *apartmentRepository) GetAvailableCount(ctx context.Context) (int, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entities.Apartment{}).
		Where("available = ? AND status = ?", true, entities.ApartmentStatusAvailable).
		Count(&count).Error
	return int(count), err
}

func (r *apartmentRepository) GetByStatus(ctx context.Context, status entities.ApartmentStatus) ([]*entities.Apartment, error) {
	var apartments []*entities.Apartment
	err := r.db.WithContext(ctx).
		Preload("Floor").
		Preload("Floor.Tower").
		Where("status = ?", status).
		Find(&apartments).Error
	return apartments, err
}