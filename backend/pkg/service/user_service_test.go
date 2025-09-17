package service

import (
	"cyberedge/pkg/models"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

// MockUserDAO 模拟UserDAO
type MockUserDAO struct {
	mock.Mock
}

func (m *MockUserDAO) Create(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserDAO) GetByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDAO) GetByEmail(email string) (*models.User, error) {
	args := m.Called(email)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDAO) GetByID(id uint) (*models.User, error) {
	args := m.Called(id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

func (m *MockUserDAO) Update(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

func (m *MockUserDAO) Delete(id uint) error {
	args := m.Called(id)
	return args.Error(0)
}

func (m *MockUserDAO) GetAll() ([]*models.User, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*models.User), args.Error(1)
}

func TestValidatePassword(t *testing.T) {
	service := &UserService{}

	tests := []struct {
		name     string
		password string
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Valid password",
			password: "Password123!!",
			wantErr:  false,
		},
		{
			name:     "Too short",
			password: "Pass1",
			wantErr:  true,
			errMsg:   "密码长度至少8位",
		},
		{
			name:     "Too long",
			password: "Password123!" + string(make([]byte, 120)),
			wantErr:  true,
			errMsg:   "密码长度不能超过128位",
		},
		{
			name:     "No uppercase",
			password: "password123",
			wantErr:  true,
			errMsg:   "密码必须包含至少一个大写字母",
		},
		{
			name:     "No lowercase",
			password: "PASSWORD123",
			wantErr:  true,
			errMsg:   "密码必须包含至少一个小写字母",
		},
		{
			name:     "No numbers",
			password: "PasswordABC",
			wantErr:  true,
			errMsg:   "密码必须包含至少一个数字",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrWeakPassword)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCreateUser(t *testing.T) {
	mockDAO := new(MockUserDAO)
	service := &UserService{
		userDAO:   mockDAO,
		jwtSecret: "test-secret",
	}

	tests := []struct {
		name        string
		username    string
		email       string
		password    string
		setup       func()
		wantErr     bool
		expectedErr error
	}{
		{
			name:     "Valid user creation",
			username: "testuser",
			email:    "test@example.com",
			password: "Password123!",
			setup: func() {
				mockDAO.On("GetByUsername", "testuser").Return(nil, errors.New("user not found"))
				mockDAO.On("GetByEmail", "test@example.com").Return(nil, errors.New("user not found"))
				mockDAO.On("Create", mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:        "Username already exists",
			username:    "existinguser",
			email:       "new@example.com",
			password:    "Password123!",
			setup: func() {
				existingUser := &models.User{Username: "existinguser"}
				mockDAO.On("GetByUsername", "existinguser").Return(existingUser, nil)
			},
			wantErr:     true,
			expectedErr: ErrUserExists,
		},
		{
			name:        "Email already exists",
			username:    "newuser",
			email:       "existing@example.com",
			password:    "Password123!",
			setup: func() {
				mockDAO.On("GetByUsername", "newuser").Return(nil, errors.New("user not found"))
				existingUser := &models.User{Email: "existing@example.com"}
				mockDAO.On("GetByEmail", "existing@example.com").Return(existingUser, nil)
			},
			wantErr:     true,
			expectedErr: ErrEmailExists,
		},
		{
			name:        "Invalid password",
			username:    "testuser2",
			email:       "test2@example.com",
			password:    "weak",
			setup:       func() {},
			wantErr:     true,
			expectedErr: ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDAO.ExpectedCalls = nil
			mockDAO.Calls = nil

			tt.setup()

			err := service.CreateUser(tt.username, tt.email, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tt.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mockDAO.AssertExpectations(t)
		})
	}
}

func TestLogin(t *testing.T) {
	mockDAO := new(MockUserDAO)
	service := &UserService{
		userDAO:   mockDAO,
		jwtSecret: "test-secret",
	}

	// 创建测试用户密码哈希
	hashedPassword, _ := bcrypt.GenerateFromPassword([]byte("Password123!"), bcrypt.DefaultCost)

	tests := []struct {
		name     string
		username string
		password string
		setup    func()
		wantErr  bool
		errMsg   string
	}{
		{
			name:     "Valid login",
			username: "testuser",
			password: "Password123!",
			setup: func() {
				user := &models.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "user",
					CreatedAt:    time.Now().Unix(),
				}
				mockDAO.On("GetByUsername", "testuser").Return(user, nil)
			},
			wantErr: false,
		},
		{
			name:     "User not found",
			username: "nonexistent",
			password: "Password123!",
			setup: func() {
				mockDAO.On("GetByUsername", "nonexistent").Return(nil, errors.New("user not found"))
			},
			wantErr: true,
			errMsg:  "INVALID_CREDENTIALS",
		},
		{
			name:     "Wrong password",
			username: "testuser",
			password: "WrongPassword",
			setup: func() {
				user := &models.User{
					ID:           1,
					Username:     "testuser",
					Email:        "test@example.com",
					PasswordHash: string(hashedPassword),
					Role:         "user",
					CreatedAt:    time.Now().Unix(),
				}
				mockDAO.On("GetByUsername", "testuser").Return(user, nil)
			},
			wantErr: true,
			errMsg:  "INVALID_CREDENTIALS",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDAO.ExpectedCalls = nil
			mockDAO.Calls = nil

			tt.setup()

			token, err := service.Login(tt.username, tt.password)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, token)
			}

			mockDAO.AssertExpectations(t)
		})
	}
}

func TestChangePassword(t *testing.T) {
	mockDAO := new(MockUserDAO)
	service := &UserService{
		userDAO:   mockDAO,
		jwtSecret: "test-secret",
	}

	// 创建当前密码哈希
	currentPasswordHash, _ := bcrypt.GenerateFromPassword([]byte("CurrentPassword123!!"), bcrypt.DefaultCost)

	tests := []struct {
		name            string
		username        string
		currentPassword string
		newPassword     string
		setup           func()
		wantErr         bool
		expectedErr     error
	}{
		{
			name:            "Valid password change",
			username:        "testuser",
			currentPassword: "CurrentPassword123!!",
			newPassword:     "NewPassword123!!",
			setup: func() {
				user := &models.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(currentPasswordHash),
				}
				mockDAO.On("GetByUsername", "testuser").Return(user, nil)
				mockDAO.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:            "User not found",
			username:        "nonexistent",
			currentPassword: "CurrentPassword123!!",
			newPassword:     "NewPassword123!!",
			setup: func() {
				mockDAO.On("GetByUsername", "nonexistent").Return(nil, errors.New("user not found"))
			},
			wantErr: true,
		},
		{
			name:            "Wrong current password",
			username:        "testuser",
			currentPassword: "WrongPassword",
			newPassword:     "NewPassword123!!",
			setup: func() {
				user := &models.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(currentPasswordHash),
				}
				mockDAO.On("GetByUsername", "testuser").Return(user, nil)
			},
			wantErr:     true,
			expectedErr: ErrInvalidPassword,
		},
		{
			name:            "Invalid new password",
			username:        "testuser",
			currentPassword: "CurrentPassword123!!",
			newPassword:     "weak",
			setup: func() {
				user := &models.User{
					ID:           1,
					Username:     "testuser",
					PasswordHash: string(currentPasswordHash),
				}
				mockDAO.On("GetByUsername", "testuser").Return(user, nil)
			},
			wantErr:     true,
			expectedErr: ErrWeakPassword,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDAO.ExpectedCalls = nil
			mockDAO.Calls = nil

			tt.setup()

			err := service.ChangePassword(tt.username, tt.currentPassword, tt.newPassword)

			if tt.wantErr {
				assert.Error(t, err)
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr)
				}
			} else {
				assert.NoError(t, err)
			}

			mockDAO.AssertExpectations(t)
		})
	}
}

func TestSetup2FA(t *testing.T) {
	mockDAO := new(MockUserDAO)
	service := &UserService{
		userDAO:   mockDAO,
		jwtSecret: "test-secret",
	}

	tests := []struct {
		name     string
		username string
		setup    func()
		wantErr  bool
	}{
		{
			name:     "Valid 2FA setup",
			username: "testuser",
			setup: func() {
				user := &models.User{
					ID:       1,
					Username: "testuser",
					Email:    "test@example.com",
				}
				mockDAO.On("GetByUsername", "testuser").Return(user, nil)
				mockDAO.On("Update", mock.AnythingOfType("*models.User")).Return(nil)
			},
			wantErr: false,
		},
		{
			name:     "User not found",
			username: "nonexistent",
			setup: func() {
				mockDAO.On("GetByUsername", "nonexistent").Return(nil, errors.New("user not found"))
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset mock
			mockDAO.ExpectedCalls = nil
			mockDAO.Calls = nil

			tt.setup()

			secret, qrCode, err := service.Setup2FA(tt.username)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Empty(t, secret)
				assert.Nil(t, qrCode)
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, secret)
				assert.NotNil(t, qrCode)
			}

			mockDAO.AssertExpectations(t)
		})
	}
}

func TestValidateToken(t *testing.T) {
	mockDAO := new(MockUserDAO)
	service := &UserService{
		userDAO:   mockDAO,
		jwtSecret: "test-secret",
	}

	tests := []struct {
		name    string
		token   string
		wantErr bool
	}{
		{
			name:    "Invalid token",
			token:   "invalid.token.here",
			wantErr: true,
		},
		{
			name:    "Empty token",
			token:   "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			user, err := service.ValidateToken(tt.token)

			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, user)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, user)
			}
		})
	}
}