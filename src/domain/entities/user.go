package entities

import (
	"database/sql/driver"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type UserRole string

const (
	UserRoleVisitor UserRole = "visitor"
	UserRoleManager UserRole = "manager"
	UserRoleAdmin   UserRole = "admin"
)

func (ur *UserRole) Scan(value interface{}) error {
	*ur = UserRole(value.(string))
	return nil
}

func (ur UserRole) Value() (driver.Value, error) {
	return string(ur), nil
}

type User struct {
	ID               uuid.UUID  `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	EnterpriseID     uuid.UUID  `json:"enterprise_id" gorm:"type:uuid;not null"`
	Enterprise       Enterprise `json:"enterprise,omitempty" gorm:"foreignKey:EnterpriseID;constraint:OnUpdate:CASCADE,OnDelete:CASCADE"`
	Name             string     `json:"name" gorm:"not null;size:255" validate:"required,min=2,max=255"`
	Email            string     `json:"email" gorm:"unique;not null;size:255" validate:"required,email"`
	PasswordHash     string     `json:"-" gorm:"not null;size:255"`
	Role             UserRole   `json:"role" gorm:"type:varchar(20);not null;default:visitor"`
	Phone            *string    `json:"phone,omitempty" gorm:"size:20"`
	AvatarFileID     *uuid.UUID `json:"avatar_file_id,omitempty" gorm:"type:uuid"`
	AvatarFile       *File      `json:"avatar_file,omitempty" gorm:"foreignKey:AvatarFileID;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	IsActive         bool       `json:"is_active" gorm:"not null;default:true"`
	EmailVerifiedAt  *time.Time `json:"email_verified_at,omitempty"`
	LastLoginAt      *time.Time `json:"last_login_at,omitempty"`
	CreatedAt        time.Time  `json:"created_at" gorm:"not null"`
	UpdatedAt        *time.Time `json:"updated_at,omitempty"`
	DeletedAt        gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	UploadedFiles []File         `json:"uploaded_files,omitempty" gorm:"foreignKey:UploadedBy"`
	AuditLogs     []AuditLog     `json:"audit_logs,omitempty" gorm:"foreignKey:UserID"`
	PropertyViews []PropertyView `json:"property_views,omitempty" gorm:"foreignKey:UserID"`
}

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.ID == uuid.Nil {
		u.ID = uuid.New()
	}
	return nil
}

func (u *User) TableName() string {
	return "users"
}