package testutils

import (
	"time"

	"api/domain/entities"
)

// User fixtures
func CreateTestUser(role entities.UserRole) *entities.User {
	return &entities.User{
		ID:        "test-user-id",
		Username:  "testuser",
		Email:     "test@example.com",
		Password:  "hashedpassword",
		Role:      role,
		Active:    true,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func CreateAdminUser() *entities.User {
	return CreateTestUser(entities.RoleAdmin)
}

func CreateViewerUser() *entities.User {
	return CreateTestUser(entities.RoleViewer)
}

// Tower fixtures
func CreateTestTower() *entities.Tower {
	return &entities.Tower{
		ID:          "test-tower-id",
		Name:        "Test Tower",
		Description: stringPtr("Test tower description"),
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

// Floor fixtures
func CreateTestFloor(towerID string) *entities.Floor {
	return &entities.Floor{
		ID:              "test-floor-id",
		Number:          "1",
		TowerID:         towerID,
		BannerURL:       stringPtr("https://example.com/banner.jpg"),
		TotalApartments: 4,
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}
}

// Apartment fixtures
func CreateTestApartment(floorID string) *entities.Apartment {
	return &entities.Apartment{
		ID:            "test-apartment-id",
		Number:        "101",
		FloorID:       floorID,
		Area:          stringPtr("85m²"),
		Suites:        intPtr(2),
		Bedrooms:      intPtr(3),
		ParkingSpots:  intPtr(1),
		Status:        entities.ApartmentStatusAvailable,
		SolarPosition: stringPtr("Norte"),
		Price:         floatPtr(450000.0),
		Available:     true,
		MainImageURL:  stringPtr("https://example.com/main.jpg"),
		FloorPlanURL:  stringPtr("https://example.com/plan.jpg"),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}
}

// Gallery fixtures
func CreateTestGalleryImage() *entities.GalleryImage {
	return &entities.GalleryImage{
		ID:           "test-gallery-id",
		Route:        "home",
		ImageURL:     "https://example.com/gallery.jpg",
		ThumbnailURL: stringPtr("https://example.com/thumb.jpg"),
		Title:        stringPtr("Test Gallery Image"),
		Description:  stringPtr("Test description"),
		DisplayOrder: 1,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}
}

// ImagePin fixtures
func CreateTestImagePin(galleryImageID string) *entities.ImagePin {
	return &entities.ImagePin{
		ID:             "test-pin-id",
		GalleryImageID: galleryImageID,
		XCoord:         0.5,
		YCoord:         0.3,
		Title:          stringPtr("Test Pin"),
		Description:    stringPtr("Test pin description"),
		CreatedAt:      time.Now(),
	}
}

// JWT Claims fixtures
func CreateTestJWTClaims(role entities.UserRole) *entities.JWTClaims {
	return &entities.JWTClaims{
		UserID:   "test-user-id",
		Username: "testuser",
		Role:     role,
		Exp:      time.Now().Add(15 * time.Minute).Unix(),
		Iat:      time.Now().Unix(),
	}
}

// Login fixtures
func CreateTestLoginRequest() *entities.LoginRequest {
	return &entities.LoginRequest{
		Username: "testuser",
		Password: "testpassword",
	}
}

func CreateTestLoginResponse() *entities.LoginResponse {
	return &entities.LoginResponse{
		Token:        "test-access-token",
		RefreshToken: "test-refresh-token",
		ExpiresAt:    time.Now().Add(15 * time.Minute),
		User:         CreateTestUser(entities.RoleViewer),
	}
}

// Helper functions
func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func floatPtr(f float64) *float64 {
	return &f
}