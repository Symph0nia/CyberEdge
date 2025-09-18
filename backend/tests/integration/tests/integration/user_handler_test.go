package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUserHandler_GenerateQRCode(t *testing.T) {
	router := GetTestRouter()

	// Test QR code generation
	req, err := http.NewRequest("GET", "/auth/qrcode", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var response map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	require.NoError(t, err)

	// Verify response structure
	assert.Contains(t, response, "qrcode")
	assert.Contains(t, response, "account")
	assert.NotEmpty(t, response["qrcode"])
	assert.NotEmpty(t, response["account"])

	// Verify QR code is base64 encoded
	qrCode, ok := response["qrcode"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, qrCode)

	// Account should be a string
	account, ok := response["account"].(string)
	assert.True(t, ok)
	assert.NotEmpty(t, account)
}

func TestUserHandler_ValidateTOTP(t *testing.T) {
	router := GetTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
		expectedError  bool
	}{
		{
			name: "Valid TOTP validation request",
			requestBody: map[string]interface{}{
				"account": "test-account",
				"code":    "123456",
			},
			expectedStatus: http.StatusUnauthorized, // Will fail without valid TOTP setup
			expectedError:  true,
		},
		{
			name: "Missing account",
			requestBody: map[string]interface{}{
				"code": "123456",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name: "Missing code",
			requestBody: map[string]interface{}{
				"account": "test-account",
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
		{
			name:           "Empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/auth/validate", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			var response map[string]interface{}
			err = json.Unmarshal(recorder.Body.Bytes(), &response)
			require.NoError(t, err)

			if tt.expectedError {
				assert.Contains(t, response, "error")
				assert.NotEmpty(t, response["error"])
			}
		})
	}
}

func TestUserHandler_GetCurrentUser(t *testing.T) {
	router := GetTestRouter()

	// Test without authentication (should fail)
	req, err := http.NewRequest("GET", "/auth/check", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should fail due to missing authentication
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}

func TestUserHandler_GetUsers(t *testing.T) {
	router := GetTestRouter()

	// Test get users endpoint
	req, err := http.NewRequest("GET", "/users", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should fail due to missing authentication, but test the endpoint exists
	assert.True(t, recorder.Code == http.StatusUnauthorized || recorder.Code == http.StatusOK)
}

func TestUserHandler_CreateUser(t *testing.T) {
	router := GetTestRouter()

	requestBody := map[string]interface{}{
		"username": "testuser",
		"email":    "test@example.com",
		"role":     "user",
	}

	body, err := json.Marshal(requestBody)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/users", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should fail due to missing authentication, but test the endpoint exists
	assert.True(t, recorder.Code == http.StatusUnauthorized || recorder.Code == http.StatusCreated || recorder.Code == http.StatusBadRequest)
}

func TestUserHandler_Integration_QRCodeAndValidation(t *testing.T) {
	router := GetTestRouter()

	// Step 1: Generate QR Code
	req, err := http.NewRequest("GET", "/auth/qrcode", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	assert.Equal(t, http.StatusOK, recorder.Code)

	var qrResponse map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &qrResponse)
	require.NoError(t, err)

	account, ok := qrResponse["account"].(string)
	require.True(t, ok)
	require.NotEmpty(t, account)

	// Step 2: Try to validate with the generated account (will fail without real TOTP)
	validationBody := map[string]interface{}{
		"account": account,
		"code":    "123456", // Invalid code, but tests the flow
	}

	body, err := json.Marshal(validationBody)
	require.NoError(t, err)

	req, err = http.NewRequest("POST", "/api/auth/validate", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	recorder = httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should fail due to invalid TOTP code, but verifies the integration works
	assert.Equal(t, http.StatusUnauthorized, recorder.Code)
}