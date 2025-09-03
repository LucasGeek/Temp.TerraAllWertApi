package interfaces

import (
	"context"

	"github.com/google/uuid"
	"terra-allwert/domain/entities"
)

type FileSearchFilters struct {
	FileType      *entities.FileType
	MimeType      *string
	Extension     *string
	UploadedBy    *uuid.UUID
	MinSize       *int64
	MaxSize       *int64
	WidthMin      *int
	WidthMax      *int
	HeightMin     *int
	HeightMax     *int
	DurationMin   *int
	DurationMax   *int
}

type FileRepository interface {
	Create(ctx context.Context, file *entities.File) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.File, error)
	GetByHash(ctx context.Context, hash string) (*entities.File, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.File, error)
	Update(ctx context.Context, file *entities.File) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByUploader(ctx context.Context, uploaderID uuid.UUID, limit, offset int) ([]*entities.File, error)
	GetByType(ctx context.Context, fileType entities.FileType, limit, offset int) ([]*entities.File, error)
	GetByMimeType(ctx context.Context, mimeType string, limit, offset int) ([]*entities.File, error)
	Search(ctx context.Context, filters FileSearchFilters, limit, offset int) ([]*entities.File, error)
	GetByStoragePath(ctx context.Context, storagePath string) (*entities.File, error)
	GetOrphaned(ctx context.Context, limit, offset int) ([]*entities.File, error)
	GetImagesByDimensions(ctx context.Context, minWidth, maxWidth, minHeight, maxHeight int, limit, offset int) ([]*entities.File, error)
}

type FileVariantRepository interface {
	Create(ctx context.Context, variant *entities.FileVariant) error
	GetByID(ctx context.Context, id uuid.UUID) (*entities.FileVariant, error)
	GetByOriginalFileID(ctx context.Context, originalFileID uuid.UUID, limit, offset int) ([]*entities.FileVariant, error)
	GetByVariantName(ctx context.Context, originalFileID uuid.UUID, variantName string) (*entities.FileVariant, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.FileVariant, error)
	Update(ctx context.Context, variant *entities.FileVariant) error
	Delete(ctx context.Context, id uuid.UUID) error
	DeleteByOriginalFile(ctx context.Context, originalFileID uuid.UUID) error
	GetByDimensions(ctx context.Context, width, height int, limit, offset int) ([]*entities.FileVariant, error)
}