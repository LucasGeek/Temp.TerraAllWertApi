package entities

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ScreenType string

const (
	ScreenTypePins      ScreenType = "pins"
	ScreenTypeFloorPlan ScreenType = "floor_plan"
	ScreenTypeCarousel  ScreenType = "carousel"
	ScreenTypeList      ScreenType = "list"
	ScreenTypeMap       ScreenType = "map"
)

func (st *ScreenType) Scan(value interface{}) error {
	*st = ScreenType(value.(string))
	return nil
}

func (st ScreenType) Value() (driver.Value, error) {
	return string(st), nil
}

type MenuType string

const (
	MenuTypeStandard MenuType = "standard"
	MenuTypeSubmenu  MenuType = "submenu"
)

func (mt *MenuType) Scan(value interface{}) error {
	*mt = MenuType(value.(string))
	return nil
}

func (mt MenuType) Value() (driver.Value, error) {
	return string(mt), nil
}

type Menu struct {
	ID             uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EnterpriseID   uuid.UUID      `json:"enterprise_id" gorm:"type:uuid;not null"`
	Enterprise     Enterprise     `json:"enterprise,omitempty" gorm:"foreignKey:EnterpriseID"`
	ParentMenuID   *uuid.UUID     `json:"parent_menu_id,omitempty" gorm:"type:uuid"`
	ParentMenu     *Menu          `json:"parent_menu,omitempty" gorm:"foreignKey:ParentMenuID"`
	Title          string         `json:"title" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Slug           string         `json:"slug" gorm:"not null;size:255" validate:"required,slug"`
	ScreenType     ScreenType     `json:"screen_type" gorm:"type:varchar(20);not null"`
	MenuType       MenuType       `json:"menu_type" gorm:"type:varchar(20);not null;default:standard"`
	Position       int            `json:"position" gorm:"not null;default:0"`
	Icon           *string        `json:"icon,omitempty" gorm:"size:50"`
	IsVisible      bool           `json:"is_visible" gorm:"not null;default:true"`
	PathHierarchy  *string        `json:"path_hierarchy,omitempty" gorm:"size:500"`
	DepthLevel     int            `json:"depth_level" gorm:"not null;default:0"`
	CreatedAt      time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt      *time.Time     `json:"updated_at,omitempty"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	SubMenus         []Menu           `json:"sub_menus,omitempty" gorm:"foreignKey:ParentMenuID"`
	MenuFloorPlan    *MenuFloorPlan   `json:"menu_floor_plan,omitempty" gorm:"foreignKey:MenuID"`
	MenuCarousel     *MenuCarousel    `json:"menu_carousel,omitempty" gorm:"foreignKey:MenuID"`
	MenuPins         *MenuPins        `json:"menu_pins,omitempty" gorm:"foreignKey:MenuID"`
}

func (m *Menu) BeforeCreate(tx *gorm.DB) error {
	if m.ID == uuid.Nil {
		m.ID = uuid.New()
	}
	return nil
}

func (m *Menu) TableName() string {
	return "menus"
}