package integration

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// MinIO integration test suite
type MinIOTestSuite struct {
	suite.Suite
	minioEndpoint string
	accessKey     string
	secretKey     string
	bucketName    string
	client        *http.Client
}

// Setup test suite
func (suite *MinIOTestSuite) SetupSuite() {
	suite.minioEndpoint = "localhost:9000"
	suite.accessKey = "minio"
	suite.secretKey = "minio123"
	suite.bucketName = "terraallwert"
	suite.client = &http.Client{Timeout: 30 * time.Second}
	
	// Wait for MinIO to be ready
	suite.waitForMinIO()
	
	// Ensure bucket exists
	suite.ensureBucketExists()
}

func (suite *MinIOTestSuite) TearDownSuite() {
	// Cleanup test files
	suite.cleanupTestFiles()
}

// Wait for MinIO to be available
func (suite *MinIOTestSuite) waitForMinIO() {
	maxAttempts := 30
	for i := 0; i < maxAttempts; i++ {
		if suite.isMinIOReady() {
			return
		}
		time.Sleep(1 * time.Second)
	}
	suite.T().Fatal("MinIO not available after 30 seconds")
}

func (suite *MinIOTestSuite) isMinIOReady() bool {
	url := fmt.Sprintf("http://%s/minio/health/live", suite.minioEndpoint)
	resp, err := suite.client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode == http.StatusOK
}

// Ensure test bucket exists
func (suite *MinIOTestSuite) ensureBucketExists() {
	// In a real implementation, this would use MinIO client to create bucket
	// For now, we assume bucket exists from docker-compose setup
	suite.T().Logf("Using bucket: %s", suite.bucketName)
}

// Test file upload to MinIO using presigned URL
func (suite *MinIOTestSuite) TestUploadFileWithPresignedURL() {
	// This test simulates the complete upload flow:
	// 1. Generate presigned upload URL (mocked)
	// 2. Upload file using the URL
	// 3. Verify file exists
	// 4. Download file and verify content
	
	// Step 1: Simulate presigned URL generation
	objectKey := suite.generateTestObjectKey("image", "test-upload.jpg")
	presignedUploadURL := suite.generateMockPresignedURL(objectKey, "PUT")
	
	// Step 2: Create test file content
	testContent := []byte("Terra Allwert API - Test Image Content\nTimestamp: " + time.Now().Format(time.RFC3339))
	contentType := "image/jpeg"
	
	// Step 3: Upload file using presigned URL
	err := suite.uploadFileWithPresignedURL(presignedUploadURL, testContent, contentType)
	suite.Assert().NoError(err, "File upload should succeed")
	
	// Step 4: Verify file exists by attempting download
	presignedDownloadURL := suite.generateMockPresignedURL(objectKey, "GET")
	downloadedContent, err := suite.downloadFileWithPresignedURL(presignedDownloadURL)
	suite.Assert().NoError(err, "File download should succeed")
	
	// Step 5: Verify content integrity
	suite.Assert().Equal(testContent, downloadedContent, "Downloaded content should match uploaded content")
	
	// Step 6: Cleanup
	suite.deleteTestFile(objectKey)
}

func (suite *MinIOTestSuite) TestUploadMultipleFiles() {
	testFiles := []struct {
		name        string
		content     string
		contentType string
		fileType    string
	}{
		{
			name:        "test-image-1.jpg",
			content:     "Image file content 1",
			contentType: "image/jpeg",
			fileType:    "image",
		},
		{
			name:        "test-image-2.png",
			content:     "Image file content 2",
			contentType: "image/png",
			fileType:    "image",
		},
		{
			name:        "test-video.mp4",
			content:     "Video file content",
			contentType: "video/mp4",
			fileType:    "video",
		},
		{
			name:        "test-document.pdf",
			content:     "Document file content",
			contentType: "application/pdf",
			fileType:    "document",
		},
	}
	
	uploadedFiles := make([]string, 0, len(testFiles))
	
	// Upload all files
	for _, file := range testFiles {
		objectKey := suite.generateTestObjectKey(file.fileType, file.name)
		presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
		
		err := suite.uploadFileWithPresignedURL(presignedURL, []byte(file.content), file.contentType)
		suite.Assert().NoError(err, "Upload should succeed for file: %s", file.name)
		
		uploadedFiles = append(uploadedFiles, objectKey)
	}
	
	// Verify all files exist and have correct content
	for i, objectKey := range uploadedFiles {
		presignedDownloadURL := suite.generateMockPresignedURL(objectKey, "GET")
		content, err := suite.downloadFileWithPresignedURL(presignedDownloadURL)
		
		suite.Assert().NoError(err, "Download should succeed for file: %s", testFiles[i].name)
		suite.Assert().Equal([]byte(testFiles[i].content), content, "Content should match for file: %s", testFiles[i].name)
	}
	
	// Cleanup all files
	for _, objectKey := range uploadedFiles {
		suite.deleteTestFile(objectKey)
	}
}

func (suite *MinIOTestSuite) TestUploadLargeFile() {
	// Test uploading a larger file (1MB)
	objectKey := suite.generateTestObjectKey("document", "large-test-file.bin")
	presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
	
	// Generate 1MB of test data
	largeContent := make([]byte, 1024*1024) // 1MB
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}
	
	// Upload large file
	err := suite.uploadFileWithPresignedURL(presignedURL, largeContent, "application/octet-stream")
	suite.Assert().NoError(err, "Large file upload should succeed")
	
	// Verify file exists and has correct size
	presignedDownloadURL := suite.generateMockPresignedURL(objectKey, "GET")
	downloadedContent, err := suite.downloadFileWithPresignedURL(presignedDownloadURL)
	suite.Assert().NoError(err, "Large file download should succeed")
	suite.Assert().Equal(len(largeContent), len(downloadedContent), "Downloaded file size should match")
	suite.Assert().Equal(largeContent, downloadedContent, "Downloaded content should match")
	
	// Cleanup
	suite.deleteTestFile(objectKey)
}

func (suite *MinIOTestSuite) TestUploadWithInvalidPresignedURL() {
	// Test with expired or invalid presigned URL
	invalidURL := "http://invalid-minio-url:9000/bucket/object?signature=invalid"
	testContent := []byte("Test content")
	
	err := suite.uploadFileWithPresignedURL(invalidURL, testContent, "text/plain")
	suite.Assert().Error(err, "Upload with invalid URL should fail")
}

func (suite *MinIOTestSuite) TestDownloadNonExistentFile() {
	// Test downloading a file that doesn't exist
	nonExistentKey := suite.generateTestObjectKey("image", "non-existent-file.jpg")
	presignedDownloadURL := suite.generateMockPresignedURL(nonExistentKey, "GET")
	
	content, err := suite.downloadFileWithPresignedURL(presignedDownloadURL)
	suite.Assert().Error(err, "Download of non-existent file should fail")
	suite.Assert().Nil(content, "Content should be nil for non-existent file")
}

func (suite *MinIOTestSuite) TestConcurrentUploads() {
	// Test concurrent uploads to ensure thread safety
	const concurrentUploads = 5
	results := make(chan error, concurrentUploads)
	objectKeys := make([]string, concurrentUploads)
	
	// Start concurrent uploads
	for i := 0; i < concurrentUploads; i++ {
		go func(index int) {
			objectKey := suite.generateTestObjectKey("image", fmt.Sprintf("concurrent-test-%d.jpg", index))
			objectKeys[index] = objectKey
			presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
			
			content := []byte(fmt.Sprintf("Concurrent upload test content %d", index))
			err := suite.uploadFileWithPresignedURL(presignedURL, content, "image/jpeg")
			results <- err
		}(i)
	}
	
	// Wait for all uploads to complete
	for i := 0; i < concurrentUploads; i++ {
		err := <-results
		suite.Assert().NoError(err, "Concurrent upload %d should succeed", i)
	}
	
	// Verify all files were uploaded correctly
	for i, objectKey := range objectKeys {
		if objectKey != "" { // Skip if objectKey wasn't set due to goroutine timing
			presignedDownloadURL := suite.generateMockPresignedURL(objectKey, "GET")
			content, err := suite.downloadFileWithPresignedURL(presignedDownloadURL)
			
			suite.Assert().NoError(err, "Concurrent download %d should succeed", i)
			expectedContent := fmt.Sprintf("Concurrent upload test content %d", i)
			suite.Assert().Equal([]byte(expectedContent), content, "Concurrent file %d content should match", i)
			
			// Cleanup
			suite.deleteTestFile(objectKey)
		}
	}
}

// Helper methods
func (suite *MinIOTestSuite) generateTestObjectKey(fileType, fileName string) string {
	timestamp := time.Now().Format("20060102-150405")
	routeID := "test-route"
	fileID := fmt.Sprintf("file-%s-%d", timestamp, time.Now().UnixNano())
	
	return fmt.Sprintf("%s/%s/%s/%s", routeID, fileType, fileID, fileName)
}

func (suite *MinIOTestSuite) generateMockPresignedURL(objectKey, method string) string {
	// In a real implementation, this would call MinIO SDK to generate actual presigned URLs
	// For testing purposes, we simulate the presigned URL structure
	baseURL := fmt.Sprintf("http://%s/%s/%s", suite.minioEndpoint, suite.bucketName, objectKey)
	
	// Add mock presigned parameters
	timestamp := time.Now().Unix()
	expires := timestamp + 3600 // 1 hour expiration
	
	return fmt.Sprintf("%s?X-Amz-Algorithm=AWS4-HMAC-SHA256&X-Amz-Expires=3600&X-Amz-SignedHeaders=host&X-Amz-Timestamp=%d", baseURL, timestamp)
}

func (suite *MinIOTestSuite) uploadFileWithPresignedURL(presignedURL string, content []byte, contentType string) error {
	req, err := http.NewRequest("PUT", presignedURL, bytes.NewReader(content))
	if err != nil {
		return fmt.Errorf("failed to create upload request: %w", err)
	}
	
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(content))
	
	resp, err := suite.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to upload file: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(body))
	}
	
	return nil
}

func (suite *MinIOTestSuite) downloadFileWithPresignedURL(presignedURL string) ([]byte, error) {
	resp, err := suite.client.Get(presignedURL)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("download failed with status %d", resp.StatusCode)
	}
	
	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read download response: %w", err)
	}
	
	return content, nil
}

func (suite *MinIOTestSuite) deleteTestFile(objectKey string) {
	// In a real implementation, this would use MinIO client to delete the object
	// For now, we'll make a DELETE request to the object URL
	deleteURL := fmt.Sprintf("http://%s/%s/%s", suite.minioEndpoint, suite.bucketName, objectKey)
	
	req, err := http.NewRequest("DELETE", deleteURL, nil)
	if err != nil {
		suite.T().Logf("Failed to create delete request for %s: %v", objectKey, err)
		return
	}
	
	resp, err := suite.client.Do(req)
	if err != nil {
		suite.T().Logf("Failed to delete file %s: %v", objectKey, err)
		return
	}
	defer resp.Body.Close()
	
	if resp.StatusCode != http.StatusNoContent && resp.StatusCode != http.StatusOK {
		suite.T().Logf("Delete request for %s returned status %d", objectKey, resp.StatusCode)
	}
}

func (suite *MinIOTestSuite) cleanupTestFiles() {
	// Clean up any remaining test files
	// In a real implementation, this would list and delete all test objects
	suite.T().Log("Cleanup completed")
}

// Benchmark tests for MinIO operations
func (suite *MinIOTestSuite) TestMinIOPerformance() {
	// Test upload performance
	objectKey := suite.generateTestObjectKey("performance", "perf-test.bin")
	presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
	
	// Test with different file sizes
	fileSizes := []int{1024, 10240, 102400, 1048576} // 1KB, 10KB, 100KB, 1MB
	
	for _, size := range fileSizes {
		suite.T().Run(fmt.Sprintf("Upload_%dB", size), func(t *testing.T) {
			content := make([]byte, size)
			for i := range content {
				content[i] = byte(i % 256)
			}
			
			start := time.Now()
			err := suite.uploadFileWithPresignedURL(presignedURL, content, "application/octet-stream")
			duration := time.Since(start)
			
			assert.NoError(t, err)
			t.Logf("Upload of %d bytes took %v (%.2f MB/s)", size, duration, float64(size)/(1024*1024)/duration.Seconds())
			
			// Cleanup
			suite.deleteTestFile(objectKey)
		})
	}
}

// Test with real MinIO operations (if available)
func (suite *MinIOTestSuite) TestRealMinIOIntegration() {
	// This test only runs if MinIO is actually available and responding
	if !suite.isMinIOReady() {
		suite.T().Skip("MinIO not available, skipping real integration test")
	}
	
	// Test basic connectivity
	url := fmt.Sprintf("http://%s/minio/health/ready", suite.minioEndpoint)
	resp, err := suite.client.Get(url)
	suite.Assert().NoError(err, "Should be able to connect to MinIO health endpoint")
	
	if resp != nil {
		defer resp.Body.Close()
		suite.Assert().True(resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusServiceUnavailable, 
			"MinIO health endpoint should respond with 200 or 503")
	}
	
	// Test bucket accessibility
	bucketURL := fmt.Sprintf("http://%s/%s/", suite.minioEndpoint, suite.bucketName)
	resp, err = suite.client.Head(bucketURL)
	if err == nil && resp != nil {
		defer resp.Body.Close()
		suite.T().Logf("Bucket %s accessibility status: %d", suite.bucketName, resp.StatusCode)
	}
}

// Run the MinIO test suite
func TestMinIOTestSuite(t *testing.T) {
	suite.Run(t, new(MinIOTestSuite))
}

// Separate benchmark tests
func BenchmarkMinIOUpload_1KB(b *testing.B) {
	suite := &MinIOTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	content := make([]byte, 1024) // 1KB
	for i := range content {
		content[i] = byte(i % 256)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		objectKey := suite.generateTestObjectKey("benchmark", fmt.Sprintf("bench-1kb-%d.bin", i))
		presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
		
		err := suite.uploadFileWithPresignedURL(presignedURL, content, "application/octet-stream")
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}
		
		// Cleanup
		suite.deleteTestFile(objectKey)
	}
}

func BenchmarkMinIOUpload_100KB(b *testing.B) {
	suite := &MinIOTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	content := make([]byte, 102400) // 100KB
	for i := range content {
		content[i] = byte(i % 256)
	}
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		objectKey := suite.generateTestObjectKey("benchmark", fmt.Sprintf("bench-100kb-%d.bin", i))
		presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
		
		err := suite.uploadFileWithPresignedURL(presignedURL, content, "application/octet-stream")
		if err != nil {
			b.Fatalf("Upload failed: %v", err)
		}
		
		// Cleanup
		suite.deleteTestFile(objectKey)
	}
}

// Table-driven test for various file types and scenarios
func (suite *MinIOTestSuite) TestFileUploadScenarios() {
	scenarios := []struct {
		name        string
		fileName    string
		content     string
		contentType string
		fileType    string
		expectError bool
	}{
		{
			name:        "Valid JPEG image",
			fileName:    "test.jpg",
			content:     "JPEG image content",
			contentType: "image/jpeg",
			fileType:    "image",
			expectError: false,
		},
		{
			name:        "Valid PNG image",
			fileName:    "test.png",
			content:     "PNG image content",
			contentType: "image/png",
			fileType:    "image",
			expectError: false,
		},
		{
			name:        "Valid MP4 video",
			fileName:    "test.mp4",
			content:     "MP4 video content",
			contentType: "video/mp4",
			fileType:    "video",
			expectError: false,
		},
		{
			name:        "Valid PDF document",
			fileName:    "test.pdf",
			content:     "PDF document content",
			contentType: "application/pdf",
			fileType:    "document",
			expectError: false,
		},
		{
			name:        "Empty file",
			fileName:    "empty.txt",
			content:     "",
			contentType: "text/plain",
			fileType:    "document",
			expectError: false,
		},
		{
			name:        "Long filename",
			fileName:    strings.Repeat("a", 200) + ".txt",
			content:     "File with very long name",
			contentType: "text/plain",
			fileType:    "document",
			expectError: false,
		},
	}
	
	for _, scenario := range scenarios {
		suite.T().Run(scenario.name, func(t *testing.T) {
			objectKey := suite.generateTestObjectKey(scenario.fileType, scenario.fileName)
			presignedURL := suite.generateMockPresignedURL(objectKey, "PUT")
			
			err := suite.uploadFileWithPresignedURL(presignedURL, []byte(scenario.content), scenario.contentType)
			
			if scenario.expectError {
				assert.Error(t, err, "Upload should fail for scenario: %s", scenario.name)
			} else {
				assert.NoError(t, err, "Upload should succeed for scenario: %s", scenario.name)
				
				// Verify file can be downloaded
				presignedDownloadURL := suite.generateMockPresignedURL(objectKey, "GET")
				downloadedContent, err := suite.downloadFileWithPresignedURL(presignedDownloadURL)
				assert.NoError(t, err, "Download should succeed for scenario: %s", scenario.name)
				assert.Equal(t, []byte(scenario.content), downloadedContent, "Content should match for scenario: %s", scenario.name)
				
				// Cleanup
				suite.deleteTestFile(objectKey)
			}
		})
	}
}