package entities

import (
	"time"
)

type Floor struct {
	ID               string         `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Number           string         `json:"number" gorm:"not null"`
	TowerID          string         `json:"towerId" gorm:"not null;type:uuid"`
	Tower            Tower          `json:"tower" gorm:"constraint:OnDelete:CASCADE"`
	BannerURL        *string        `json:"bannerUrl"`
	BannerMetadata   *FileMetadata  `json:"bannerMetadata" gorm:"embedded;embeddedPrefix:banner_"`
	Apartments       []Apartment    `json:"apartments" gorm:"foreignKey:FloorID;constraint:OnDelete:CASCADE"`
	TotalApartments  int            `json:"totalApartments" gorm:"-"`
	CreatedAt        time.Time      `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt        time.Time      `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Floor) TableName() string {
	return "floors"
}