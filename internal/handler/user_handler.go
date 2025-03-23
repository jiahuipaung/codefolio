package handler

import (
	"codefolio/internal/common"
	"codefolio/internal/service"
	"codefolio/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
)

// UserHandler 处理用户相关的HTTP请求
type UserHandler struct {
	userService service.UserService
}

// NewUserHandler 创建UserHandler实例
func NewUserHandler(userService service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// RegisterRequest 用户注册请求结构
type RegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=6"`
}

// LoginRequest 用户登录请求结构
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UserResponse 用户信息响应结构
type UserResponse struct {
	ID       uint   `json:"id"`
	Username string `json:"username"`
	Email    string `json:"email"`
}

// AuthResponse 认证响应结构
type AuthResponse struct {
	Token string       `json:"token"`
	User  UserResponse `json:"user"`
}

// Register 处理用户注册请求
func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		return // 错误已在BindAndValidate中处理
	}

	user, token, err := h.userService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrUserExists:
			common.ResponseWithError(c, common.CodeUserAlreadyExists)
		default:
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	common.ResponseWithData(c, AuthResponse{
		Token: token,
		User: UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// Login 处理用户登录请求
func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := util.BindAndValidate(c, &req); err != nil {
		return // 错误已在BindAndValidate中处理
	}

	user, token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			common.ResponseWithError(c, common.CodeInvalidCredentials)
		case service.ErrUserNotFound:
			common.ResponseWithError(c, common.CodeUserNotFound)
		default:
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	common.ResponseWithData(c, AuthResponse{
		Token: token,
		User: UserResponse{
			ID:       user.ID,
			Username: user.Username,
			Email:    user.Email,
		},
	})
}

// GetMe 获取当前认证用户信息
func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("userID")
	if !exists {
		common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
		return
	}

	user, err := h.userService.GetUserByID(userID.(string))
	if err != nil {
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	common.ResponseWithData(c, UserResponse{
		ID:       user.ID,
		Username: user.Username,
		Email:    user.Email,
	})
}
