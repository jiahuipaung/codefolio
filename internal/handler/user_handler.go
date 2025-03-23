package handler

import (
	"codefolio/internal/domain"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UserHandler struct {
	userService domain.UserService
}

func NewUserHandler(userService domain.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=6"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
}

type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

func (h *UserHandler) Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseWithError(c, http.StatusBadRequest, INVALID_PARAMS, err.Error())
		return
	}

	user, err := h.userService.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		ResponseWithError(c, http.StatusBadRequest, ERROR, err.Error())
		return
	}

	ResponseWithData(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		ResponseWithError(c, http.StatusBadRequest, INVALID_PARAMS, err.Error())
		return
	}

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		ResponseWithError(c, http.StatusUnauthorized, UNAUTHORIZED, "无效的凭据")
		return
	}

	ResponseWithData(c, gin.H{"token": token})
}

func (h *UserHandler) GetMe(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		ResponseWithError(c, http.StatusUnauthorized, UNAUTHORIZED, "")
		return
	}

	user, err := h.userService.GetUserByID(userID.(uint))
	if err != nil {
		ResponseWithError(c, http.StatusNotFound, NOT_FOUND, "用户不存在")
		return
	}

	ResponseWithData(c, user)
}
