package entities

import (
	"time"
)

type ApartmentStatus string

const (
	ApartmentStatusAvailable   ApartmentStatus = "AVAILABLE"
	ApartmentStatusReserved    ApartmentStatus = "RESERVED"
	ApartmentStatusSold        ApartmentStatus = "SOLD"
	ApartmentStatusMaintenance ApartmentStatus = "MAINTENANCE"
)

type Apartment struct {
	ID                    string             `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Number                string             `json:"number" gorm:"not null"`
	FloorID               string             `json:"floorId" gorm:"not null;type:uuid"`
	Floor                 Floor              `json:"floor" gorm:"constraint:OnDelete:CASCADE"`
	Area                  *string            `json:"area"`
	Suites                *int               `json:"suites"`
	Bedrooms              *int               `json:"bedrooms"`
	ParkingSpots          *int               `json:"parkingSpots"`
	Status                ApartmentStatus    `json:"status" gorm:"default:AVAILABLE"`
	MainImageURL          *string            `json:"mainImageUrl"`
	FloorPlanURL          *string            `json:"floorPlanUrl"`
	SolarPosition         *string            `json:"solarPosition"`
	Price                 *float64           `json:"price"`
	Available             bool               `json:"available" gorm:"default:true"`
	MainImageMetadata     *FileMetadata      `json:"mainImageMetadata" gorm:"embedded;embeddedPrefix:main_image_"`
	FloorPlanMetadata     *FileMetadata      `json:"floorPlanMetadata" gorm:"embedded;embeddedPrefix:floor_plan_"`
	Images                []ApartmentImage   `json:"images" gorm:"foreignKey:ApartmentID;constraint:OnDelete:CASCADE"`
	CreatedAt             time.Time          `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt             time.Time          `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Apartment) TableName() string {
	return "apartments"
}

type ApartmentImage struct {
	ID            string        `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	ApartmentID   string        `json:"apartmentId" gorm:"not null;type:uuid"`
	Apartment     Apartment     `json:"apartment" gorm:"constraint:OnDelete:CASCADE"`
	ImageURL      string        `json:"imageUrl" gorm:"not null"`
	ImageMetadata FileMetadata  `json:"imageMetadata" gorm:"embedded;embeddedPrefix:image_"`
	Description   *string       `json:"description"`
	Order         int           `json:"order" gorm:"default:0"`
	CreatedAt     time.Time     `json:"createdAt" gorm:"autoCreateTime"`
}

func (ApartmentImage) TableName() string {
	return "apartment_images"
}

type ApartmentSearchCriteria struct {
	Number        *string          `json:"number"`
	Suites        *int             `json:"suites"`
	Bedrooms      *int             `json:"bedrooms"`
	ParkingSpots  *int             `json:"parkingSpots"`
	SolarPosition *string          `json:"solarPosition"`
	TowerID       *string          `json:"towerId"`
	FloorID       *string          `json:"floorId"`
	PriceMin      *float64         `json:"priceMin"`
	PriceMax      *float64         `json:"priceMax"`
	AreaMin       *string          `json:"areaMin"`
	AreaMax       *string          `json:"areaMax"`
	Status        *ApartmentStatus `json:"status"`
	Available     *bool            `json:"available"`
	Limit         *int             `json:"limit"`
	Offset        *int             `json:"offset"`
}