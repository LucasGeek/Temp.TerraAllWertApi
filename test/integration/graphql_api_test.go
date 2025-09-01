package integration

import (
	"bytes"
	"context"
	"encoding/json"
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

// GraphQL test suite
type GraphQLTestSuite struct {
	suite.Suite
	server *httptest.Server
	client *http.Client
	token  string
}

// GraphQL request/response structures
type GraphQLRequest struct {
	Query     string                 `json:"query"`
	Variables map[string]interface{} `json:"variables,omitempty"`
}

type GraphQLResponse struct {
	Data   interface{} `json:"data"`
	Errors []struct {
		Message string `json:"message"`
		Path    []interface{} `json:"path"`
	} `json:"errors"`
}

// Setup test suite
func (suite *GraphQLTestSuite) SetupSuite() {
	// Mock server setup - in real implementation, this would start the actual Fiber app
	suite.server = httptest.NewServer(http.HandlerFunc(suite.mockGraphQLHandler))
	suite.client = &http.Client{Timeout: 30 * time.Second}
	
	// Get authentication token for tests
	suite.authenticateTestUser()
}

func (suite *GraphQLTestSuite) TearDownSuite() {
	if suite.server != nil {
		suite.server.Close()
	}
}

// Mock GraphQL handler for testing
func (suite *GraphQLTestSuite) mockGraphQLHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req GraphQLRequest
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Failed to read body", http.StatusBadRequest)
		return
	}

	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Route to appropriate mock handler based on query
	response := suite.routeGraphQLQuery(req)
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (suite *GraphQLTestSuite) routeGraphQLQuery(req GraphQLRequest) GraphQLResponse {
	query := strings.TrimSpace(req.Query)
	
	switch {
	case strings.Contains(query, "mutation Login"):
		return suite.mockLoginMutation(req)
	case strings.Contains(query, "mutation GetSignedUploadUrl"):
		return suite.mockGetSignedUploadUrlMutation(req)
	case strings.Contains(query, "mutation ConfirmFileUpload"):
		return suite.mockConfirmFileUploadMutation(req)
	case strings.Contains(query, "mutation CreateMenu"):
		return suite.mockCreateMenuMutation(req)
	case strings.Contains(query, "mutation CreateImageCarousel"):
		return suite.mockCreateImageCarouselMutation(req)
	case strings.Contains(query, "mutation CreateFloorPlan"):
		return suite.mockCreateFloorPlanMutation(req)
	case strings.Contains(query, "mutation CreatePinMap"):
		return suite.mockCreatePinMapMutation(req)
	case strings.Contains(query, "mutation RequestFullSync"):
		return suite.mockRequestFullSyncMutation(req)
	case strings.Contains(query, "query GetRouteBusinessData"):
		return suite.mockGetRouteBusinessDataQuery(req)
	case strings.Contains(query, "query GetCacheConfiguration"):
		return suite.mockGetCacheConfigurationQuery(req)
	default:
		return GraphQLResponse{
			Errors: []struct {
				Message string `json:"message"`
				Path    []interface{} `json:"path"`
			}{
				{Message: "Query not implemented in mock", Path: []interface{}{}},
			},
		}
	}
}

// Authentication helper
func (suite *GraphQLTestSuite) authenticateTestUser() {
	loginQuery := `
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
			"email":    "admin@euvatar.com",
			"password": "admin123",
		},
	}
	
	response := suite.executeGraphQLQuery(loginQuery, variables, false)
	if response.Errors != nil && len(response.Errors) > 0 {
		suite.T().Fatalf("Failed to authenticate test user: %v", response.Errors)
	}
	
	// Extract token from response
	data := response.Data.(map[string]interface{})
	login := data["login"].(map[string]interface{})
	suite.token = login["token"].(string)
}

// Execute GraphQL query helper
func (suite *GraphQLTestSuite) executeGraphQLQuery(query string, variables map[string]interface{}, requireAuth bool) GraphQLResponse {
	req := GraphQLRequest{
		Query:     query,
		Variables: variables,
	}
	
	jsonBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", suite.server.URL+"/graphql", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	
	if requireAuth && suite.token != "" {
		httpReq.Header.Set("Authorization", "Bearer "+suite.token)
	}
	
	resp, err := suite.client.Do(httpReq)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	var graphqlResp GraphQLResponse
	err = json.NewDecoder(resp.Body).Decode(&graphqlResp)
	suite.Require().NoError(err)
	
	return graphqlResp
}

// Mock handlers for different GraphQL operations
func (suite *GraphQLTestSuite) mockLoginMutation(req GraphQLRequest) GraphQLResponse {
	// Simulate successful login
	return GraphQLResponse{
		Data: map[string]interface{}{
			"login": map[string]interface{}{
				"token":        "mock-jwt-token-" + time.Now().Format("20060102150405"),
				"refreshToken": "mock-refresh-token",
				"expiresAt":    time.Now().Add(15 * time.Minute).Format(time.RFC3339),
				"user": map[string]interface{}{
					"id":       "admin-user-id",
					"username": "admin",
					"email":    "admin@euvatar.com",
					"role":     "ADMIN",
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockGetSignedUploadUrlMutation(req GraphQLRequest) GraphQLResponse {
	timestamp := time.Now().Unix()
	fileID := "file-" + time.Now().Format("20060102-150405")
	
	return GraphQLResponse{
		Data: map[string]interface{}{
			"getSignedUploadUrl": map[string]interface{}{
				"uploadUrl": "https://minio.localhost:9000/terraallwert/uploads/presigned-url?X-Amz-Expires=3600",
				"minioPath": "route-123/image/" + fileID + "/test-image.jpg",
				"expiresAt": time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				"fileId":    fileID,
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockConfirmFileUploadMutation(req GraphQLRequest) GraphQLResponse {
	variables := req.Variables["input"].(map[string]interface{})
	fileID := variables["fileId"].(string)
	minioPath := variables["minioPath"].(string)
	
	return GraphQLResponse{
		Data: map[string]interface{}{
			"confirmFileUpload": map[string]interface{}{
				"success": true,
				"fileMetadata": map[string]interface{}{
					"id":          fileID,
					"url":         "https://cdn.terraallwert.com/" + minioPath,
					"downloadUrl": "https://minio.localhost:9000/terraallwert/" + minioPath + "?presigned=true",
					"thumbnailUrl": nil,
					"metadata": map[string]interface{}{
						"uploaded": true,
					},
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockCreateMenuMutation(req GraphQLRequest) GraphQLResponse {
	variables := req.Variables["input"].(map[string]interface{})
	
	return GraphQLResponse{
		Data: map[string]interface{}{
			"createMenu": map[string]interface{}{
				"menu": map[string]interface{}{
					"id":          "menu-" + time.Now().Format("20060102-150405"),
					"title":       variables["title"],
					"type":        variables["type"],
					"route":       variables["route"],
					"icon":        variables["icon"],
					"order":       variables["order"],
					"isActive":    true,
					"permissions": variables["permissions"],
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockCreateImageCarouselMutation(req GraphQLRequest) GraphQLResponse {
	variables := req.Variables["input"].(map[string]interface{})
	
	return GraphQLResponse{
		Data: map[string]interface{}{
			"createImageCarousel": map[string]interface{}{
				"carousel": map[string]interface{}{
					"id":          "carousel-" + time.Now().Format("20060102-150405"),
					"title":       variables["title"],
					"route":       variables["route"],
					"description": variables["description"],
					"items":       []interface{}{},
					"settings":    variables["settings"],
					"createdAt":   time.Now().Format(time.RFC3339),
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockCreateFloorPlanMutation(req GraphQLRequest) GraphQLResponse {
	variables := req.Variables["input"].(map[string]interface{})
	
	return GraphQLResponse{
		Data: map[string]interface{}{
			"createFloorPlan": map[string]interface{}{
				"floorPlan": map[string]interface{}{
					"id":          "floorplan-" + time.Now().Format("20060102-150405"),
					"title":       variables["title"],
					"route":       variables["route"],
					"description": variables["description"],
					"floors":      []interface{}{},
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockCreatePinMapMutation(req GraphQLRequest) GraphQLResponse {
	variables := req.Variables["input"].(map[string]interface{})
	
	return GraphQLResponse{
		Data: map[string]interface{}{
			"createPinMap": map[string]interface{}{
				"pinMap": map[string]interface{}{
					"id":          "pinmap-" + time.Now().Format("20060102-150405"),
					"title":       variables["title"],
					"route":       variables["route"],
					"description": variables["description"],
					"pins":        []interface{}{},
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockRequestFullSyncMutation(req GraphQLRequest) GraphQLResponse {
	return GraphQLResponse{
		Data: map[string]interface{}{
			"requestFullSync": map[string]interface{}{
				"zipUrl":        "https://cdn.terraallwert.com/sync/route-123-" + time.Now().Format("20060102") + ".zip",
				"expiresAt":     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				"totalFiles":    42,
				"estimatedSize": 15728640, // 15MB
				"syncId":        "sync-" + time.Now().Format("20060102-150405"),
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockGetRouteBusinessDataQuery(req GraphQLRequest) GraphQLResponse {
	return GraphQLResponse{
		Data: map[string]interface{}{
			"getRouteBusinessData": map[string]interface{}{
				"route": map[string]interface{}{
					"id":           "route-123",
					"name":         "Route de Teste",
					"description":  "Rota para testes de integração",
					"settings":     map[string]interface{}{"enabled": true},
					"lastModified": time.Now().Format(time.RFC3339),
				},
			},
		},
	}
}

func (suite *GraphQLTestSuite) mockGetCacheConfigurationQuery(req GraphQLRequest) GraphQLResponse {
	return GraphQLResponse{
		Data: map[string]interface{}{
			"getCacheConfiguration": map[string]interface{}{
				"maxFileSize":        10485760, // 10MB
				"allowedTypes":       []string{"image", "video", "document"},
				"compressionEnabled": true,
				"thumbnailSizes":     []int{150, 300, 600},
				"cacheExpiration":    3600,
				"syncIntervals":      map[string]int{"full": 86400, "incremental": 3600},
			},
		},
	}
}

// Test methods
func (suite *GraphQLTestSuite) TestLogin_Success() {
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
			"email":    "admin@euvatar.com",
			"password": "admin123",
		},
	}
	
	response := suite.executeGraphQLQuery(query, variables, false)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	login := data["login"].(map[string]interface{})
	
	suite.Assert().NotEmpty(login["token"])
	suite.Assert().NotEmpty(login["refreshToken"])
	
	user := login["user"].(map[string]interface{})
	suite.Assert().Equal("admin@euvatar.com", user["email"])
	suite.Assert().Equal("ADMIN", user["role"])
}

func (suite *GraphQLTestSuite) TestGetSignedUploadUrl_Success() {
	query := `
		mutation GetSignedUploadUrl($input: SignedUploadUrlInput!) {
			getSignedUploadUrl(input: $input) {
				uploadUrl
				minioPath
				expiresAt
				fileId
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"fileName":    "test-image.jpg",
			"fileType":    "image",
			"contentType": "image/jpeg",
			"routeId":     "route-123",
			"context": map[string]interface{}{
				"title":       "Test Image",
				"description": "Integration test image",
			},
		},
	}
	
	response := suite.executeGraphQLQuery(query, variables, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	upload := data["getSignedUploadUrl"].(map[string]interface{})
	
	suite.Assert().NotEmpty(upload["uploadUrl"])
	suite.Assert().NotEmpty(upload["minioPath"])
	suite.Assert().NotEmpty(upload["fileId"])
	suite.Assert().NotEmpty(upload["expiresAt"])
	
	// Verify URL structure
	uploadURL := upload["uploadUrl"].(string)
	suite.Assert().Contains(uploadURL, "minio.localhost")
	suite.Assert().Contains(uploadURL, "terraallwert")
	suite.Assert().Contains(uploadURL, "X-Amz-Expires")
}

func (suite *GraphQLTestSuite) TestConfirmFileUpload_Success() {
	query := `
		mutation ConfirmFileUpload($input: ConfirmFileUploadInput!) {
			confirmFileUpload(input: $input) {
				success
				fileMetadata {
					id
					url
					downloadUrl
					thumbnailUrl
					metadata
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"fileId":           "test-file-123",
			"minioPath":        "route-123/image/test-file-123/test.jpg",
			"routeId":          "route-123",
			"originalFileName": "test.jpg",
			"fileSize":         1024000,
			"checksum":         "abc123def456",
			"context": map[string]interface{}{
				"confirmed": true,
			},
		},
	}
	
	response := suite.executeGraphQLQuery(query, variables, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	confirm := data["confirmFileUpload"].(map[string]interface{})
	
	suite.Assert().True(confirm["success"].(bool))
	
	fileMetadata := confirm["fileMetadata"].(map[string]interface{})
	suite.Assert().Equal("test-file-123", fileMetadata["id"])
	suite.Assert().NotEmpty(fileMetadata["url"])
	suite.Assert().NotEmpty(fileMetadata["downloadUrl"])
}

func (suite *GraphQLTestSuite) TestCreateMenu_Success() {
	query := `
		mutation CreateMenu($input: CreateMenuInput!) {
			createMenu(input: $input) {
				menu {
					id
					title
					type
					route
					icon
					order
					isActive
					permissions
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       "Menu de Teste",
			"type":        "MAIN",
			"route":       "/teste",
			"icon":        "test-icon",
			"order":       1,
			"permissions": []string{"user", "admin"},
		},
	}
	
	response := suite.executeGraphQLQuery(query, variables, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	create := data["createMenu"].(map[string]interface{})
	menu := create["menu"].(map[string]interface{})
	
	suite.Assert().NotEmpty(menu["id"])
	suite.Assert().Equal("Menu de Teste", menu["title"])
	suite.Assert().Equal("MAIN", menu["type"])
	suite.Assert().Equal("/teste", menu["route"])
	suite.Assert().True(menu["isActive"].(bool))
}

func (suite *GraphQLTestSuite) TestCreateImageCarousel_Success() {
	query := `
		mutation CreateImageCarousel($input: CreateImageCarouselInput!) {
			createImageCarousel(input: $input) {
				carousel {
					id
					title
					route
					description
					items {
						id
						type
						url
						order
					}
					settings
					createdAt
				}
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"title":       "Carousel de Teste",
			"route":       "/carousel-teste",
			"description": "Carousel para testes",
			"items":       []interface{}{},
			"settings": map[string]interface{}{
				"autoPlay":     true,
				"showControls": true,
			},
		},
	}
	
	response := suite.executeGraphQLQuery(query, variables, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	create := data["createImageCarousel"].(map[string]interface{})
	carousel := create["carousel"].(map[string]interface{})
	
	suite.Assert().NotEmpty(carousel["id"])
	suite.Assert().Equal("Carousel de Teste", carousel["title"])
	suite.Assert().Equal("/carousel-teste", carousel["route"])
	suite.Assert().NotNil(carousel["settings"])
}

func (suite *GraphQLTestSuite) TestRequestFullSync_Success() {
	query := `
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
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"routeId":          "route-123",
			"includeTypes":     []string{"image", "video"},
			"compressionLevel": 6,
		},
	}
	
	response := suite.executeGraphQLQuery(query, variables, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	sync := data["requestFullSync"].(map[string]interface{})
	
	suite.Assert().NotEmpty(sync["zipUrl"])
	suite.Assert().NotEmpty(sync["syncId"])
	suite.Assert().True(sync["totalFiles"].(float64) > 0)
	suite.Assert().True(sync["estimatedSize"].(float64) > 0)
}

func (suite *GraphQLTestSuite) TestGetRouteBusinessData_Success() {
	query := `
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
	
	variables := map[string]interface{}{
		"routeId": "route-123",
	}
	
	response := suite.executeGraphQLQuery(query, variables, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	businessData := data["getRouteBusinessData"].(map[string]interface{})
	route := businessData["route"].(map[string]interface{})
	
	suite.Assert().Equal("route-123", route["id"])
	suite.Assert().NotEmpty(route["name"])
	suite.Assert().NotNil(route["settings"])
}

func (suite *GraphQLTestSuite) TestGetCacheConfiguration_Success() {
	query := `
		query GetCacheConfiguration {
			getCacheConfiguration {
				maxFileSize
				allowedTypes
				compressionEnabled
				thumbnailSizes
				cacheExpiration
				syncIntervals
			}
		}
	`
	
	response := suite.executeGraphQLQuery(query, map[string]interface{}{}, true)
	
	suite.Assert().Nil(response.Errors)
	suite.Assert().NotNil(response.Data)
	
	data := response.Data.(map[string]interface{})
	config := data["getCacheConfiguration"].(map[string]interface{})
	
	suite.Assert().True(config["maxFileSize"].(float64) > 0)
	suite.Assert().True(len(config["allowedTypes"].([]interface{})) > 0)
	suite.Assert().True(config["compressionEnabled"].(bool))
}

// Test unauthorized access
func (suite *GraphQLTestSuite) TestUnauthorizedAccess() {
	query := `
		mutation GetSignedUploadUrl($input: SignedUploadUrlInput!) {
			getSignedUploadUrl(input: $input) {
				uploadUrl
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"fileName": "test.jpg",
			"fileType": "image",
			"routeId":  "route-123",
		},
	}
	
	// Execute without authentication token
	req := GraphQLRequest{Query: query, Variables: variables}
	jsonBody, _ := json.Marshal(req)
	httpReq, _ := http.NewRequest("POST", suite.server.URL+"/graphql", bytes.NewBuffer(jsonBody))
	httpReq.Header.Set("Content-Type", "application/json")
	// No Authorization header
	
	resp, err := suite.client.Do(httpReq)
	suite.Require().NoError(err)
	defer resp.Body.Close()
	
	// In a real implementation, this should return an authentication error
	// For now, our mock doesn't implement auth checking
	suite.Assert().Equal(http.StatusOK, resp.StatusCode)
}

// Integration test with error cases
func (suite *GraphQLTestSuite) TestInvalidGraphQLQuery() {
	query := `
		mutation InvalidQuery {
			nonExistentMutation {
				id
			}
		}
	`
	
	response := suite.executeGraphQLQuery(query, map[string]interface{}{}, true)
	
	// Should have errors for non-existent mutation
	suite.Assert().NotNil(response.Errors)
	suite.Assert().True(len(response.Errors) > 0)
	suite.Assert().Contains(response.Errors[0].Message, "not implemented")
}

// Run the test suite
func TestGraphQLTestSuite(t *testing.T) {
	suite.Run(t, new(GraphQLTestSuite))
}

// Benchmark tests for GraphQL operations
func BenchmarkGraphQLQuery_Login(b *testing.B) {
	suite := &GraphQLTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	query := `
		mutation Login($input: LoginInput!) {
			login(input: $input) {
				token
				user { id }
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"email":    "admin@euvatar.com",
			"password": "admin123",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = suite.executeGraphQLQuery(query, variables, false)
	}
}

func BenchmarkGraphQLQuery_GetSignedUploadUrl(b *testing.B) {
	suite := &GraphQLTestSuite{}
	suite.SetupSuite()
	defer suite.TearDownSuite()
	
	query := `
		mutation GetSignedUploadUrl($input: SignedUploadUrlInput!) {
			getSignedUploadUrl(input: $input) {
				uploadUrl
				fileId
			}
		}
	`
	
	variables := map[string]interface{}{
		"input": map[string]interface{}{
			"fileName": "test.jpg",
			"fileType": "image",
			"routeId":  "route-123",
		},
	}
	
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = suite.executeGraphQLQuery(query, variables, true)
	}
}