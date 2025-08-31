package database

import (
	"api/domain/entities"
	"gorm.io/gorm"
)

func AutoMigrate(db *gorm.DB) error {
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";").Error
	if err != nil {
		return err
	}

	return db.AutoMigrate(
		&entities.User{},
		&entities.Tower{},
		&entities.Floor{},
		&entities.Apartment{},
		&entities.ApartmentImage{},
		&entities.GalleryImage{},
		&entities.ImagePin{},
		&entities.AppConfig{},
	)
}