package main

import (
	"testing"
	"net/http"
	"net/http/httptest"
	"github.com/gin-gonic/gin"
	"encoding/json"
)

// TestBasicAPIRouting tests basic API routing without external dependencies
func TestBasicAPIRouting(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	// Create a basic router
	router := gin.New()

	// Add basic routes
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	router.GET("/api/version", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"version": "1.0.0"})
	})

	// Test health endpoint
	req, err := http.NewRequest("GET", "/health", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", recorder.Code)
	}

	// Test JSON response structure
	var response map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}
}

// TestAPIErrorHandling tests error response handling
func TestAPIErrorHandling(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	router.GET("/error", func(c *gin.Context) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "bad request"})
	})

	req, err := http.NewRequest("GET", "/error", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	recorder := httptest.NewRecorder()
	router.ServeHTTP(recorder, req)

	if recorder.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", recorder.Code)
	}

	var response map[string]interface{}
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	if err != nil {
		t.Errorf("Failed to parse JSON response: %v", err)
	}

	if response["error"] != "bad request" {
		t.Errorf("Expected error 'bad request', got %v", response["error"])
	}
}

// TestAPIMethods tests different HTTP methods
func TestAPIMethods(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add routes for different methods
	router.GET("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "GET"})
	})

	router.POST("/test", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"method": "POST"})
	})

	methods := []string{"GET", "POST"}

	for _, method := range methods {
		req, err := http.NewRequest(method, "/test", nil)
		if err != nil {
			t.Fatalf("Failed to create %s request: %v", method, err)
		}

		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, req)

		if recorder.Code != http.StatusOK {
			t.Errorf("Method %s: expected status 200, got %d", method, recorder.Code)
		}

		var response map[string]interface{}
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		if err != nil {
			t.Errorf("Failed to parse JSON response for %s: %v", method, err)
		}

		if response["method"] != method {
			t.Errorf("Method %s: expected method '%s', got %v", method, method, response["method"])
		}
	}
}