package service

import (
	"codefolio/internal/common"
	"codefolio/internal/domain"
	"codefolio/internal/repository"
	"errors"
	"fmt"
	"time"

	"github.com/dgrijalva/jwt-go"
	"golang.org/x/crypto/bcrypt"
)

var (
	// ErrInvalidCredentials 无效的凭证
	ErrInvalidCredentials = errors.New("无效的用户名或密码")
	// ErrUserNotFound 用户不存在
	ErrUserNotFound = errors.New("用户不存在")
	// ErrUserAlreadyExists 用户已存在
	ErrUserAlreadyExists = errors.New("用户已存在")
)

// AuthClaims JWT声明结构
type AuthClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

// UserService 用户服务接口
type UserService interface {
	Register(username, password, email string) (*domain.User, error)
	Login(username, password string) (string, error)
	GetUserByID(id uint) (*domain.User, error)
}

// userService 用户服务实现
type userService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	expireHours int
}

// NewUserService 创建用户服务
func NewUserService(userRepo repository.UserRepository, jwtSecret string, expireHours int) UserService {
	return &userService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		expireHours: expireHours,
	}
}

// Register 用户注册
func (s *userService) Register(username, password, email string) (*domain.User, error) {
	// 检查用户是否已存在
	existingUser, err := s.userRepo.FindByUsername(username)
	if err != nil && !errors.Is(err, common.ErrRecordNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, ErrUserAlreadyExists
	}

	// 检查邮箱是否已存在
	existingEmail, err := s.userRepo.FindByEmail(email)
	if err != nil && !errors.Is(err, common.ErrRecordNotFound) {
		return nil, err
	}
	if existingEmail != nil {
		return nil, errors.New("邮箱已被注册")
	}

	// 加密密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("密码加密失败: %w", err)
	}

	user := &domain.User{
		Username: username,
		Password: string(hashedPassword),
		Email:    email,
	}

	// 保存用户
	err = s.userRepo.Create(user)
	if err != nil {
		return nil, err
	}

	return user, nil
}

// Login 用户登录
func (s *userService) Login(username, password string) (string, error) {
	// 查找用户
	user, err := s.userRepo.FindByUsername(username)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return "", ErrInvalidCredentials
		}
		return "", err
	}

	// 验证密码
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if err != nil {
		return "", ErrInvalidCredentials
	}

	// 生成JWT token
	token, err := s.generateToken(user.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id uint) (*domain.User, error) {
	user, err := s.userRepo.FindByID(id)
	if err != nil {
		if errors.Is(err, common.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}

// generateToken 生成JWT令牌
func (s *userService) generateToken(userID uint) (string, error) {
	// 设置过期时间
	expirationTime := time.Now().Add(time.Duration(s.expireHours) * time.Hour)

	// 创建声明
	claims := &AuthClaims{
		UserID: fmt.Sprintf("%d", userID),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: expirationTime.Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) UpdateUser(user *domain.User) error {
	return s.userRepo.Update(user)
}
