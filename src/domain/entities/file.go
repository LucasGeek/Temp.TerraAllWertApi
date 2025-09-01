package entities

import "time"

// FileContextInput representa o contexto de um arquivo
type FileContextInput struct {
	PinID        *string            `json:"pinId"`
	Coordinates  *CoordinatesInput  `json:"coordinates"`
	FloorID      *string            `json:"floorId"`
	FloorNumber  *string            `json:"floorNumber"`
	IsReference  *bool              `json:"isReference"`
	CarouselID   *string            `json:"carouselId"`
	Order        *int               `json:"order"`
	Title        *string            `json:"title"`
	Description  *string            `json:"description"`
	Tags         []string           `json:"tags"`
}

// CoordinatesInput representa coordenadas geográficas
type CoordinatesInput struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
}

// SignedUploadURLInput representa entrada para URL de upload
type SignedUploadURLInput struct {
	FileName    string            `json:"fileName"`
	FileType    string            `json:"fileType"`
	ContentType string            `json:"contentType"`
	RouteID     string            `json:"routeId"`
	Context     *FileContextInput `json:"context"`
}

// SignedUploadURLResponse representa resposta de URL de upload
type SignedUploadURLResponse struct {
	UploadURL string    `json:"uploadUrl"`
	MinioPath string    `json:"minioPath"`
	ExpiresAt time.Time `json:"expiresAt"`
	FileID    string    `json:"fileId"`
}

// ConfirmFileUploadInput representa entrada para confirmação de upload
type ConfirmFileUploadInput struct {
	FileID           string            `json:"fileId"`
	MinioPath        string            `json:"minioPath"`
	RouteID          string            `json:"routeId"`
	OriginalFileName string            `json:"originalFileName"`
	FileSize         int               `json:"fileSize"`
	Checksum         string            `json:"checksum"`
	FileType         string            `json:"fileType"`
	Context          *FileContextInput `json:"context"`
}

// ConfirmFileUploadResponse representa resposta de confirmação de upload
type ConfirmFileUploadResponse struct {
	Success      bool                     `json:"success"`
	FileMetadata *FileMetadataExtended    `json:"fileMetadata"`
}

// FileMetadataExtended representa metadados estendidos de arquivo
type FileMetadataExtended struct {
	ID           string                 `json:"id"`
	URL          string                 `json:"url"`
	DownloadURL  string                 `json:"downloadUrl"`
	ThumbnailURL string                 `json:"thumbnailUrl"`
	Metadata     map[string]interface{} `json:"metadata"`
}

// SignedDownloadURLsInput representa entrada para URLs de download
type SignedDownloadURLsInput struct {
	RouteID           string   `json:"routeId"`
	FileIDs           []string `json:"fileIds"`
	ExpirationMinutes int      `json:"expirationMinutes"`
}

// FileDownloadURL representa URL de download de arquivo
type FileDownloadURL struct {
	FileID      string    `json:"fileId"`
	DownloadURL string    `json:"downloadUrl"`
	ExpiresAt   time.Time `json:"expiresAt"`
}

// SignedDownloadURLsResponse representa resposta de URLs de download
type SignedDownloadURLsResponse struct {
	URLs []*FileDownloadURL `json:"urls"`
}