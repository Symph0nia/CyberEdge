package handlers

import (
	"bytes"
	"cyberedge/pkg/models"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockUserService 模拟UserService
type MockUserService struct {
	mock.Mock
}

func (m *MockUserService) CreateUser(username, email, password string) error {
	args := m.Called(username, email, password)
	return args.Error(0)
}

func (m *MockUserService) Login(username, password string) (string, error) {
	args := m.Called(username, password)
	return args.String(0), args.Error(1)
}

func (m *MockUserService) ValidateToken(token string) (*models.User, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) GetAllUsers() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func (m *MockUserService) GetUserByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserService) DeleteUser(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserService) ChangePassword(username, currentPassword, newPassword string) error {
	args := m.Called(username, currentPassword, newPassword)
	return args.Error(0)
}

func (m *MockUserService) Setup2FA(username string) (string, []byte, error) {
	args := m.Called(username)
	return args.String(0), args.Get(1).([]byte), args.Error(2)
}

func (m *MockUserService) Verify2FA(username, code string) error {
	args := m.Called(username, code)
	return args.Error(0)
}

func (m *MockUserService) Disable2FA(username string) error {
	args := m.Called(username)
	return args.Error(0)
}

func (m *MockUserService) ValidatePassword(password string) error {
	args := m.Called(password)
	return args.Error(0)
}

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestUserHandler_Login(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setup          func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid login",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "Password123",
			},
			setup: func(mockService *MockUserService) {
				mockService.On("Login", "testuser", "Password123").Return("fake-jwt-token", nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"token":   "fake-jwt-token",
				"message": "登录成功",
			},
		},
		{
			name: "Invalid credentials",
			requestBody: map[string]interface{}{
				"username": "testuser",
				"password": "wrongpassword",
			},
			setup: func(mockService *MockUserService) {
				mockService.On("Login", "testuser", "wrongpassword").Return("", errors.New("INVALID_CREDENTIALS"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "用户名或密码错误",
			},
		},
		{
			name: "Missing username",
			requestBody: map[string]interface{}{
				"password": "Password123",
			},
			setup:          func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "请求参数无效",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setup(mockService)

			handler := NewUserHandler(mockService)
			router := setupRouter()
			router.POST("/login", handler.Login)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/login", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_Register(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    map[string]interface{}
		setup          func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name: "Valid registration",
			requestBody: map[string]interface{}{
				"username": "newuser",
				"email":    "new@example.com",
				"password": "Password123",
			},
			setup: func(mockService *MockUserService) {
				mockService.On("CreateUser", "newuser", "new@example.com", "Password123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "用户注册成功",
			},
		},
		{
			name: "Username exists",
			requestBody: map[string]interface{}{
				"username": "existinguser",
				"email":    "new@example.com",
				"password": "Password123",
			},
			setup: func(mockService *MockUserService) {
				mockService.On("CreateUser", "existinguser", "new@example.com", "Password123").Return(errors.New("用户名已存在"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "用户名已存在",
			},
		},
		{
			name: "Invalid email",
			requestBody: map[string]interface{}{
				"username": "newuser",
				"email":    "invalid-email",
				"password": "Password123",
			},
			setup:          func(mockService *MockUserService) {},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "请求参数无效",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setup(mockService)

			handler := NewUserHandler(mockService)
			router := setupRouter()
			router.POST("/register", handler.Register)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/register", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_CheckAuth(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		setup          func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Valid token",
			authHeader: "Bearer valid-token",
			setup: func(mockService *MockUserService) {
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
					Role:     "user",
				}
				mockService.On("ValidateToken", "valid-token").Return(user, nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"authenticated": true,
				"user": map[string]interface{}{
					"id":       float64(1), // JSON numbers are float64
					"username": "testuser",
					"email":    "test@example.com",
					"role":     "user",
				},
			},
		},
		{
			name:       "Invalid token",
			authHeader: "Bearer invalid-token",
			setup: func(mockService *MockUserService) {
				mockService.On("ValidateToken", "invalid-token").Return(nil, errors.New("invalid token"))
			},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"authenticated": false,
			},
		},
		{
			name:           "No auth header",
			authHeader:     "",
			setup:          func(mockService *MockUserService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"authenticated": false,
			},
		},
		{
			name:           "Invalid auth format",
			authHeader:     "InvalidFormat token",
			setup:          func(mockService *MockUserService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"authenticated": false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setup(mockService)

			handler := NewUserHandler(mockService)
			router := setupRouter()
			router.GET("/auth/check", handler.CheckAuth)

			req, _ := http.NewRequest("GET", "/auth/check", nil)
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			for key, expectedValue := range tt.expectedBody {
				if key == "user" && response[key] != nil {
					// Deep compare user object
					userMap := response[key].(map[string]interface{})
					expectedUserMap := expectedValue.(map[string]interface{})
					for userKey, userValue := range expectedUserMap {
						assert.Equal(t, userValue, userMap[userKey])
					}
				} else {
					assert.Equal(t, expectedValue, response[key])
				}
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_ChangePassword(t *testing.T) {
	tests := []struct {
		name           string
		authHeader     string
		requestBody    map[string]interface{}
		setup          func(*MockUserService)
		expectedStatus int
		expectedBody   map[string]interface{}
	}{
		{
			name:       "Valid password change",
			authHeader: "Bearer valid-token",
			requestBody: map[string]interface{}{
				"currentPassword": "OldPassword123",
				"newPassword":     "NewPassword123",
			},
			setup: func(mockService *MockUserService) {
				user := &models.User{
					ID:       1,
					Username: "testuser",
				}
				mockService.On("ValidateToken", "valid-token").Return(user, nil)
				mockService.On("ChangePassword", "testuser", "OldPassword123", "NewPassword123").Return(nil)
			},
			expectedStatus: http.StatusOK,
			expectedBody: map[string]interface{}{
				"success": true,
				"message": "密码修改成功",
			},
		},
		{
			name:       "Wrong current password",
			authHeader: "Bearer valid-token",
			requestBody: map[string]interface{}{
				"currentPassword": "WrongPassword",
				"newPassword":     "NewPassword123",
			},
			setup: func(mockService *MockUserService) {
				user := &models.User{
					ID:       1,
					Username: "testuser",
				}
				mockService.On("ValidateToken", "valid-token").Return(user, nil)
				mockService.On("ChangePassword", "testuser", "WrongPassword", "NewPassword123").Return(errors.New("INVALID_PASSWORD"))
			},
			expectedStatus: http.StatusBadRequest,
			expectedBody: map[string]interface{}{
				"error": "当前密码错误",
			},
		},
		{
			name:       "No auth header",
			authHeader: "",
			requestBody: map[string]interface{}{
				"currentPassword": "test",
				"newPassword":     "test123",
			},
			setup:          func(mockService *MockUserService) {},
			expectedStatus: http.StatusUnauthorized,
			expectedBody: map[string]interface{}{
				"error": "未提供认证token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setup(mockService)

			handler := NewUserHandler(mockService)
			router := setupRouter()
			router.POST("/change-password", handler.ChangePassword)

			body, _ := json.Marshal(tt.requestBody)
			req, _ := http.NewRequest("POST", "/change-password", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			if tt.authHeader != "" {
				req.Header.Set("Authorization", tt.authHeader)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			for key, expectedValue := range tt.expectedBody {
				assert.Equal(t, expectedValue, response[key])
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestUserHandler_GetUsers(t *testing.T) {
	tests := []struct {
		name           string
		setup          func(*MockUserService)
		expectedStatus int
		expectedUsers  int
	}{
		{
			name: "Get users successfully",
			setup: func(mockService *MockUserService) {
				users := []*models.User{
					{Username: "user1"},
					{Username: "user2"},
				}
				mockService.On("GetAllUsers").Return(users, nil)
			},
			expectedStatus: http.StatusOK,
			expectedUsers:  2,
		},
		{
			name: "Database error",
			setup: func(mockService *MockUserService) {
				mockService.On("GetAllUsers").Return(nil, errors.New("database error"))
			},
			expectedStatus: http.StatusInternalServerError,
			expectedUsers:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockService := new(MockUserService)
			tt.setup(mockService)

			handler := NewUserHandler(mockService)
			router := setupRouter()
			router.GET("/users", handler.GetUsers)

			req, _ := http.NewRequest("GET", "/users", nil)
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			assert.Equal(t, tt.expectedStatus, w.Code)

			if tt.expectedStatus == http.StatusOK {
				var users []map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &users)
				assert.Len(t, users, tt.expectedUsers)
			}

			mockService.AssertExpectations(t)
		})
	}
}