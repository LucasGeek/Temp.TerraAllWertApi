package entities

import (
	"time"
)

type Tower struct {
	ID              string    `json:"id" gorm:"primaryKey;type:uuid;default:uuid_generate_v4()"`
	Name            string    `json:"name" gorm:"not null"`
	Description     *string   `json:"description"`
	Floors          []Floor   `json:"floors" gorm:"foreignKey:TowerID;constraint:OnDelete:CASCADE"`
	TotalApartments int       `json:"totalApartments" gorm:"-"`
	CreatedAt       time.Time `json:"createdAt" gorm:"autoCreateTime"`
	UpdatedAt       time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (Tower) TableName() string {
	return "towers"
}