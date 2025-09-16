package service

import (
	"bytes"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/models"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pquerna/otp/totp"
	"golang.org/x/crypto/bcrypt"
	"image/png"
	"regexp"
	"time"
)

type UserService struct {
	userDAO   *dao.UserDAO
	jwtSecret string
}

func NewUserService(userDAO *dao.UserDAO, jwtSecret string) *UserService {
	return &UserService{
		userDAO:   userDAO,
		jwtSecret: jwtSecret,
	}
}

// ValidatePassword 验证密码强度
func (s *UserService) ValidatePassword(password string) error {
	if len(password) < 8 {
		return errors.New("密码长度至少8位")
	}

	if len(password) > 128 {
		return errors.New("密码长度不能超过128位")
	}

	// 检查是否包含大写字母
	hasUpper, _ := regexp.MatchString(`[A-Z]`, password)
	if !hasUpper {
		return errors.New("密码必须包含至少一个大写字母")
	}

	// 检查是否包含小写字母
	hasLower, _ := regexp.MatchString(`[a-z]`, password)
	if !hasLower {
		return errors.New("密码必须包含至少一个小写字母")
	}

	// 检查是否包含数字
	hasDigit, _ := regexp.MatchString(`[0-9]`, password)
	if !hasDigit {
		return errors.New("密码必须包含至少一个数字")
	}

	// 检查是否包含特殊字符
	hasSpecial, _ := regexp.MatchString(`[!@#$%^&*()_+\-=\[\]{};':"\\|,.<>\/?]`, password)
	if !hasSpecial {
		return errors.New("密码必须包含至少一个特殊字符")
	}

	return nil
}

// GetAllUsers 获取所有用户
func (s *UserService) GetAllUsers() ([]*models.User, error) {
	return s.userDAO.GetAll()
}

// GetUserByUsername 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*models.User, error) {
	return s.userDAO.GetByUsername(username)
}

// GetUserByEmail 根据邮箱获取用户
func (s *UserService) GetUserByEmail(email string) (*models.User, error) {
	return s.userDAO.GetByEmail(email)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uint) (*models.User, error) {
	return s.userDAO.GetByID(id)
}

// CreateUser 创建用户
func (s *UserService) CreateUser(username, email, password string) error {
	// 验证密码强度
	if err := s.ValidatePassword(password); err != nil {
		return err
	}

	// 检查用户名是否已存在
	if existingUser, _ := s.userDAO.GetByUsername(username); existingUser != nil {
		return errors.New("用户名已存在")
	}

	// 检查邮箱是否已存在
	if existingUser, _ := s.userDAO.GetByEmail(email); existingUser != nil {
		return errors.New("邮箱已被使用")
	}

	// 生成密码哈希
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hashedPassword),
		Role:         "user",
	}

	return s.userDAO.Create(user)
}

// DeleteUser 删除用户
func (s *UserService) DeleteUser(id uint) error {
	return s.userDAO.Delete(id)
}

// Login 用户登录
func (s *UserService) Login(username, password string) (string, error) {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		return "", errors.New("INVALID_CREDENTIALS")
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return "", errors.New("INVALID_CREDENTIALS")
	}

	// 生成JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"id":       user.ID,
		"exp":      time.Now().Add(time.Hour * 24).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

// ValidateToken 验证JWT token
func (s *UserService) ValidateToken(tokenString string) (*models.User, error) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		username, ok := claims["username"].(string)
		if !ok {
			return nil, errors.New("invalid token claims")
		}

		user, err := s.userDAO.GetByUsername(username)
		if err != nil {
			// 用户不存在时返回无效token错误，而不是数据库错误
			return nil, errors.New("invalid token: user not found")
		}

		return user, nil
	}

	return nil, errors.New("invalid token")
}

// Setup2FA 设置双因子认证
func (s *UserService) Setup2FA(username string) (string, []byte, error) {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		return "", nil, err
	}

	// 生成TOTP密钥
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CyberEdge",
		AccountName: user.Username,
	})
	if err != nil {
		return "", nil, err
	}

	// 保存密钥到数据库，但暂时不启用2FA
	user.TOTPSecret = key.Secret()
	user.Is2FAEnabled = false // 仅在验证成功后启用
	if err := s.userDAO.Update(user); err != nil {
		return "", nil, err
	}

	// 生成二维码
	img, err := key.Image(256, 256)
	if err != nil {
		return "", nil, err
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		return "", nil, err
	}

	return key.Secret(), buf.Bytes(), nil
}

// Verify2FA 验证双因子认证，验证成功后启用2FA
func (s *UserService) Verify2FA(username, code string) error {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		return err
	}

	if user.TOTPSecret == "" {
		return errors.New("请先设置双因子认证")
	}

	valid := totp.Validate(code, user.TOTPSecret)
	if !valid {
		return errors.New("验证码无效")
	}

	// 验证成功后才启用2FA
	if !user.Is2FAEnabled {
		user.Is2FAEnabled = true
		if err := s.userDAO.Update(user); err != nil {
			return err
		}
	}

	return nil
}

// Disable2FA 禁用双因子认证
func (s *UserService) Disable2FA(username string) error {
	user, err := s.userDAO.GetByUsername(username)
	if err != nil {
		return err
	}

	user.Is2FAEnabled = false
	user.TOTPSecret = ""
	return s.userDAO.Update(user)
}