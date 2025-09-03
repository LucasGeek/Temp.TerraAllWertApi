package entities

import (
	"database/sql/driver"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type FileType string

const (
	FileTypeImage    FileType = "image"
	FileTypeVideo    FileType = "video"
	FileTypeDocument FileType = "document"
)

func (ft *FileType) Scan(value interface{}) error {
	*ft = FileType(value.(string))
	return nil
}

func (ft FileType) Value() (driver.Value, error) {
	return string(ft), nil
}

type FileMetadata map[string]interface{}

func (fm *FileMetadata) Scan(value interface{}) error {
	if value == nil {
		*fm = make(FileMetadata)
		return nil
	}
	
	switch v := value.(type) {
	case []byte:
		return json.Unmarshal(v, fm)
	case string:
		return json.Unmarshal([]byte(v), fm)
	}
	
	return nil
}

func (fm FileMetadata) Value() (driver.Value, error) {
	if fm == nil {
		return nil, nil
	}
	return json.Marshal(fm)
}

type File struct {
	ID             uuid.UUID     `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	FileType       FileType      `json:"file_type" gorm:"type:varchar(20);not null"`
	MimeType       string        `json:"mime_type" gorm:"not null;size:100" validate:"required"`
	Extension      string        `json:"extension" gorm:"not null;size:10" validate:"required"`
	OriginalName   string        `json:"original_name" gorm:"not null;size:255" validate:"required"`
	StoragePath    string        `json:"storage_path" gorm:"not null;size:500" validate:"required"`
	CdnURL         *string       `json:"cdn_url,omitempty" gorm:"size:500"`
	FileSizeBytes  int64         `json:"file_size_bytes" gorm:"not null" validate:"min=1"`
	FileHash       *string       `json:"file_hash,omitempty" gorm:"unique;size:64"`
	Width          *int          `json:"width,omitempty"`
	Height         *int          `json:"height,omitempty"`
	DurationSeconds *int         `json:"duration_seconds,omitempty"`
	Metadata       FileMetadata  `json:"metadata,omitempty" gorm:"type:jsonb"`
	UploadedBy     *uuid.UUID    `json:"uploaded_by,omitempty" gorm:"type:uuid"`
	Uploader       *User         `json:"uploader,omitempty" gorm:"foreignKey:UploadedBy;constraint:OnUpdate:CASCADE,OnDelete:SET NULL"`
	CreatedAt      time.Time     `json:"created_at" gorm:"not null"`
	UpdatedAt      *time.Time    `json:"updated_at,omitempty"`
	DeletedAt      gorm.DeletedAt `json:"deleted_at,omitempty" gorm:"index"`

	// Relationships
	Variants []FileVariant `json:"variants,omitempty" gorm:"foreignKey:OriginalFileID"`
}

func (f *File) BeforeCreate(tx *gorm.DB) error {
	if f.ID == uuid.Nil {
		f.ID = uuid.New()
	}
	return nil
}

func (f *File) TableName() string {
	return "files"
}

type FileVariant struct {
	ID             uuid.UUID `json:"id" gorm:"primaryKey;type:uuid;default:gen_random_uuid()"`
	OriginalFileID uuid.UUID `json:"original_file_id" gorm:"type:uuid;not null"`
	OriginalFile   File      `json:"original_file,omitempty" gorm:"foreignKey:OriginalFileID"`
	VariantName    string    `json:"variant_name" gorm:"not null;size:50" validate:"required"`
	StoragePath    string    `json:"storage_path" gorm:"not null;size:500" validate:"required"`
	CdnURL         *string   `json:"cdn_url,omitempty" gorm:"size:500"`
	Width          int       `json:"width" gorm:"not null"`
	Height         int       `json:"height" gorm:"not null"`
	FileSizeBytes  int64     `json:"file_size_bytes" gorm:"not null" validate:"min=1"`
	CreatedAt      time.Time `json:"created_at" gorm:"not null"`
}

func (fv *FileVariant) BeforeCreate(tx *gorm.DB) error {
	if fv.ID == uuid.Nil {
		fv.ID = uuid.New()
	}
	return nil
}

func (fv *FileVariant) TableName() string {
	return "file_variants"
}