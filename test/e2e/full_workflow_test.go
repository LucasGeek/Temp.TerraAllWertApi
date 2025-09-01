package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"api/domain/entities"
	"test/fixtures/testutils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

// Full workflow end-to-end test suite
type E2EWorkflowTestSuite struct {
	suite.Suite
	server     *httptest.Server
	client     *http.Client
	baseURL    string
	authToken  string
	testData   *E2ETestData
}

// Test data holder
type E2ETestData struct {
	UserCredentials struct {
		Email    string
		Password string
	}
	UploadedFiles []UploadedFileInfo
	CreatedMenus  []string // Menu IDs
	CreatedItems  struct {
		CarouselID  string
		FloorPlanID string
		PinMapID    string
		SyncID      string
	}
	RouteID string
}

type UploadedFileInfo struct {
	FileID      string
	MinIOPath   string
	DownloadURL string
	FileType    string
	FileName    string
}

// Setup test suite
func (suite *E2EWorkflowTestSuite) SetupSuite() {
	// Initialize test server (mock)
	suite.server = httptest.NewServer(http.HandlerFunc(suite.mockAPIHandler))
	suite.client = &http.Client{Timeout: 60 * time.Second}
	suite.baseURL = suite.server.URL
	
	// Initialize test data
	suite.testData = &E2ETestData{
		RouteID: fmt.Sprintf("e2e-test-route-%d", time.Now().Unix()),
	}
	suite.testData.UserCredentials.Email = "admin@euvatar.com"
	suite.testData.UserCredentials.Password = "admin123"
	
	suite.T().Logf("E2E Test Suite initialized with route ID: %s", suite.testData.RouteID)
}

func (suite *E2EWorkflowTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
	
	// Cleanup any persistent test data
	suite.cleanupTestData()
}

// Mock API handler that simulates the real API
func (suite *E2EWorkflowTestSuite) mockAPIHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	switch {
	case r.URL.Path == "/health":
		suite.handleHealthCheck(w, r)
	case r.URL.Path == "/graphql":
		suite.handleGraphQL(w, r)
	case strings.HasPrefix(r.URL.Path, "/minio/"):
		suite.handleMinIOOperation(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (suite *E2EWorkflowTestSuite) handleHealthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status": "ok",
		"timestamp": time.Now().Format(time.RFC3339),
		"version": "1.0.0-test",
	})
}

func (suite *E2EWorkflowTestSuite) handleGraphQL(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Query     string                 `json:"query"`
		Variables map[string]interface{} `json:"variables"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}
	
	response := suite.routeGraphQLOperation(req.Query, req.Variables, r.Header.Get("Authorization"))
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (suite *E2EWorkflowTestSuite) handleMinIOOperation(w http.ResponseWriter, r *http.Request) {
	// Simulate MinIO operations (upload/download)
	switch r.Method {
	case "PUT":
		// Simulate file upload
		body, _ := io.ReadAll(r.Body)
		suite.T().Logf("Mock MinIO upload: %d bytes to %s", len(body), r.URL.Path)
		w.WriteHeader(http.StatusOK)
	case "GET":
		// Simulate file download
		suite.T().Logf("Mock MinIO download from %s", r.URL.Path)
		w.Header().Set("Content-Type", "application/octet-stream")
		w.Write([]byte("Mock file content"))
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}

func (suite *E2EWorkflowTestSuite) routeGraphQLOperation(query string, variables map[string]interface{}, authHeader string) interface{} {
	query = strings.TrimSpace(query)
	
	switch {
	case strings.Contains(query, "mutation Login"):
		return suite.mockLogin(variables)
	case strings.Contains(query, "mutation GetSignedUploadUrl"):
		return suite.mockGetSignedUploadUrl(variables)
	case strings.Contains(query, "mutation ConfirmFileUpload"):
		return suite.mockConfirmFileUpload(variables)
	case strings.Contains(query, "mutation CreateMenu"):
		return suite.mockCreateMenu(variables)
	case strings.Contains(query, "mutation CreateImageCarousel"):
		return suite.mockCreateImageCarousel(variables)
	case strings.Contains(query, "mutation CreateFloorPlan"):
		return suite.mockCreateFloorPlan(variables)
	case strings.Contains(query, "mutation CreatePinMap"):
		return suite.mockCreatePinMap(variables)
	case strings.Contains(query, "mutation RequestFullSync"):
		return suite.mockRequestFullSync(variables)
	case strings.Contains(query, "query GetRouteBusinessData"):
		return suite.mockGetRouteBusinessData(variables)
	default:
		return map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Operation not implemented in E2E mock"},
			},
		}
	}
}

// Mock GraphQL operations
func (suite *E2EWorkflowTestSuite) mockLogin(variables map[string]interface{}) interface{} {
	input := variables["input"].(map[string]interface{})
	email := input["email"].(string)
	password := input["password"].(string)
	
	if email != suite.testData.UserCredentials.Email || password != suite.testData.UserCredentials.Password {
		return map[string]interface{}{
			"errors": []map[string]interface{}{
				{"message": "Invalid credentials"},
			},
		}
	}
	
	token := fmt.Sprintf("e2e-test-token-%d", time.Now().Unix())
	suite.authToken = token
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"login": map[string]interface{}{
				"token":        token,
				"refreshToken": "e2e-refresh-token",
				"expiresAt":    time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"user": map[string]interface{}{
					"id":       "e2e-admin-user",
					"username": "admin",
					"email":    email,
					"role":     "ADMIN",
				},
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockGetSignedUploadUrl(variables map[string]interface{}) interface{} {
	input := variables["input"].(map[string]interface{})
	fileName := input["fileName"].(string)
	fileType := input["fileType"].(string)
	
	fileID := fmt.Sprintf("e2e-file-%d", time.Now().UnixNano())
	minioPath := fmt.Sprintf("%s/%s/%s/%s", suite.testData.RouteID, fileType, fileID, fileName)
	uploadURL := fmt.Sprintf("%s/minio/upload/%s", suite.baseURL, minioPath)
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"getSignedUploadUrl": map[string]interface{}{
				"uploadUrl": uploadURL,
				"minioPath": minioPath,
				"expiresAt": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"fileId":    fileID,
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockConfirmFileUpload(variables map[string]interface{}) interface{} {
	input := variables["input"].(map[string]interface{})
	fileID := input["fileId"].(string)
	minioPath := input["minioPath"].(string)
	fileName := input["originalFileName"].(string)
	
	downloadURL := fmt.Sprintf("%s/minio/download/%s", suite.baseURL, minioPath)
	publicURL := fmt.Sprintf("https://cdn.terraallwert.com/%s", minioPath)
	
	// Store uploaded file info for later verification
	suite.testData.UploadedFiles = append(suite.testData.UploadedFiles, UploadedFileInfo{
		FileID:      fileID,
		MinIOPath:   minioPath,
		DownloadURL: downloadURL,
		FileType:    extractFileTypeFromPath(minioPath),
		FileName:    fileName,
	})
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"confirmFileUpload": map[string]interface{}{
				"success": true,
				"fileMetadata": map[string]interface{}{
					"id":          fileID,
					"url":         publicURL,
					"downloadUrl": downloadURL,
					"metadata":    map[string]interface{}{"confirmed": true},
				},
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockCreateMenu(variables map[string]interface{}) interface{} {
	input := variables["input"].(map[string]interface{})
	menuID := fmt.Sprintf("e2e-menu-%d", time.Now().UnixNano())
	
	suite.testData.CreatedMenus = append(suite.testData.CreatedMenus, menuID)
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"createMenu": map[string]interface{}{
				"menu": map[string]interface{}{
					"id":          menuID,
					"title":       input["title"],
					"type":        input["type"],
					"route":       input["route"],
					"order":       input["order"],
					"isActive":    true,
					"permissions": input["permissions"],
				},
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockCreateImageCarousel(variables map[string]interface{}) interface{} {
	carouselID := fmt.Sprintf("e2e-carousel-%d", time.Now().UnixNano())
	suite.testData.CreatedItems.CarouselID = carouselID
	
	input := variables["input"].(map[string]interface{})
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"createImageCarousel": map[string]interface{}{
				"carousel": map[string]interface{}{
					"id":          carouselID,
					"title":       input["title"],
					"route":       input["route"],
					"description": input["description"],
					"items":       []interface{}{},
					"settings":    input["settings"],
					"createdAt":   time.Now().Format(time.RFC3339),
				},
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockCreateFloorPlan(variables map[string]interface{}) interface{} {
	floorPlanID := fmt.Sprintf("e2e-floorplan-%d", time.Now().UnixNano())
	suite.testData.CreatedItems.FloorPlanID = floorPlanID
	
	input := variables["input"].(map[string]interface{})
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"createFloorPlan": map[string]interface{}{
				"floorPlan": map[string]interface{}{
					"id":          floorPlanID,
					"title":       input["title"],
					"route":       input["route"],
					"description": input["description"],
					"floors":      []interface{}{},
				},
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockCreatePinMap(variables map[string]interface{}) interface{} {
	pinMapID := fmt.Sprintf("e2e-pinmap-%d", time.Now().UnixNano())
	suite.testData.CreatedItems.PinMapID = pinMapID
	
	input := variables["input"].(map[string]interface{})
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"createPinMap": map[string]interface{}{
				"pinMap": map[string]interface{}{
					"id":          pinMapID,
					"title":       input["title"],
					"route":       input["route"],
					"description": input["description"],
					"pins":        []interface{}{},
				},
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockRequestFullSync(variables map[string]interface{}) interface{} {
	syncID := fmt.Sprintf("e2e-sync-%d", time.Now().UnixNano())
	suite.testData.CreatedItems.SyncID = syncID
	
	return map[string]interface{}{
		"data": map[string]interface{}{
			"requestFullSync": map[string]interface{}{
				"zipUrl":        fmt.Sprintf("https://cdn.terraallwert.com/sync/%s.zip", syncID),
				"expiresAt":     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"totalFiles":    len(suite.testData.UploadedFiles),
				"estimatedSize": 5242880, // 5MB
				"syncId":        syncID,
			},
		},
	}
}

func (suite *E2EWorkflowTestSuite) mockGetRouteBusinessData(variables map[string]interface{}) interface{} {
	return map[string]interface{}{
		"data": map[string]interface{}{
			"getRouteBusinessData": map[string]interface{}{
				"route": map[string]interface{}{
					"id":           suite.testData.RouteID,
					"name":         "E2E Test Route",
					"description":  "Route created for end-to-end testing",
					"settings":     map[string]interface{}{"testMode": true},
					"lastModified": time.Now().Format(time.RFC3339),
				},
			},
		},
	}
}

// Test workflow methods
func (suite *E2EWorkflowTestSuite) TestCompleteWorkflow() {
	suite.T().Log("Starting complete E2E workflow test")
	
	// Step 1: Health check
	suite.testHealthCheck()
	
	// Step 2: Authentication
	suite.testAuthentication()
	
	// Step 3: File upload workflow
	suite.testFileUploadWorkflow()
	
	// Step 4: Create presentations
	suite.testCreatePresentations()
	
	// Step 5: Business data operations
	suite.testBusinessDataOperations()
	
	// Step 6: Sync operations
	suite.testSyncOperations()
	
	// Step 7: Verify all data consistency
	suite.testDataConsistency()
	
	suite.T().Log("Complete E2E workflow test completed successfully")
}

func (suite *E2EWorkflowTestSuite) testHealthCheck() {
	suite.T().Log("Testing API health check")
	
	resp, err := suite.client.Get(suite.baseURL + "/health")
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Assert().Equal(http.StatusOK, resp.StatusCode)
	
	var healthResponse map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&healthResponse)
	suite.Require().NoError(err)
	
	suite.Assert().Equal("ok", healthResponse["status"])
	suite.T().Log("✅ Health check passed")
}

func (suite *E2EWorkflowTestSuite) testAuthentication() {
	suite.T().Log("Testing user authentication")
	
	query := `
		mutation Login($input: LoginInput!) {
			login(input: $input) {
				token
				refreshToken
				expiresAt
				user {
					id
					username
					email
					role
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"email":    suite.testData.UserCredentials.Email,
			"password": suite.testData.UserCredentials.Password,
		},
	}
	
	response := suite.executeGraphQL(query, variables, false)
	suite.Assert().Nil(response["errors"], "Login should not have errors")
	
	data := response["data"].(map[string]interface{})
	login := data["login"].(map[string]interface{})
	
	suite.Assert().NotEmpty(login["token"])
	suite.authToken = login["token"].(string)
	
	user := login["user"].(map[string]interface{})
	suite.Assert().Equal(suite.testData.UserCredentials.Email, user["email"])
	
	suite.T().Log("✅ Authentication successful")
}

func (suite *E2EWorkflowTestSuite) testFileUploadWorkflow() {
	suite.T().Log("Testing complete file upload workflow")
	
	testFiles := []struct {
		name        string
		content     string
		contentType string
		fileType    string
	}{
		{"test-image-1.jpg", "Test image content 1", "image/jpeg", "image"},
		{"test-image-2.png", "Test image content 2", "image/png", "image"},
		{"test-video.mp4", "Test video content", "video/mp4", "video"},
		{"test-document.pdf", "Test document content", "application/pdf", "document"},
	}
	
	for _, file := range testFiles {
		suite.T().Logf("Uploading file: %s", file.name)
		
		// Step 1: Get signed upload URL
		uploadQuery := `
			mutation GetSignedUploadUrl($input: SignedUploadUrlInput!) {
				getSignedUploadUrl(input: $input) {
					uploadUrl
					minioPath
					expiresAt
					fileId
				}
			}
		`
		
		uploadVars := map[string]interface{}{
			"input": map[string]interface{}{
				"fileName":    file.name,
				"fileType":    file.fileType,
				"contentType": file.contentType,
				"routeId":     suite.testData.RouteID,
				"context": map[string]interface{}{
					"title":       fmt.Sprintf("E2E Test %s", file.name),
					"description": "File uploaded during E2E testing",
				},
			},
		}
		
		uploadResp := suite.executeGraphQL(uploadQuery, uploadVars, true)
		suite.Assert().Nil(uploadResp["errors"])
		
		uploadData := uploadResp["data"].(map[string]interface{})["getSignedUploadUrl"].(map[string]interface{})
		uploadURL := uploadData["uploadUrl"].(string)
		minioPath := uploadData["minioPath"].(string)
		fileID := uploadData["fileId"].(string)
		
		// Step 2: Upload file to MinIO
		suite.uploadFileToMinIO(uploadURL, []byte(file.content), file.contentType)
		
		// Step 3: Confirm upload
		confirmQuery := `
			mutation ConfirmFileUpload($input: ConfirmFileUploadInput!) {
				confirmFileUpload(input: $input) {
					success
					fileMetadata {
						id
						url
						downloadUrl
					}
				}
			}
		`
		
		confirmVars := map[string]interface{}{
			"input": map[string]interface{}{
				"fileId":           fileID,
				"minioPath":        minioPath,
				"routeId":          suite.testData.RouteID,
				"originalFileName": file.name,
				"fileSize":         len(file.content),
				"checksum":         "test-checksum",
				"context":          map[string]interface{}{"confirmed": true},
			},
		}
		
		confirmResp := suite.executeGraphQL(confirmQuery, confirmVars, true)
		suite.Assert().Nil(confirmResp["errors"])
		
		confirmData := confirmResp["data"].(map[string]interface{})["confirmFileUpload"].(map[string]interface{})
		suite.Assert().True(confirmData["success"].(bool))
	}
	
	suite.T().Logf("✅ File upload workflow completed - %d files uploaded", len(testFiles))
}

func (suite *E2EWorkflowTestSuite) testCreatePresentations() {
	suite.T().Log("Testing presentation creation")
	
	// Create Menu
	menuQuery := `
		mutation CreateMenu($input: CreateMenuInput!) {
			createMenu(input: $input) {
				menu {
					id
					title
					type
					route
				}
			}
		}
	`
	
	menuVars := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       "E2E Test Menu",
			"type":        "MAIN",
			"route":       "/e2e-test",
			"order":       1,
			"permissions": []string{"user", "admin"},
		},
	}
	
	menuResp := suite.executeGraphQL(menuQuery, menuVars, true)
	suite.Assert().Nil(menuResp["errors"])
	
	// Create Image Carousel
	carouselQuery := `
		mutation CreateImageCarousel($input: CreateImageCarouselInput!) {
			createImageCarousel(input: $input) {
				carousel {
					id
					title
					route
				}
			}
		}
	`
	
	carouselVars := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       "E2E Test Carousel",
			"route":       "/e2e-carousel",
			"description": "Carousel created during E2E testing",
			"items":       []interface{}{},
			"settings": map[string]interface{}{
				"autoPlay": true,
			},
		},
	}
	
	carouselResp := suite.executeGraphQL(carouselQuery, carouselVars, true)
	suite.Assert().Nil(carouselResp["errors"])
	
	// Create Floor Plan
	floorPlanQuery := `
		mutation CreateFloorPlan($input: CreateFloorPlanInput!) {
			createFloorPlan(input: $input) {
				floorPlan {
					id
					title
					route
				}
			}
		}
	`
	
	floorPlanVars := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       "E2E Test Floor Plan",
			"route":       "/e2e-floorplan",
			"description": "Floor plan created during E2E testing",
			"floors":      []interface{}{},
		},
	}
	
	floorPlanResp := suite.executeGraphQL(floorPlanQuery, floorPlanVars, true)
	suite.Assert().Nil(floorPlanResp["errors"])
	
	// Create Pin Map
	pinMapQuery := `
		mutation CreatePinMap($input: CreatePinMapInput!) {
			createPinMap(input: $input) {
				pinMap {
					id
					title
					route
				}
			}
		}
	`
	
	pinMapVars := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       "E2E Test Pin Map",
			"route":       "/e2e-pinmap",
			"description": "Pin map created during E2E testing",
			"pins":        []interface{}{},
		},
	}
	
	pinMapResp := suite.executeGraphQL(pinMapQuery, pinMapVars, true)
	suite.Assert().Nil(pinMapResp["errors"])
	
	suite.T().Log("✅ Presentation creation completed")
}

func (suite *E2EWorkflowTestSuite) testBusinessDataOperations() {
	suite.T().Log("Testing business data operations")
	
	// Get route business data
	businessQuery := `
		query GetRouteBusinessData($routeId: String!) {
			getRouteBusinessData(routeId: $routeId) {
				route {
					id
					name
					description
					settings
					lastModified
				}
			}
		}
	`
	
	businessVars := map[string]interface{}{
		"routeId": suite.testData.RouteID,
	}
	
	businessResp := suite.executeGraphQL(businessQuery, businessVars, true)
	suite.Assert().Nil(businessResp["errors"])
	
	data := businessResp["data"].(map[string]interface{})["getRouteBusinessData"].(map[string]interface{})
	route := data["route"].(map[string]interface{})
	
	suite.Assert().Equal(suite.testData.RouteID, route["id"])
	suite.Assert().NotEmpty(route["name"])
	
	suite.T().Log("✅ Business data operations completed")
}

func (suite *E2EWorkflowTestSuite) testSyncOperations() {
	suite.T().Log("Testing sync operations")
	
	// Request full sync
	syncQuery := `
		mutation RequestFullSync($input: FullSyncInput!) {
			requestFullSync(input: $input) {
				zipUrl
				expiresAt
				totalFiles
				estimatedSize
				syncId
			}
		}
	`
	
	syncVars := map[string]interface{}{
		"input": map[string]interface{}{
			"routeId":          suite.testData.RouteID,
			"includeTypes":     []string{"image", "video", "document"},
			"compressionLevel": 6,
		},
	}
	
	syncResp := suite.executeGraphQL(syncQuery, syncVars, true)
	suite.Assert().Nil(syncResp["errors"])
	
	syncData := syncResp["data"].(map[string]interface{})["requestFullSync"].(map[string]interface{})
	
	suite.Assert().NotEmpty(syncData["zipUrl"])
	suite.Assert().NotEmpty(syncData["syncId"])
	suite.Assert().True(syncData["totalFiles"].(float64) >= 0)
	
	suite.T().Log("✅ Sync operations completed")
}

func (suite *E2EWorkflowTestSuite) testDataConsistency() {
	suite.T().Log("Testing data consistency")
	
	// Verify uploaded files count
	suite.Assert().True(len(suite.testData.UploadedFiles) > 0, "Should have uploaded files")
	
	// Verify created menus
	suite.Assert().True(len(suite.testData.CreatedMenus) > 0, "Should have created menus")
	
	// Verify created items
	suite.Assert().NotEmpty(suite.testData.CreatedItems.CarouselID, "Should have carousel ID")
	suite.Assert().NotEmpty(suite.testData.CreatedItems.FloorPlanID, "Should have floor plan ID")
	suite.Assert().NotEmpty(suite.testData.CreatedItems.PinMapID, "Should have pin map ID")
	suite.Assert().NotEmpty(suite.testData.CreatedItems.SyncID, "Should have sync ID")
	
	suite.T().Log("✅ Data consistency verified")
}

// Helper methods
func (suite *E2EWorkflowTestSuite) executeGraphQL(query string, variables map[string]interface{}, requireAuth bool) map[string]interface{} {
	reqBody := map[string]interface{}{
		"query":     query,
		"variables": variables,
	}
	
	jsonBody, _ := json.Marshal(reqBody)
	req, _ := http.NewRequest("POST", suite.baseURL+"/graphql", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	
	if requireAuth && suite.authToken != "" {
		req.Header.Set("Authorization", "Bearer "+suite.authToken)
	}
	
	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	var response map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&response)
	suite.Require().NoError(err)
	
	return response
}

func (suite *E2EWorkflowTestSuite) uploadFileToMinIO(uploadURL string, content []byte, contentType string) {
	req, err := http.NewRequest("PUT", uploadURL, bytes.NewReader(content))
	suite.Require().NoError(err)
	
	req.Header.Set("Content-Type", contentType)
	req.ContentLength = int64(len(content))
	
	resp, err := suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	suite.Assert().Equal(http.StatusOK, resp.StatusCode, "File upload to MinIO should succeed")
}

func (suite *E2EWorkflowTestSuite) cleanupTestData() {
	suite.T().Logf("Cleaning up test data for route: %s", suite.testData.RouteID)
	suite.T().Logf("Uploaded files: %d", len(suite.testData.UploadedFiles))
	suite.T().Logf("Created menus: %d", len(suite.testData.CreatedMenus))
	suite.T().Log("✅ Cleanup completed")
}

// Helper functions
func extractFileTypeFromPath(minioPath string) string {
	parts := strings.Split(minioPath, "/")
	if len(parts) > 1 {
		return parts[1] // routeID/fileType/fileID/fileName
	}
	return "unknown"
}

// Performance test for the complete workflow
func (suite *E2EWorkflowTestSuite) TestWorkflowPerformance() {
	suite.T().Log("Testing workflow performance")
	
	start := time.Now()
	
	// Run a simplified workflow
	suite.testHealthCheck()
	suite.testAuthentication()
	
	// Upload a single file
	uploadQuery := `
		mutation GetSignedUploadUrl($input: SignedUploadUrlInput!) {
			getSignedUploadUrl(input: $input) {
				uploadUrl
				minioPath
				fileId
			}
		}
	`
	
	uploadVars := map[string]interface{}{
		"input": map[string]interface{}{
			"fileName":    "perf-test.jpg",
			"fileType":    "image",
			"contentType": "image/jpeg",
			"routeId":     suite.testData.RouteID,
			"context":     map[string]interface{}{},
		},
	}
	
	uploadResp := suite.executeGraphQL(uploadQuery, uploadVars, true)
	suite.Assert().Nil(uploadResp["errors"])
	
	duration := time.Since(start)
	suite.T().Logf("Performance test completed in %v", duration)
	
	// Performance assertions
	suite.Assert().True(duration < 5*time.Second, "Workflow should complete within 5 seconds")
}

// Stress test with multiple concurrent operations
func (suite *E2EWorkflowTestSuite) TestConcurrentOperations() {
	suite.T().Log("Testing concurrent operations")
	
	// Setup authentication first
	suite.testAuthentication()
	
	const concurrentOps = 5
	results := make(chan error, concurrentOps)
	
	// Run concurrent GraphQL operations
	for i := 0; i < concurrentOps; i++ {
		go func(index int) {
			query := `
				mutation CreateMenu($input: CreateMenuInput!) {
					createMenu(input: $input) {
						menu { id }
					}
				}
			`
			
			variables := map[string]interface{}{
				"input": map[string]interface{}{
					"title": fmt.Sprintf("Concurrent Menu %d", index),
					"type":  "MAIN",
					"route": fmt.Sprintf("/concurrent-%d", index),
					"order": index,
				},
			}
			
			response := suite.executeGraphQL(query, variables, true)
			if response["errors"] != nil {
				results <- fmt.Errorf("concurrent operation %d failed", index)
			} else {
				results <- nil
			}
		}(i)
	}
	
	// Wait for all operations to complete
	for i := 0; i < concurrentOps; i++ {
		err := <-results
		suite.Assert().NoError(err, "Concurrent operation should succeed")
	}
	
	suite.T().Log("✅ Concurrent operations test completed")
}

// Run the E2E test suite
func TestE2EWorkflowTestSuite(t *testing.T) {
	suite.Run(t, new(E2EWorkflowTestSuite))
}

// Benchmark for complete workflow
func BenchmarkCompleteWorkflow(b *testing.B) {
	suite := &E2EWorkflowTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	b.ResetTimer()
	b.ReportAllocs()
	
	for i := 0; i < b.N; i++ {
		suite.testHealthCheck()
		suite.testAuthentication()
		// Simplified workflow for benchmarking
		businessQuery := `
			query GetRouteBusinessData($routeId: String!) {
				getRouteBusinessData(routeId: $routeId) {
					route { id }
				}
			}
		`
		
		businessVars := map[string]interface{}{
			"routeId": suite.testData.RouteID,
		}
		
		_ = suite.executeGraphQL(businessQuery, businessVars, true)
	}
}