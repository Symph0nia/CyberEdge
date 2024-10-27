// pkg/service/user_service.go

package service

import (
	"bytes"
	"crypto/rand"
	"cyberedge/pkg/dao"
	"cyberedge/pkg/logging"
	"cyberedge/pkg/models"
	"errors"
	"fmt"
	"github.com/golang-jwt/jwt"
	"github.com/pquerna/otp/totp"
	"image/png"
	"math/big"
	"strings"
	"time"
)

type UserService struct {
	userDAO   *dao.UserDAO
	jwtSecret string
}

func NewUserService(userDAO *dao.UserDAO, jwtSecret string) *UserService {
	return &UserService{userDAO: userDAO, jwtSecret: jwtSecret}
}

// User management methods

func (s *UserService) GetAllUsers() ([]models.User, error) {
	logging.Info("正在获取所有用户")
	users, err := s.userDAO.GetAllUsers()
	if err != nil {
		logging.Error("获取所有用户失败: %v", err)
		return nil, err
	}
	logging.Info("成功获取所有用户，共 %d 个", len(users))
	return users, nil
}

func (s *UserService) GetUserByAccount(account string) (*models.User, error) {
	logging.Info("正在获取用户: %s", account)
	user, err := s.userDAO.GetUserByAccount(account)
	if err != nil {
		logging.Error("获取用户失败: %s, 错误: %v", account, err)
		return nil, err
	}
	logging.Info("成功获取用户: %s", account)
	return user, nil
}

func (s *UserService) CreateUser(user *models.User) error {
	logging.Info("正在创建新用户: %s", user.Account)
	err := s.userDAO.CreateUser(user)
	if err != nil {
		logging.Error("创建用户失败: %s, 错误: %v", user.Account, err)
		return err
	}
	logging.Info("成功创建用户: %s", user.Account)
	return nil
}

func (s *UserService) DeleteUser(account string) error {
	logging.Info("正在删除用户: %s", account)
	err := s.userDAO.DeleteUser(account)
	if err != nil {
		logging.Error("删除用户失败: %s, 错误: %v", account, err)
		return err
	}
	logging.Info("成功删除用户: %s", account)
	return nil
}

// Authentication methods

func (s *UserService) GenerateQRCode() ([]byte, error) {
	logging.Info("开始生成二维码")

	qrcodeEnabled, err := s.userDAO.GetQRCodeStatus()
	if err != nil {
		logging.Error("获取二维码状态失败: %v", err)
		return nil, errors.New("无法获取二维码状态")
	}

	if !qrcodeEnabled {
		logging.Warn("二维码接口已关闭")
		return nil, errors.New("二维码接口已关闭")
	}

	accountName, err := generateRandomString(16)
	if err != nil {
		logging.Error("生成账户名称失败: %v", err)
		return nil, errors.New("无法生成账户名称")
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      "CyberEdgeAdmin",
		AccountName: accountName,
	})
	if err != nil {
		logging.Error("生成TOTP密钥失败: %v", err)
		return nil, errors.New("无法生成密钥")
	}

	newUser := &models.User{
		Account: accountName,
		Secret:  key.Secret(),
	}
	err = s.CreateUser(newUser)
	if err != nil {
		logging.Error("存储用户信息失败: %v", err)
		return nil, errors.New("无法存储密钥")
	}

	img, err := key.Image(200, 200)
	if err != nil {
		logging.Error("生成二维码图像失败: %v", err)
		return nil, errors.New("无法生成二维码")
	}

	var buf bytes.Buffer
	if err := png.Encode(&buf, img); err != nil {
		logging.Error("编码二维码图像失败: %v", err)
		return nil, errors.New("无法编码二维码")
	}

	logging.Info("二维码生成成功")
	return buf.Bytes(), nil
}

func (s *UserService) ValidateTOTP(code, account string) (string, int, error) {
	logging.Info("开始验证TOTP: 账户 %s", account)

	user, err := s.GetUserByAccount(account)
	if err != nil {
		logging.Error("获取用户信息失败: %v", err)
		return "", 0, errors.New("无法找到密钥")
	}

	if !totp.Validate(code, user.Secret) {
		logging.Warn("TOTP验证失败: 账户 %s", account)
		return "", 0, errors.New("验证码无效")
	}

	newLoginCount, err := s.userDAO.IncrementLoginCount(account)
	if err != nil {
		logging.Error("更新登录次数失败: %v", err)
		return "", 0, errors.New("无法更新登录次数")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"account": account,
		"exp":     time.Now().Add(time.Hour * 72).Unix(),
	})

	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		logging.Error("生成JWT令牌失败: %v", err)
		return "", 0, errors.New("无法生成令牌")
	}

	logging.Info("TOTP验证成功: 账户 %s", account)
	return tokenString, newLoginCount, nil
}

func (s *UserService) CheckAuth(tokenString string) (bool, string, error) {
	logging.Info("开始验证JWT令牌")

	if tokenString == "" {
		logging.Warn("未提供JWT令牌")
		return false, "", errors.New("未提供令牌")
	}

	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	if tokenString == "" {
		logging.Warn("无效的JWT令牌格式")
		return false, "", errors.New("无效的令牌格式")
	}

	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.jwtSecret), nil
	})

	if err != nil || !token.Valid {
		logging.Warn("无效的JWT令牌")
		return false, "", errors.New("无效的令牌")
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		logging.Warn("无法解析JWT令牌声明")
		return false, "", errors.New("无效的令牌")
	}

	account, ok := claims["account"].(string)
	if !ok {
		logging.Warn("JWT令牌中缺少账户信息")
		return false, "", errors.New("无效的令牌")
	}

	logging.Info("JWT令牌验证成功: 账户 %s", account)
	return true, account, nil
}

// Helper function
func generateRandomString(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("无法生成随机字符串: %w", err)
		}
		result[i] = charset[num.Int64()]
	}
	return string(result), nil
}
