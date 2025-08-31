package entities

import (
	"time"
)

type GalleryImage struct {
	ID                string        `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Route             string        `json:"route" gorm:"not null"`
	ImageURL          string        `json:"imageUrl" gorm:"not null"`
	ThumbnailURL      *string       `json:"thumbnailUrl"`
	ImageMetadata     FileMetadata  `json:"imageMetadata" gorm:"embedded;embeddedPrefix:image_"`
	ThumbnailMetadata *FileMetadata `json:"thumbnailMetadata" gorm:"embedded;embeddedPrefix:thumbnail_"`
	Title             *string       `json:"title"`
	Description       *string       `json:"description"`
	DisplayOrder      int           `json:"displayOrder" gorm:"default:0"`
	Pins              []ImagePin    `json:"pins" gorm:"foreignKey:GalleryImageID;constraint:OnDelete:CASCADE"`
	CreatedAt         time.Time     `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt         time.Time     `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (GalleryImage) TableName() string {
	return "gallery_images"
}

type ImagePin struct {
	ID             string       `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	GalleryImageID string       `json:"galleryImageId" gorm:"not null;type:uuid"`
	GalleryImage   GalleryImage `json:"galleryImage" gorm:"constraint:OnDelete:CASCADE"`
	XCoord         float64      `json:"xCoord" gorm:"not null"`
	YCoord         float64      `json:"yCoord" gorm:"not null"`
	Title          *string      `json:"title"`
	Description    *string      `json:"description"`
	ApartmentID    *string      `json:"apartmentId" gorm:"type:uuid"`
	Apartment      *Apartment   `json:"apartment,omitempty" gorm:"constraint:OnDelete:SET NULL"`
	LinkURL        *string      `json:"linkUrl"`
	CreatedAt      time.Time    `json:"createdAt" gorm:"autoCreateTime"`
}

func (ImagePin) TableName() string {
	return "image_pins"
}