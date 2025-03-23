package service

import (
	"codefolio/internal/domain"
	"codefolio/internal/repository"
	"errors"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// 定义错误常量
var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserExists         = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInternalError      = errors.New("internal server error")
)

type UserService interface {
	Register(username, email, password string) (*domain.User, string, error)
	Login(email, password string) (*domain.User, string, error)
	GetUserByID(id string) (*domain.User, error)
}

type userService struct {
	userRepo    repository.UserRepository
	jwtSecret   string
	tokenExpiry time.Duration
}

func NewUserService(userRepo repository.UserRepository, jwtSecret string) UserService {
	return &userService{
		userRepo:    userRepo,
		jwtSecret:   jwtSecret,
		tokenExpiry: 24 * time.Hour, // 令牌有效期24小时
	}
}

// Register 注册新用户
func (s *userService) Register(username, email, password string) (*domain.User, string, error) {
	// 检查邮箱是否已存在
	existingUser, err := s.userRepo.FindByEmail(email)
	if err == nil && existingUser != nil {
		return nil, "", ErrUserExists
	}

	// 哈希密码
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, "", ErrInternalError
	}

	// 创建用户
	user := &domain.User{
		Username: username,
		Email:    email,
		Password: string(hashedPassword),
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, "", ErrInternalError
	}

	// 生成JWT令牌
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", ErrInternalError
	}

	return user, token, nil
}

// Login 用户登录
func (s *userService) Login(email, password string) (*domain.User, string, error) {
	// 查找用户
	user, err := s.userRepo.FindByEmail(email)
	if err != nil || user == nil {
		return nil, "", ErrUserNotFound
	}

	// 验证密码
	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", ErrInvalidCredentials
	}

	// 生成JWT令牌
	token, err := s.generateToken(user)
	if err != nil {
		return nil, "", ErrInternalError
	}

	return user, token, nil
}

// GetUserByID 根据ID获取用户
func (s *userService) GetUserByID(id string) (*domain.User, error) {
	userID, err := strconv.ParseUint(id, 10, 32)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	user, err := s.userRepo.FindByID(uint(userID))
	if err != nil || user == nil {
		return nil, ErrUserNotFound
	}
	return user, nil
}

// generateToken 生成JWT令牌
func (s *userService) generateToken(user *domain.User) (string, error) {
	// 设置JWT声明
	claims := jwt.MapClaims{
		"sub": strconv.FormatUint(uint64(user.ID), 10), // subject (用户ID)
		"iat": time.Now().Unix(),                       // issued at (签发时间)
		"exp": time.Now().Add(s.tokenExpiry).Unix(),    // expiry (过期时间)
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名字符串
	tokenString, err := token.SignedString([]byte(s.jwtSecret))
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func (s *userService) UpdateUser(user *domain.User) error {
	return s.userRepo.Update(user)
}
