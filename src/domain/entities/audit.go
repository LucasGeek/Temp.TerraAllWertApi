package entities

import (
	"database/sql/driver"
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AuditAction string

const (
	AuditActionCreate  AuditAction = "create"
	AuditActionUpdate  AuditAction = "update"
	AuditActionDelete  AuditAction = "delete"
	AuditActionRestore AuditAction = "restore"
	AuditActionLogin   AuditAction = "login"
	AuditActionLogout  AuditAction = "logout"
)

func (aa *AuditAction) Scan(value interface{}) error {
	*aa = AuditAction(value.(string))
	return nil
}

func (aa AuditAction) Value() (driver.Value, error) {
	return string(aa), nil
}

type JSONValues map[string]interface{}

func (jv *JSONValues) Scan(value interface{}) error {
	if value == nil {
		*jv = make(JSONValues)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, jv)
	case string:
		return json.Unmarshal([]byte(v), jv)
	}
	
	return nil
}

func (jv JSONValues) Value() (driver.Value, error) {
	if jv == nil {
		return nil, nil
	}
	return json.Marshal(jv)
}

type AuditLog struct {
	ID           uuid.UUID    `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	UserID       *uuid.UUID   `json:"user_id,omitempty" gorm:"type:uuid"`
	User         *User        `json:"user,omitempty" gorm:"foreignKey:UserID"`
	EnterpriseID *uuid.UUID   `json:"enterprise_id,omitempty" gorm:"type:uuid"`
	Enterprise   *Enterprise  `json:"enterprise,omitempty" gorm:"foreignKey:EnterpriseID"`
	EntityType   string       `json:"entity_type" gorm:"not null;size:100" validate:"required"`
	EntityID     uuid.UUID    `json:"entity_id" gorm:"type:uuid;not null"`
	Action       AuditAction  `json:"action" gorm:"type:varchar(20);not null"`
	OldValues    JSONValues   `json:"old_values,omitempty" gorm:"type:jsonb"`
	NewValues    JSONValues   `json:"new_values,omitempty" gorm:"type:jsonb"`
	IPAddress    *net.IP      `json:"ip_address,omitempty" gorm:"type:inet"`
	UserAgent    *string      `json:"user_agent,omitempty" gorm:"type:text"`
	CreatedAt    time.Time    `json:"created_at" gorm:"not null"`
}

func (al *AuditLog) BeforeCreate(db *gorm.DB) error {
	if al.ID == uuid.Nil {
		al.ID = uuid.New()
	}
	return nil
}

func (al *AuditLog) TableName() string {
	return "audit_logs"
}

type PropertyView struct {
	ID                  uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	SuiteID             uuid.UUID `json:"suite_id" gorm:"type:uuid;not null"`
	Suite               Suite     `json:"suite,omitempty" gorm:"foreignKey:SuiteID"`
	UserID              *uuid.UUID `json:"user_id,omitempty" gorm:"type:uuid"`
	User                *User     `json:"user,omitempty" gorm:"foreignKey:UserID"`
	SessionID           string    `json:"session_id" gorm:"not null;size:100" validate:"required"`
	IPAddress           *net.IP   `json:"ip_address,omitempty" gorm:"type:inet"`
	UserAgent           *string   `json:"user_agent,omitempty" gorm:"type:text"`
	Referrer            *string   `json:"referrer,omitempty" gorm:"type:text"`
	ViewDurationSeconds *int      `json:"view_duration_seconds,omitempty"`
	CreatedAt           time.Time `json:"created_at" gorm:"not null"`
}

func (pv *PropertyView) BeforeCreate(db *gorm.DB) error {
	if pv.ID == uuid.Nil {
		pv.ID = uuid.New()
	}
	return nil
}

func (pv *PropertyView) TableName() string {
	return "property_views"
}