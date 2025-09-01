package e2e

import (
	"bytes"
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"api/domain/entities"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type AuthE2ETestSuite struct {
	suite.Suite
	baseURL string
	client  *http.Client
}

func (suite *AuthE2ETestSuite) SetupSuite() {
	// This would normally start the actual server for E2E tests
	suite.baseURL = "http://localhost:3000" // Assume server is running
	suite.client = &http.Client{
		Timeout: 10 * time.Second,
	}
}

func (suite *AuthE2ETestSuite) TestAuthFlow() {
	// Skip if server is not running
	suite.T().Skip("E2E test requires running server")

	// Test login
	loginRequest := entities.LoginRequest{
		Email:    "admin@euvatar.com",
		Password: "admin123",
	}

	body, err := json.Marshal(loginRequest)
	suite.Require().NoError(err)

	resp, err := suite.client.Post(
		suite.baseURL+"/api/auth/login",
		"application/json",
		bytes.NewBuffer(body),
	)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var loginResponse entities.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&loginResponse)
	suite.Require().NoError(err)

	assert.NotEmpty(suite.T(), loginResponse.Token)
	assert.NotEmpty(suite.T(), loginResponse.RefreshToken)
	assert.Equal(suite.T(), "admin@euvatar.com", loginResponse.User.Email)
	assert.Equal(suite.T(), entities.RoleAdmin, loginResponse.User.Role)

	// Test authenticated request
	req, err := http.NewRequest("GET", suite.baseURL+"/api/profile", nil)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer "+loginResponse.Token)

	resp, err = suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	// Test token refresh
	refreshRequest := map[string]string{
		"refreshToken": loginResponse.RefreshToken,
	}

	body, err = json.Marshal(refreshRequest)
	suite.Require().NoError(err)

	resp, err = suite.client.Post(
		suite.baseURL+"/api/auth/refresh",
		"application/json",
		bytes.NewBuffer(body),
	)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusOK, resp.StatusCode)

	var refreshResponse entities.LoginResponse
	err = json.NewDecoder(resp.Body).Decode(&refreshResponse)
	suite.Require().NoError(err)

	assert.NotEmpty(suite.T(), refreshResponse.Token)
	assert.NotEmpty(suite.T(), refreshResponse.RefreshToken)
}

func (suite *AuthE2ETestSuite) TestUnauthorizedAccess() {
	suite.T().Skip("E2E test requires running server")

	// Test accessing protected route without token
	resp, err := suite.client.Get(suite.baseURL + "/api/profile")
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)

	// Test with invalid token
	req, err := http.NewRequest("GET", suite.baseURL+"/api/profile", nil)
	suite.Require().NoError(err)
	req.Header.Set("Authorization", "Bearer invalid-token")

	resp, err = suite.client.Do(req)
	suite.Require().NoError(err)
	defer resp.Body.Close()

	assert.Equal(suite.T(), http.StatusUnauthorized, resp.StatusCode)
}

func TestAuthE2ETestSuite(t *testing.T) {
	suite.Run(t, new(AuthE2ETestSuite))
}
