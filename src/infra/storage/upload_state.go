package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// UploadState represents the state of a resumable upload
type UploadState struct {
	UploadID     string            `json:"upload_id"`
	ObjectKey    string            `json:"object_key"`
	TotalSize    int64             `json:"total_size"`
	UploadedSize int64             `json:"uploaded_size"`
	Parts        map[int]string    `json:"parts"` // part number -> etag
	ContentType  string            `json:"content_type"`
	UserID       string            `json:"user_id"`
	CreatedAt    time.Time         `json:"created_at"`
	UpdatedAt    time.Time         `json:"updated_at"`
	Status       UploadStatus      `json:"status"`
}

type UploadStatus string

const (
	UploadStatusInitiated  UploadStatus = "initiated"
	UploadStatusInProgress UploadStatus = "in_progress"
	UploadStatusCompleted  UploadStatus = "completed"
	UploadStatusFailed     UploadStatus = "failed"
	UploadStatusAborted    UploadStatus = "aborted"
)

// UploadStateManager manages resumable upload states using Redis
type UploadStateManager struct {
	redisClient *redis.Client
	keyPrefix   string
	ttl         time.Duration
}

func NewUploadStateManager(redisClient *redis.Client) *UploadStateManager {
	return &UploadStateManager{
		redisClient: redisClient,
		keyPrefix:   "upload_state:",
		ttl:         24 * time.Hour, // Upload states expire after 24 hours
	}
}

// SaveUploadState saves or updates an upload state
func (usm *UploadStateManager) SaveUploadState(ctx context.Context, state *UploadState) error {
	state.UpdatedAt = time.Now()
	
	data, err := json.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal upload state: %w", err)
	}

	key := usm.keyPrefix + state.UploadID
	err = usm.redisClient.Set(ctx, key, data, usm.ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to save upload state: %w", err)
	}

	return nil
}

// GetUploadState retrieves an upload state by upload ID
func (usm *UploadStateManager) GetUploadState(ctx context.Context, uploadID string) (*UploadState, error) {
	key := usm.keyPrefix + uploadID
	data, err := usm.redisClient.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, fmt.Errorf("upload state not found: %s", uploadID)
		}
		return nil, fmt.Errorf("failed to get upload state: %w", err)
	}

	var state UploadState
	err = json.Unmarshal([]byte(data), &state)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal upload state: %w", err)
	}

	return &state, nil
}

// DeleteUploadState removes an upload state
func (usm *UploadStateManager) DeleteUploadState(ctx context.Context, uploadID string) error {
	key := usm.keyPrefix + uploadID
	err := usm.redisClient.Del(ctx, key).Err()
	if err != nil {
		return fmt.Errorf("failed to delete upload state: %w", err)
	}
	return nil
}

// UpdatePartProgress updates the progress of a specific part
func (usm *UploadStateManager) UpdatePartProgress(ctx context.Context, uploadID string, partNumber int, etag string, partSize int64) error {
	state, err := usm.GetUploadState(ctx, uploadID)
	if err != nil {
		return err
	}

	if state.Parts == nil {
		state.Parts = make(map[int]string)
	}

	state.Parts[partNumber] = etag
	state.UploadedSize += partSize
	state.Status = UploadStatusInProgress

	return usm.SaveUploadState(ctx, state)
}

// GetUserUploads returns all active uploads for a user
func (usm *UploadStateManager) GetUserUploads(ctx context.Context, userID string) ([]*UploadState, error) {
	pattern := usm.keyPrefix + "*"
	keys, err := usm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get upload keys: %w", err)
	}

	var userUploads []*UploadState
	for _, key := range keys {
		data, err := usm.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue // Skip invalid entries
		}

		var state UploadState
		if err := json.Unmarshal([]byte(data), &state); err != nil {
			continue
		}

		if state.UserID == userID {
			userUploads = append(userUploads, &state)
		}
	}

	return userUploads, nil
}

// CleanupExpiredUploads removes expired upload states (called by background job)
func (usm *UploadStateManager) CleanupExpiredUploads(ctx context.Context, minioService *MinIOService) error {
	pattern := usm.keyPrefix + "*"
	keys, err := usm.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to get upload keys for cleanup: %w", err)
	}

	now := time.Now()
	for _, key := range keys {
		data, err := usm.redisClient.Get(ctx, key).Result()
		if err != nil {
			continue
		}

		var state UploadState
		if err := json.Unmarshal([]byte(data), &state); err != nil {
			continue
		}

		// Clean up uploads older than 24 hours that are not completed
		if state.Status != UploadStatusCompleted && now.Sub(state.UpdatedAt) > usm.ttl {
			// Abort the multipart upload in MinIO
			minioService.AbortMultipartUpload(ctx, state.ObjectKey, state.UploadID)
			
			// Remove from Redis
			usm.redisClient.Del(ctx, key)
		}
	}

	return nil
}

// CalculateUploadProgress calculates the completion percentage
func (state *UploadState) CalculateProgress() float64 {
	if state.TotalSize == 0 {
		return 0.0
	}
	return (float64(state.UploadedSize) / float64(state.TotalSize)) * 100.0
}

// IsCompleted checks if all parts have been uploaded
func (state *UploadState) IsCompleted() bool {
	if state.TotalSize == 0 || len(state.Parts) == 0 {
		return false
	}
	return state.UploadedSize >= state.TotalSize
}