package entities

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type CarouselItemType string

const (
	CarouselItemTypeImage CarouselItemType = "image"
	CarouselItemTypeVideo CarouselItemType = "video"
	CarouselItemTypeMap   CarouselItemType = "map"
	CarouselItemTypeHTML  CarouselItemType = "html"
)

func (cit *CarouselItemType) Scan(value interface{}) error {
	*cit = CarouselItemType(value.(string))
	return nil
}

func (cit CarouselItemType) Value() (driver.Value, error) {
	return string(cit), nil
}

type MapType string

const (
	MapTypeStandard  MapType = "standard"
	MapTypeSatellite MapType = "satellite"
	MapTypeTerrain   MapType = "terrain"
	MapTypeHybrid    MapType = "hybrid"
)

func (mt *MapType) Scan(value interface{}) error {
	if value != nil {
		*mt = MapType(value.(string))
	}
	return nil
}

func (mt MapType) Value() (driver.Value, error) {
	return string(mt), nil
}

type MenuCarousel struct {
	ID                   uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MenuID               uuid.UUID  `json:"menu_id" gorm:"type:uuid;unique;not null"`
	Menu                 Menu       `json:"menu,omitempty" gorm:"foreignKey:MenuID"`
	PromotionalVideoID   *uuid.UUID `json:"promotional_video_id,omitempty" gorm:"type:uuid"`
	PromotionalVideo     *File      `json:"promotional_video,omitempty" gorm:"foreignKey:PromotionalVideoID"`
	Autoplay             bool       `json:"autoplay" gorm:"default:true"`
	AutoplayInterval     int        `json:"autoplay_interval" gorm:"default:5000"`
	ShowIndicators       bool       `json:"show_indicators" gorm:"default:true"`
	ShowControls         bool       `json:"show_controls" gorm:"default:true"`
	TransitionType       string     `json:"transition_type" gorm:"size:50;default:slide"`
	CreatedAt            time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt            *time.Time `json:"updated_at,omitempty"`

	// Relationships
	CarouselItems []CarouselItem `json:"carousel_items,omitempty" gorm:"foreignKey:MenuCarouselID"`
}

func (mc *MenuCarousel) BeforeCreate(tx *gorm.DB) error {
	if mc.ID == uuid.Nil {
		mc.ID = uuid.New()
	}
	return nil
}

func (mc *MenuCarousel) TableName() string {
	return "menu_carousels"
}

type CarouselItem struct {
	ID               uuid.UUID        `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MenuCarouselID   uuid.UUID        `json:"menu_carousel_id" gorm:"type:uuid;not null"`
	MenuCarousel     MenuCarousel     `json:"menu_carousel,omitempty" gorm:"foreignKey:MenuCarouselID"`
	ItemType         CarouselItemType `json:"item_type" gorm:"type:varchar(20);not null"`
	BackgroundFileID *uuid.UUID       `json:"background_file_id,omitempty" gorm:"type:uuid"`
	BackgroundFile   *File            `json:"background_file,omitempty" gorm:"foreignKey:BackgroundFileID"`
	Position         int              `json:"position" gorm:"not null;default:0"`
	Title            *string          `json:"title,omitempty" gorm:"size:255"`
	Subtitle         *string          `json:"subtitle,omitempty" gorm:"size:500"`
	CtaText          *string          `json:"cta_text,omitempty" gorm:"size:100"`
	CtaURL           *string          `json:"cta_url,omitempty" gorm:"size:500"`
	MapType          *MapType         `json:"map_type,omitempty" gorm:"type:varchar(20)"`
	MapLatitude      *float64         `json:"map_latitude,omitempty" gorm:"type:decimal(10,8)"`
	MapLongitude     *float64         `json:"map_longitude,omitempty" gorm:"type:decimal(11,8)"`
	MapZoom          *int             `json:"map_zoom,omitempty" gorm:"default:15"`
	IsActive         bool             `json:"is_active" gorm:"default:true"`
	ValidFrom        *time.Time       `json:"valid_from,omitempty"`
	ValidUntil       *time.Time       `json:"valid_until,omitempty"`
	CreatedAt        time.Time        `json:"created_at" gorm:"not null"`
	UpdatedAt        *time.Time       `json:"updated_at,omitempty"`
	DeletedAt        gorm.DeletedAt   `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	TextOverlays []CarouselTextOverlay `json:"text_overlays,omitempty" gorm:"foreignKey:CarouselItemID"`
}

func (ci *CarouselItem) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == uuid.Nil {
		ci.ID = uuid.New()
	}
	return nil
}

func (ci *CarouselItem) TableName() string {
	return "carousel_items"
}

type CarouselTextOverlay struct {
	ID              uuid.UUID    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	CarouselItemID  uuid.UUID    `json:"carousel_item_id" gorm:"type:uuid;not null"`
	CarouselItem    CarouselItem `json:"carousel_item,omitempty" gorm:"foreignKey:CarouselItemID"`
	Title           *string      `json:"title,omitempty" gorm:"size:255"`
	Description     *string      `json:"description,omitempty" gorm:"type:text"`
	TextColor       string       `json:"text_color" gorm:"size:7;default:#FFFFFF"`
	TextSize        string       `json:"text_size" gorm:"size:20;default:medium"`
	BackgroundColor *string      `json:"background_color,omitempty" gorm:"size:9"`
	PositionX       float64      `json:"position_x" gorm:"type:decimal(5,2);not null"`
	PositionY       float64      `json:"position_y" gorm:"type:decimal(5,2);not null"`
	AnimationType   *string      `json:"animation_type,omitempty" gorm:"size:50"`
	CreatedAt       time.Time    `json:"created_at" gorm:"not null"`
	UpdatedAt       *time.Time   `json:"updated_at,omitempty"`
}

func (cto *CarouselTextOverlay) BeforeCreate(tx *gorm.DB) error {
	if cto.ID == uuid.Nil {
		cto.ID = uuid.New()
	}
	return nil
}

func (cto *CarouselTextOverlay) TableName() string {
	return "carousel_text_overlays"
}