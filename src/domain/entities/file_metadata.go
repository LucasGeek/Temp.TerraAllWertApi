package entities

import (
	"time"
)

type FileMetadata struct {
	FileName    string     `json:"fileName" gorm:"not null"`
	FileSize    int64      `json:"fileSize" gorm:"not null"`
	ContentType string     `json:"contentType" gorm:"not null"`
	UploadedAt  time.Time  `json:"uploadedAt" gorm:"autoCreateTime"`
	Checksum    *string    `json:"checksum"`
	Width       *int       `json:"width"`
	Height      *int       `json:"height"`
}

type SignedUploadURL struct {
	UploadURL string                 `json:"uploadUrl"`
	AccessURL string                 `json:"accessUrl"`
	ExpiresIn int                    `json:"expiresIn"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
}

type BulkDownload struct {
	DownloadURL string    `json:"downloadUrl"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	ExpiresIn   int       `json:"expiresIn"`
	CreatedAt   time.Time `json:"createdAt"`
}

type BulkDownloadStatus struct {
	ID          string              `json:"id"`
	Status      BulkDownloadState   `json:"status"`
	Progress    int                 `json:"progress"`
	TotalFiles  int                 `json:"totalFiles"`
	ProcessedFiles int              `json:"processedFiles"`
	DownloadURL *string             `json:"downloadUrl,omitempty"`
	ErrorMessage *string            `json:"errorMessage,omitempty"`
	CreatedAt   time.Time           `json:"createdAt"`
	UpdatedAt   time.Time           `json:"updatedAt"`
}

type BulkDownloadState string

const (
	BulkDownloadStatePending    BulkDownloadState = "PENDING"
	BulkDownloadStateProcessing BulkDownloadState = "PROCESSING"
	BulkDownloadStateCompleted  BulkDownloadState = "COMPLETED"
	BulkDownloadStateFailed     BulkDownloadState = "FAILED"
)