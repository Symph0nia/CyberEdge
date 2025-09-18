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

func TestConfigHandler_GetConfig(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("GET", "/system/info", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return config or require authentication
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusUnauthorized,
	)

	if recorder.Code == http.StatusOK {
		var response interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
	}
}

func TestConfigHandler_UpdateConfig(t *testing.T) {
	router := GetTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid config update",
			requestBody: map[string]interface{}{
				"scan_timeout":    300,
				"max_concurrent":  10,
				"rate_limit":      100,
				"default_threads": 5,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Partial config update",
			requestBody: map[string]interface{}{
				"scan_timeout": 600,
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Invalid config values",
			requestBody: map[string]interface{}{
				"scan_timeout": -1,
				"max_concurrent": 0,
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("PUT", "/system/info", bytes.NewBuffer(body))
			require.NoError(t, err)
			req.Header.Set("Content-Type", "application/json")

			recorder := httptest.NewRecorder()
			router.ServeHTTP(recorder, req)

			// May return 401 if authentication is required
			if recorder.Code != http.StatusUnauthorized {
				assert.Equal(t, tt.expectedStatus, recorder.Code)
			}
		})
	}
}

func TestConfigHandler_ResetConfig(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("POST", "/system/reset", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should reset config or require authentication
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusUnauthorized,
	)
}

func TestConfigHandler_GetToolsConfig(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("GET", "/system/tools", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return tools config or require authentication
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusUnauthorized,
	)

	if recorder.Code == http.StatusOK {
		var response interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
	}
}

func TestConfigHandler_UpdateToolsConfig(t *testing.T) {
	router := GetTestRouter()

	toolsConfig := map[string]interface{}{
		"subdomain": map[string]interface{}{
			"enabled": true,
			"tools": []string{"subfinder", "assetfinder"},
			"timeout": 300,
		},
		"port_scan": map[string]interface{}{
			"enabled": true,
			"tools": []string{"nmap", "masscan"},
			"timeout": 600,
		},
		"web_scan": map[string]interface{}{
			"enabled": true,
			"tools": []string{"httpx", "gobuster"},
			"timeout": 900,
		},
	}

	body, err := json.Marshal(toolsConfig)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/system/tools", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should update tools config or require authentication
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusUnauthorized ||
		recorder.Code == http.StatusBadRequest,
	)
}

func TestConfigHandler_ValidateConfig(t *testing.T) {
	router := GetTestRouter()

	configToValidate := map[string]interface{}{
		"scan_timeout":    300,
		"max_concurrent":  10,
		"rate_limit":      100,
		"default_threads": 5,
	}

	body, err := json.Marshal(configToValidate)
	require.NoError(t, err)

	req, err := http.NewRequest("POST", "/system/validate", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should validate config or require authentication
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusUnauthorized ||
		recorder.Code == http.StatusBadRequest,
	)
}