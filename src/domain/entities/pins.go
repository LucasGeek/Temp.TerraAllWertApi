package entities

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PinAction string

const (
	PinActionInfo     PinAction = "info"
	PinActionCarousel PinAction = "carousel"
	PinActionLink     PinAction = "link"
	PinActionModal    PinAction = "modal"
)

func (pa *PinAction) Scan(value interface{}) error {
	*pa = PinAction(value.(string))
	return nil
}

func (pa PinAction) Value() (driver.Value, error) {
	return string(pa), nil
}

type ActionData map[string]interface{}

func (ad *ActionData) Scan(value interface{}) error {
	if value == nil {
		*ad = make(ActionData)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, ad)
	case string:
		return json.Unmarshal([]byte(v), ad)
	}
	
	return nil
}

func (ad ActionData) Value() (driver.Value, error) {
	if ad == nil {
		return nil, nil
	}
	return json.Marshal(ad)
}

type MenuPins struct {
	ID                  uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MenuID              uuid.UUID  `json:"menu_id" gorm:"type:uuid;unique;not null"`
	Menu                Menu       `json:"menu,omitempty" gorm:"foreignKey:MenuID"`
	BackgroundFileID    *uuid.UUID `json:"background_file_id,omitempty" gorm:"type:uuid"`
	BackgroundFile      *File      `json:"background_file,omitempty" gorm:"foreignKey:BackgroundFileID"`
	PromotionalVideoID  *uuid.UUID `json:"promotional_video_id,omitempty" gorm:"type:uuid"`
	PromotionalVideo    *File      `json:"promotional_video,omitempty" gorm:"foreignKey:PromotionalVideoID"`
	EnableZoom          bool       `json:"enable_zoom" gorm:"default:true"`
	EnablePan           bool       `json:"enable_pan" gorm:"default:true"`
	MinZoom             float64    `json:"min_zoom" gorm:"type:decimal(3,2);default:0.5"`
	MaxZoom             float64    `json:"max_zoom" gorm:"type:decimal(3,2);default:3.0"`
	CreatedAt           time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt           *time.Time `json:"updated_at,omitempty"`

	// Relationships
	PinMarkers []PinMarker `json:"pin_markers,omitempty" gorm:"foreignKey:MenuPinID"`
}

func (mp *MenuPins) BeforeCreate(tx *gorm.DB) error {
	if mp.ID == uuid.Nil {
		mp.ID = uuid.New()
	}
	return nil
}

func (mp *MenuPins) TableName() string {
	return "menu_pins"
}

type PinMarker struct {
	ID           uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MenuPinID    uuid.UUID      `json:"menu_pin_id" gorm:"type:uuid;not null"`
	MenuPin      MenuPins       `json:"menu_pin,omitempty" gorm:"foreignKey:MenuPinID"`
	Title        string         `json:"title" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Description  *string        `json:"description,omitempty" gorm:"type:text"`
	PositionX    float64        `json:"position_x" gorm:"type:decimal(5,2);not null"`
	PositionY    float64        `json:"position_y" gorm:"type:decimal(5,2);not null"`
	IconType     string         `json:"icon_type" gorm:"size:50;default:default"`
	IconColor    string         `json:"icon_color" gorm:"size:7;default:#FF0000"`
	ActionType   PinAction      `json:"action_type" gorm:"type:varchar(20);default:info"`
	ActionData   ActionData     `json:"action_data,omitempty" gorm:"type:jsonb"`
	IsVisible    bool           `json:"is_visible" gorm:"default:true"`
	CreatedAt    time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt    *time.Time     `json:"updated_at,omitempty"`
	DeletedAt    gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Images []PinMarkerImage `json:"images,omitempty" gorm:"foreignKey:PinMarkerID"`
}

func (pm *PinMarker) BeforeCreate(tx *gorm.DB) error {
	if pm.ID == uuid.Nil {
		pm.ID = uuid.New()
	}
	return nil
}

func (pm *PinMarker) TableName() string {
	return "pin_markers"
}

type PinMarkerImage struct {
	ID           uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	PinMarkerID  uuid.UUID `json:"pin_marker_id" gorm:"type:uuid;not null"`
	PinMarker    PinMarker `json:"pin_marker,omitempty" gorm:"foreignKey:PinMarkerID"`
	FileID       uuid.UUID `json:"file_id" gorm:"type:uuid;not null"`
	File         File      `json:"file,omitempty" gorm:"foreignKey:FileID"`
	Position     int       `json:"position" gorm:"not null;default:0"`
	Caption      *string   `json:"caption,omitempty" gorm:"size:500"`
	CreatedAt    time.Time `json:"created_at" gorm:"not null"`
}

func (pmi *PinMarkerImage) BeforeCreate(tx *gorm.DB) error {
	if pmi.ID == uuid.Nil {
		pmi.ID = uuid.New()
	}
	return nil
}

func (pmi *PinMarkerImage) TableName() string {
	return "pin_marker_images"
}