package entities

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type EnterpriseStatus string

const (
	EnterpriseStatusActive       EnterpriseStatus = "active"
	EnterpriseStatusInactive     EnterpriseStatus = "inactive"
	EnterpriseStatusConstruction EnterpriseStatus = "construction"
	EnterpriseStatusCompleted    EnterpriseStatus = "completed"
)

func (es *EnterpriseStatus) Scan(value interface{}) error {
	*es = EnterpriseStatus(value.(string))
	return nil
}

func (es EnterpriseStatus) Value() (driver.Value, error) {
	return string(es), nil
}

type Enterprise struct {
	ID                   uuid.UUID        `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	Title                string           `json:"title" gorm:"not null;size:255" validate:"required,min=3,max=255"`
	Description          *string          `json:"description,omitempty" gorm:"type:text"`
	LogoFileID           *uuid.UUID       `json:"logo_file_id,omitempty" gorm:"type:uuid"`
	LogoFile             *File            `json:"logo_file,omitempty" gorm:"foreignKey:LogoFileID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	Slug                 string           `json:"slug" gorm:"unique;not null;size:255" validate:"required,slug"`
	AddressStreet        *string          `json:"address_street,omitempty" gorm:"size:255"`
	AddressNumber        *string          `json:"address_number,omitempty" gorm:"size:20"`
	AddressComplement    *string          `json:"address_complement,omitempty" gorm:"size:100"`
	AddressNeighborhood  *string          `json:"address_neighborhood,omitempty" gorm:"size:100"`
	AddressCity          string           `json:"address_city" gorm:"not null;size:100" validate:"required"`
	AddressState         string           `json:"address_state" gorm:"not null;size:2" validate:"required,len=2"`
	AddressZipCode       *string          `json:"address_zip_code,omitempty" gorm:"size:10"`
	Latitude             *float64         `json:"latitude,omitempty" gorm:"type:decimal(10,8)"`
	Longitude            *float64         `json:"longitude,omitempty" gorm:"type:decimal(11,8)"`
	Status               EnterpriseStatus `json:"status" gorm:"type:varchar(20);not null;default:active"`
	CreatedAt            time.Time        `json:"created_at" gorm:"not null"`
	UpdatedAt            *time.Time       `json:"updated_at,omitempty"`
	DeletedAt            gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Users    []User    `json:"users,omitempty" gorm:"foreignKey:EnterpriseID"`
	Menus    []Menu    `json:"menus,omitempty" gorm:"foreignKey:EnterpriseID"`
	AuditLog []AuditLog `json:"audit_logs,omitempty" gorm:"foreignKey:EnterpriseID"`
}

func (e *Enterprise) BeforeCreate(tx *gorm.DB) error {
	if e.ID == uuid.Nil {
		e.ID = uuid.New()
	}
	return nil
}

func (e *Enterprise) TableName() string {
	return "enterprises"
}