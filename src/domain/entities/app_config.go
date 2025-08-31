package entities

import (
	"time"
)

type AppConfig struct {
	ID                  int       `json:"id" gorm:"primaryKey;autoIncrement"`
	LogoURL             *string   `json:"logoUrl"`
	APIBaseURL          string    `json:"apiBaseUrl" gorm:"not null"`
	MinioBaseURL        string    `json:"minioBaseUrl" gorm:"not null"`
	AppVersion          string    `json:"appVersion" gorm:"not null"`
	CacheControlMaxAge  int       `json:"cacheControlMaxAge" gorm:"default:3600"`
	UpdatedAt           time.Time `json:"updatedAt" gorm:"autoUpdateTime"`
}

func (AppConfig) TableName() string {
	return "app_config"
}