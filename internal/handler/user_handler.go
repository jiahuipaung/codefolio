package handler

import (
	"codefolio/internal/common"
	"codefolio/internal/service"
	"codefolio/internal/util"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
	Username string `json:"username" binding:"required,min=3,max=50"`
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

	user, err := h.userService.Register(req.Username, req.Password, req.Email)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			common.ResponseWithError(c, common.CodeUserAlreadyExists)
		default:
			util.GetLogger().Error("注册失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 登录获取token
	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		util.GetLogger().Error("注册后登录失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
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

	token, err := h.userService.Login(req.Username, req.Password)
	if err != nil {
		switch err {
		case service.ErrInvalidCredentials:
			common.ResponseWithError(c, common.CodeInvalidCredentials)
		case service.ErrUserNotFound:
			common.ResponseWithError(c, common.CodeUserNotFound)
		default:
			util.GetLogger().Error("登录失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 获取用户信息
	userIDFromToken, err := service.ExtractUserIDFromToken(token, h.userService)
	if err != nil {
		util.GetLogger().Error("从令牌获取用户ID失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	user, err := h.userService.GetUserByID(userIDFromToken)
	if err != nil {
		util.GetLogger().Error("获取用户信息失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
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

	// 将userID转换为uint
	userIDStr, ok := userID.(string)
	if !ok {
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	idUint, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	user, err := h.userService.GetUserByID(uint(idUint))
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
