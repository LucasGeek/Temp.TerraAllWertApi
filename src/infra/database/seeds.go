package database

import (
	"context"
	"log"

	"api/domain/entities"
	"api/domain/interfaces"

	"gorm.io/gorm"
)

func CreateInitialUser(db *gorm.DB, authService interfaces.AuthService) error {
	ctx := context.Background()

	var count int64
	err := db.Model(&entities.User{}).Count(&count).Error
	if err != nil {
		return err
	}

	if count > 0 {
		log.Println("Users already exist, skipping seed")
		return nil
	}

	hashedPassword, err := authService.HashPassword("admin123")
	if err != nil {
		return err
	}

	adminUser := &entities.User{
		Username: "admin",
		Email:    "admin@terraallwert.com",
		Password: hashedPassword,
		Role:     entities.RoleAdmin,
		Active:   true,
	}

	err = db.WithContext(ctx).Create(adminUser).Error
	if err != nil {
		return err
	}

	hashedPassword, err = authService.HashPassword("viewer123")
	if err != nil {
		return err
	}

	viewerUser := &entities.User{
		Username: "viewer",
		Email:    "viewer@terraallwert.com",
		Password: hashedPassword,
		Role:     entities.RoleViewer,
		Active:   true,
	}

	err = db.WithContext(ctx).Create(viewerUser).Error
	if err != nil {
		return err
	}

	log.Println("✅ Initial users created successfully")
	log.Println("   - Admin: admin / admin123")
	log.Println("   - Viewer: viewer / viewer123")

	return nil
}