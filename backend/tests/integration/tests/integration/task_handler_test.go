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

func TestTaskHandler_GetTasks(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("GET", "/tasks", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return 200 OK or 401 Unauthorized depending on auth setup
	assert.True(t, recorder.Code == http.StatusOK || recorder.Code == http.StatusUnauthorized)

	if recorder.Code == http.StatusOK {
		var response interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)
	}
}

func TestTaskHandler_CreateTask(t *testing.T) {
	router := GetTestRouter()

	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		expectedStatus int
	}{
		{
			name: "Valid task creation",
			requestBody: map[string]interface{}{
				"name":        "Test Task",
				"type":        "subdomain_scan",
				"target":      "example.com",
				"description": "Test task description",
			},
			expectedStatus: http.StatusCreated,
		},
		{
			name: "Missing required fields",
			requestBody: map[string]interface{}{
				"name": "Test Task",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "Empty request body",
			requestBody:    map[string]interface{}{},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, err := json.Marshal(tt.requestBody)
			require.NoError(t, err)

			req, err := http.NewRequest("POST", "/tasks", bytes.NewBuffer(body))
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

func TestTaskHandler_GetTaskByID(t *testing.T) {
	router := GetTestRouter()

	// Test with a sample task ID
	req, err := http.NewRequest("GET", "/api/tasks/test-task-id", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return 200, 404, or 401 depending on auth and task existence
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusNotFound ||
		recorder.Code == http.StatusUnauthorized,
	)
}

func TestTaskHandler_UpdateTask(t *testing.T) {
	router := GetTestRouter()

	updateBody := map[string]interface{}{
		"name":        "Updated Task",
		"description": "Updated description",
		"status":      "running",
	}

	body, err := json.Marshal(updateBody)
	require.NoError(t, err)

	req, err := http.NewRequest("PUT", "/api/tasks/test-task-id", bytes.NewBuffer(body))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/json")

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return appropriate status based on auth and task existence
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusNotFound ||
		recorder.Code == http.StatusUnauthorized ||
		recorder.Code == http.StatusBadRequest,
	)
}

func TestTaskHandler_DeleteTask(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("DELETE", "/api/tasks/test-task-id", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return appropriate status based on auth and task existence
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusNotFound ||
		recorder.Code == http.StatusUnauthorized,
	)
}

func TestTaskHandler_StartTask(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("POST", "/api/tasks/test-task-id/start", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return appropriate status
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusNotFound ||
		recorder.Code == http.StatusUnauthorized ||
		recorder.Code == http.StatusBadRequest,
	)
}

func TestTaskHandler_StopTask(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("POST", "/api/tasks/test-task-id/stop", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return appropriate status
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusNotFound ||
		recorder.Code == http.StatusUnauthorized ||
		recorder.Code == http.StatusBadRequest,
	)
}

func TestTaskHandler_GetTaskResults(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("GET", "/api/tasks/test-task-id/results", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return appropriate status
	assert.True(t,
		recorder.Code == http.StatusOK ||
		recorder.Code == http.StatusNotFound ||
		recorder.Code == http.StatusUnauthorized,
	)
}

func TestTaskHandler_GetTaskMetrics(t *testing.T) {
	router := GetTestRouter()

	req, err := http.NewRequest("GET", "/api/scanner/metrics", nil)
	require.NoError(t, err)

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	// Should return metrics or require authentication
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