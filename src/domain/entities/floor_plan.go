package entities

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type SunPosition string

const (
	SunPositionN  SunPosition = "N"
	SunPositionNE SunPosition = "NE"
	SunPositionE  SunPosition = "E"
	SunPositionSE SunPosition = "SE"
	SunPositionS  SunPosition = "S"
	SunPositionSW SunPosition = "SW"
	SunPositionW  SunPosition = "W"
	SunPositionNW SunPosition = "NW"
)

func (sp *SunPosition) Scan(value interface{}) error {
	if value != nil {
		*sp = SunPosition(value.(string))
	}
	return nil
}

func (sp SunPosition) Value() (driver.Value, error) {
	return string(sp), nil
}

type SuiteStatus string

const (
	SuiteStatusAvailable   SuiteStatus = "available"
	SuiteStatusReserved    SuiteStatus = "reserved"
	SuiteStatusSold        SuiteStatus = "sold"
	SuiteStatusUnavailable SuiteStatus = "unavailable"
)

func (ss *SuiteStatus) Scan(value interface{}) error {
	*ss = SuiteStatus(value.(string))
	return nil
}

func (ss SuiteStatus) Value() (driver.Value, error) {
	return string(ss), nil
}

type MenuFloorPlan struct {
	ID                     uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MenuID                 uuid.UUID `json:"menu_id" gorm:"type:uuid;unique;not null"`
	Menu                   Menu      `json:"menu,omitempty" gorm:"foreignKey:MenuID"`
	DefaultView            *string   `json:"default_view,omitempty" gorm:"size:50"`
	EnableUnitFilters      bool      `json:"enable_unit_filters" gorm:"default:true"`
	EnableUnitComparison   bool      `json:"enable_unit_comparison" gorm:"default:true"`
	CreatedAt              time.Time `json:"created_at" gorm:"not null"`
	UpdatedAt              *time.Time `json:"updated_at,omitempty"`

	// Relationships
	Towers []Tower `json:"towers,omitempty" gorm:"foreignKey:MenuFloorPlanID"`
}

func (mfp *MenuFloorPlan) BeforeCreate(tx *gorm.DB) error {
	if mfp.ID == uuid.Nil {
		mfp.ID = uuid.New()
	}
	return nil
}

func (mfp *MenuFloorPlan) TableName() string {
	return "menu_floor_plans"
}

type Tower struct {
	ID                uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	MenuFloorPlanID   uuid.UUID      `json:"menu_floor_plan_id" gorm:"type:uuid;not null"`
	MenuFloorPlan     MenuFloorPlan  `json:"menu_floor_plan,omitempty" gorm:"foreignKey:MenuFloorPlanID"`
	Title             string         `json:"title" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Description       *string        `json:"description,omitempty" gorm:"type:text"`
	BuildingCode      *string        `json:"building_code,omitempty" gorm:"size:50"`
	TotalFloors       *int           `json:"total_floors,omitempty"`
	UnitsPerFloor     *int           `json:"units_per_floor,omitempty"`
	Position          int            `json:"position" gorm:"not null;default:0"`
	CreatedAt         time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt         *time.Time     `json:"updated_at,omitempty"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Floors []Floor `json:"floors,omitempty" gorm:"foreignKey:TowerID"`
}

func (t *Tower) BeforeCreate(tx *gorm.DB) error {
	if t.ID == uuid.Nil {
		t.ID = uuid.New()
	}
	return nil
}

func (t *Tower) TableName() string {
	return "towers"
}

type Floor struct {
	ID               uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	TowerID          uuid.UUID      `json:"tower_id" gorm:"type:uuid;not null"`
	Tower            Tower          `json:"tower,omitempty" gorm:"foreignKey:TowerID"`
	FloorNumber      int            `json:"floor_number" gorm:"not null"`
	FloorName        *string        `json:"floor_name,omitempty" gorm:"size:100"`
	BannerFileID     *uuid.UUID     `json:"banner_file_id,omitempty" gorm:"type:uuid"`
	BannerFile       *File          `json:"banner_file,omitempty" gorm:"foreignKey:BannerFileID"`
	FloorPlanFileID  *uuid.UUID     `json:"floor_plan_file_id,omitempty" gorm:"type:uuid"`
	FloorPlanFile    *File          `json:"floor_plan_file,omitempty" gorm:"foreignKey:FloorPlanFileID"`
	CreatedAt        time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt        *time.Time     `json:"updated_at,omitempty"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Suites []Suite `json:"suites,omitempty" gorm:"foreignKey:FloorID"`
}

func (f *Floor) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

func (f *Floor) TableName() string {
	return "floors"
}

type Suite struct {
	ID               uuid.UUID      `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FloorID          uuid.UUID      `json:"floor_id" gorm:"type:uuid;not null"`
	Floor            Floor          `json:"floor,omitempty" gorm:"foreignKey:FloorID"`
	UnitNumber       string         `json:"unit_number" gorm:"not null;size:20" validate:"required"`
	Title            string         `json:"title" gorm:"not null;size:255" validate:"required,min=1,max=255"`
	Description      *string        `json:"description,omitempty" gorm:"type:text"`
	PositionX        *float64       `json:"position_x,omitempty" gorm:"type:decimal(6,2)"`
	PositionY        *float64       `json:"position_y,omitempty" gorm:"type:decimal(6,2)"`
	AreaSqm          float64        `json:"area_sqm" gorm:"type:decimal(10,2);not null" validate:"required,min=1"`
	Bedrooms         int            `json:"bedrooms" gorm:"not null;default:0"`
	SuitesCount      int            `json:"suites_count" gorm:"not null;default:0"`
	Bathrooms        int            `json:"bathrooms" gorm:"not null;default:0"`
	ParkingSpaces    *int           `json:"parking_spaces,omitempty" gorm:"default:0"`
	SunPosition      *SunPosition   `json:"sun_position,omitempty" gorm:"type:varchar(2)"`
	Status           SuiteStatus    `json:"status" gorm:"type:varchar(20);not null;default:available"`
	FloorPlanFileID  *uuid.UUID     `json:"floor_plan_file_id,omitempty" gorm:"type:uuid"`
	FloorPlanFile    *File          `json:"floor_plan_file,omitempty" gorm:"foreignKey:FloorPlanFileID"`
	Price            *float64       `json:"price,omitempty" gorm:"type:decimal(15,2)"`
	CreatedAt        time.Time      `json:"created_at" gorm:"not null"`
	UpdatedAt        *time.Time     `json:"updated_at,omitempty"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	PropertyViews []PropertyView `json:"property_views,omitempty" gorm:"foreignKey:SuiteID"`
}

func (s *Suite) BeforeCreate(tx *gorm.DB) error {
	if s.ID == uuid.Nil {
		s.ID = uuid.New()
	}
	return nil
}

func (s *Suite) TableName() string {
	return "suites"
}