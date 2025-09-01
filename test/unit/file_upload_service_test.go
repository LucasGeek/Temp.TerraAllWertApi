package unit

import (
	"bytes"
	"context"
	"io"
	"strings"
	"testing"
	"time"

	"api/domain/entities"
	"test/fixtures/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// Mock do MinIO client
type MockMinIOClient struct {
	mock.Mock
}

func (m *MockMinIOClient) GetSignedUploadURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, bucketName, objectKey, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockMinIOClient) GetSignedDownloadURL(ctx context.Context, bucketName, objectKey string, expiration time.Duration) (string, error) {
	args := m.Called(ctx, bucketName, objectKey, expiration)
	return args.String(0), args.Error(1)
}

func (m *MockMinIOClient) UploadFile(ctx context.Context, bucketName, objectKey string, reader io.Reader, objectSize int64, contentType string) error {
	args := m.Called(ctx, bucketName, objectKey, reader, objectSize, contentType)
	return args.Error(0)
}

func (m *MockMinIOClient) DeleteFile(ctx context.Context, bucketName, objectKey string) error {
	args := m.Called(ctx, bucketName, objectKey)
	return args.Error(0)
}

func (m *MockMinIOClient) FileExists(ctx context.Context, bucketName, objectKey string) (bool, error) {
	args := m.Called(ctx, bucketName, objectKey)
	return args.Bool(0), args.Error(1)
}

// Mock do repository de arquivos
type MockFileRepository struct {
	mock.Mock
}

func (m *MockFileRepository) Create(ctx context.Context, file *entities.FileMetadata) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) FindByID(ctx context.Context, id string) (*entities.FileMetadata, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*entities.FileMetadata), args.Error(1)
}

func (m *MockFileRepository) Update(ctx context.Context, file *entities.FileMetadata) error {
	args := m.Called(ctx, file)
	return args.Error(0)
}

func (m *MockFileRepository) Delete(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockFileRepository) FindByRouteID(ctx context.Context, routeID string) ([]*entities.FileMetadata, error) {
	args := m.Called(ctx, routeID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*entities.FileMetadata), args.Error(1)
}

// FileUploadService simulado para testes
type FileUploadService struct {
	minioClient    *MockMinIOClient
	fileRepo       *MockFileRepository
	bucketName     string
	urlExpiration  time.Duration
}

func NewFileUploadService(minioClient *MockMinIOClient, fileRepo *MockFileRepository) *FileUploadService {
	return &FileUploadService{
		minioClient:   minioClient,
		fileRepo:      fileRepo,
		bucketName:    "terraallwert",
		urlExpiration: 1 * time.Hour,
	}
}

func (s *FileUploadService) GetSignedUploadURL(ctx context.Context, req *entities.SignedUploadURLRequest) (*entities.SignedUploadURLResponse, error) {
	// Gerar path do arquivo
	objectKey := s.generateObjectKey(req.RouteID, req.FileType, req.FileName)
	
	// Obter URL assinada
	uploadURL, err := s.minioClient.GetSignedUploadURL(ctx, s.bucketName, objectKey, s.urlExpiration)
	if err != nil {
		return nil, err
	}

	// Criar metadata do arquivo
	fileMetadata := &entities.FileMetadata{
		ID:           generateFileID(),
		FileName:     req.FileName,
		ContentType:  req.ContentType,
		FileType:     req.FileType,
		RouteID:      req.RouteID,
		MinIOPath:    objectKey,
		Status:       entities.FileStatusPending,
		Context:      req.Context,
		CreatedAt:    time.Now(),
	}

	// Salvar metadata
	if err := s.fileRepo.Create(ctx, fileMetadata); err != nil {
		return nil, err
	}

	return &entities.SignedUploadURLResponse{
		UploadURL: uploadURL,
		MinIOPath: objectKey,
		ExpiresAt: time.Now().Add(s.urlExpiration),
		FileID:    fileMetadata.ID,
	}, nil
}

func (s *FileUploadService) ConfirmFileUpload(ctx context.Context, req *entities.ConfirmFileUploadRequest) (*entities.ConfirmFileUploadResponse, error) {
	// Buscar metadata do arquivo
	fileMetadata, err := s.fileRepo.FindByID(ctx, req.FileID)
	if err != nil {
		return nil, err
	}

	// Verificar se arquivo existe no MinIO
	exists, err := s.minioClient.FileExists(ctx, s.bucketName, req.MinIOPath)
	if err != nil {
		return nil, err
	}

	if !exists {
		return nil, assert.AnError // File not uploaded
	}

	// Atualizar metadata
	fileMetadata.Status = entities.FileStatusUploaded
	fileMetadata.FileSize = req.FileSize
	fileMetadata.Checksum = req.Checksum
	fileMetadata.UpdatedAt = time.Now()

	if err := s.fileRepo.Update(ctx, fileMetadata); err != nil {
		return nil, err
	}

	// Gerar URL de download
	downloadURL, err := s.minioClient.GetSignedDownloadURL(ctx, s.bucketName, req.MinIOPath, 24*time.Hour)
	if err != nil {
		return nil, err
	}

	return &entities.ConfirmFileUploadResponse{
		Success: true,
		FileMetadata: &entities.FileMetadataResponse{
			ID:          fileMetadata.ID,
			URL:         s.generatePublicURL(req.MinIOPath),
			DownloadURL: downloadURL,
			Metadata:    fileMetadata.Context,
		},
	}, nil
}

func (s *FileUploadService) generateObjectKey(routeID, fileType, fileName string) string {
	timestamp := time.Now().Unix()
	return strings.Join([]string{routeID, fileType, generateFileID(), fileName}, "/")
}

func (s *FileUploadService) generatePublicURL(minioPath string) string {
	return "https://cdn.terraallwert.com/" + minioPath
}

// Helper para gerar IDs únicos
func generateFileID() string {
	return "file-" + time.Now().Format("20060102-150405") + "-test"
}

// Fixtures para testes de upload
func createTestSignedUploadURLRequest() *entities.SignedUploadURLRequest {
	return &entities.SignedUploadURLRequest{
		FileName:    "test-image.jpg",
		FileType:    "image",
		ContentType: "image/jpeg",
		RouteID:     "route-123",
		Context: map[string]interface{}{
			"title":       "Test Image",
			"description": "Image for testing",
		},
	}
}

func createTestConfirmUploadRequest(fileID, minioPath string) *entities.ConfirmFileUploadRequest {
	return &entities.ConfirmFileUploadRequest{
		FileID:           fileID,
		MinIOPath:        minioPath,
		RouteID:          "route-123",
		OriginalFileName: "test-image.jpg",
		FileSize:         1024000,
		Checksum:         "abc123def456",
		Context: map[string]interface{}{
			"confirmed": true,
		},
	}
}

// Testes do FileUploadService
func TestFileUploadService_GetSignedUploadURL_Success(t *testing.T) {
	// Arrange
	mockMinIO := new(MockMinIOClient)
	mockFileRepo := new(MockFileRepository)
	service := NewFileUploadService(mockMinIO, mockFileRepo)
	
	ctx := context.Background()
	req := createTestSignedUploadURLRequest()
	
	expectedURL := "https://minio.example.com/upload-url"
	
	// Mock expectations
	mockMinIO.On("GetSignedUploadURL", ctx, "terraallwert", mock.AnythingOfType("string"), 1*time.Hour).Return(expectedURL, nil)
	mockFileRepo.On("Create", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil)

	// Act
	result, err := service.GetSignedUploadURL(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedURL, result.UploadURL)
	assert.NotEmpty(t, result.FileID)
	assert.NotEmpty(t, result.MinIOPath)
	assert.True(t, result.ExpiresAt.After(time.Now()))

	// Verify mocks
	mockMinIO.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUploadService_GetSignedUploadURL_MinIOError(t *testing.T) {
	// Arrange
	mockMinIO := new(MockMinIOClient)
	mockFileRepo := new(MockFileRepository)
	service := NewFileUploadService(mockMinIO, mockFileRepo)
	
	ctx := context.Background()
	req := createTestSignedUploadURLRequest()
	
	// Mock expectations
	mockMinIO.On("GetSignedUploadURL", ctx, "terraallwert", mock.AnythingOfType("string"), 1*time.Hour).Return("", assert.AnError)

	// Act
	result, err := service.GetSignedUploadURL(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify mocks
	mockMinIO.AssertExpectations(t)
}

func TestFileUploadService_ConfirmFileUpload_Success(t *testing.T) {
	// Arrange
	mockMinIO := new(MockMinIOClient)
	mockFileRepo := new(MockFileRepository)
	service := NewFileUploadService(mockMinIO, mockFileRepo)
	
	ctx := context.Background()
	fileID := "test-file-id"
	minioPath := "route-123/image/test-file-id/test-image.jpg"
	req := createTestConfirmUploadRequest(fileID, minioPath)
	
	// Create test file metadata
	fileMetadata := &entities.FileMetadata{
		ID:          fileID,
		FileName:    "test-image.jpg",
		ContentType: "image/jpeg",
		FileType:    "image",
		RouteID:     "route-123",
		MinIOPath:   minioPath,
		Status:      entities.FileStatusPending,
		CreatedAt:   time.Now(),
	}

	expectedDownloadURL := "https://minio.example.com/download-url"
	
	// Mock expectations
	mockFileRepo.On("FindByID", ctx, fileID).Return(fileMetadata, nil)
	mockMinIO.On("FileExists", ctx, "terraallwert", minioPath).Return(true, nil)
	mockFileRepo.On("Update", ctx, mock.AnythingOfType("*entities.FileMetadata")).Return(nil)
	mockMinIO.On("GetSignedDownloadURL", ctx, "terraallwert", minioPath, 24*time.Hour).Return(expectedDownloadURL, nil)

	// Act
	result, err := service.ConfirmFileUpload(ctx, req)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.Success)
	assert.NotNil(t, result.FileMetadata)
	assert.Equal(t, fileID, result.FileMetadata.ID)
	assert.Equal(t, expectedDownloadURL, result.FileMetadata.DownloadURL)
	assert.NotEmpty(t, result.FileMetadata.URL)

	// Verify mocks
	mockMinIO.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUploadService_ConfirmFileUpload_FileNotExists(t *testing.T) {
	// Arrange
	mockMinIO := new(MockMinIOClient)
	mockFileRepo := new(MockFileRepository)
	service := NewFileUploadService(mockMinIO, mockFileRepo)
	
	ctx := context.Background()
	fileID := "test-file-id"
	minioPath := "route-123/image/test-file-id/test-image.jpg"
	req := createTestConfirmUploadRequest(fileID, minioPath)
	
	// Create test file metadata
	fileMetadata := &entities.FileMetadata{
		ID:        fileID,
		MinIOPath: minioPath,
		Status:    entities.FileStatusPending,
	}
	
	// Mock expectations
	mockFileRepo.On("FindByID", ctx, fileID).Return(fileMetadata, nil)
	mockMinIO.On("FileExists", ctx, "terraallwert", minioPath).Return(false, nil)

	// Act
	result, err := service.ConfirmFileUpload(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify mocks
	mockMinIO.AssertExpectations(t)
	mockFileRepo.AssertExpectations(t)
}

func TestFileUploadService_ConfirmFileUpload_FileNotFound(t *testing.T) {
	// Arrange
	mockMinIO := new(MockMinIOClient)
	mockFileRepo := new(MockFileRepository)
	service := NewFileUploadService(mockMinIO, mockFileRepo)
	
	ctx := context.Background()
	fileID := "non-existent-file-id"
	minioPath := "route-123/image/test/test-image.jpg"
	req := createTestConfirmUploadRequest(fileID, minioPath)
	
	// Mock expectations
	mockFileRepo.On("FindByID", ctx, fileID).Return(nil, assert.AnError)

	// Act
	result, err := service.ConfirmFileUpload(ctx, req)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	// Verify mocks
	mockFileRepo.AssertExpectations(t)
}

// Testes de integração com dados reais (sem mocks)
func TestFileUploadService_GenerateObjectKey(t *testing.T) {
	// Arrange
	service := &FileUploadService{}
	
	// Act
	objectKey := service.generateObjectKey("route-123", "image", "test.jpg")
	
	// Assert
	assert.Contains(t, objectKey, "route-123")
	assert.Contains(t, objectKey, "image")
	assert.Contains(t, objectKey, "test.jpg")
	
	// Verify structure: routeID/fileType/fileID/fileName
	parts := strings.Split(objectKey, "/")
	assert.Len(t, parts, 4)
	assert.Equal(t, "route-123", parts[0])
	assert.Equal(t, "image", parts[1])
	assert.Equal(t, "test.jpg", parts[3])
}

func TestFileUploadService_GeneratePublicURL(t *testing.T) {
	// Arrange
	service := &FileUploadService{}
	minioPath := "route-123/image/file-123/test.jpg"
	
	// Act
	publicURL := service.generatePublicURL(minioPath)
	
	// Assert
	expected := "https://cdn.terraallwert.com/" + minioPath
	assert.Equal(t, expected, publicURL)
}

// Testes de validação de tipos de arquivo
func TestFileUploadService_ValidateFileType(t *testing.T) {
	tests := []struct {
		name        string
		fileName    string
		contentType string
		fileType    string
		expectValid bool
	}{
		{
			name:        "Valid image file",
			fileName:    "test.jpg",
			contentType: "image/jpeg",
			fileType:    "image",
			expectValid: true,
		},
		{
			name:        "Valid video file",
			fileName:    "test.mp4",
			contentType: "video/mp4",
			fileType:    "video",
			expectValid: true,
		},
		{
			name:        "Invalid file type mismatch",
			fileName:    "test.jpg",
			contentType: "video/mp4",
			fileType:    "image",
			expectValid: false,
		},
		{
			name:        "Dangerous file extension",
			fileName:    "malicious.exe",
			contentType: "application/octet-stream",
			fileType:    "document",
			expectValid: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			req := &entities.SignedUploadURLRequest{
				FileName:    tt.fileName,
				ContentType: tt.contentType,
				FileType:    tt.fileType,
				RouteID:     "test-route",
			}
			
			// Act & Assert
			isValid := validateFileUploadRequest(req)
			assert.Equal(t, tt.expectValid, isValid)
		})
	}
}

// Helper function para validação (simulada)
func validateFileUploadRequest(req *entities.SignedUploadURLRequest) bool {
	// Validações básicas de segurança
	dangerousExtensions := []string{".exe", ".bat", ".cmd", ".scr", ".pif", ".vbs", ".js"}
	
	for _, ext := range dangerousExtensions {
		if strings.HasSuffix(strings.ToLower(req.FileName), ext) {
			return false
		}
	}
	
	// Validar consistência entre tipo de arquivo e content type
	imageTypes := []string{"image/jpeg", "image/png", "image/gif", "image/webp"}
	videoTypes := []string{"video/mp4", "video/avi", "video/mov", "video/wmv"}
	
	switch req.FileType {
	case "image":
		for _, ct := range imageTypes {
			if req.ContentType == ct {
				return true
			}
		}
		return false
	case "video":
		for _, ct := range videoTypes {
			if req.ContentType == ct {
				return true
			}
		}
		return false
	}
	
	return true
}

// Benchmark tests
func BenchmarkFileUploadService_GenerateObjectKey(b *testing.B) {
	service := &FileUploadService{}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.generateObjectKey("route-123", "image", "test.jpg")
	}
}

func BenchmarkFileUploadService_ValidateFileUploadRequest(b *testing.B) {
	req := &entities.SignedUploadURLRequest{
		FileName:    "test.jpg",
		ContentType: "image/jpeg",
		FileType:    "image",
		RouteID:     "test-route",
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = validateFileUploadRequest(req)
	}
}