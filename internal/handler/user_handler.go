package handler

import (
	"codefolio/internal/domain"
	"codefolio/internal/util"
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
	if !util.BindAndValidate(c, &req) {
		return
	}

	user, err := h.userService.Register(req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err.Error() == "email already exists" {
			ResponseWithError(c, http.StatusBadRequest, USER_ALREADY_EXISTS, "")
			return
		}
		ResponseWithError(c, http.StatusInternalServerError, SERVER_ERROR, err.Error())
		return
	}

	ResponseWithData(c, user)
}

func (h *UserHandler) Login(c *gin.Context) {
	var req LoginRequest
	if !util.BindAndValidate(c, &req) {
		return
	}

	token, err := h.userService.Login(req.Email, req.Password)
	if err != nil {
		ResponseWithError(c, http.StatusUnauthorized, INVALID_CREDENTIALS, "")
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
		ResponseWithError(c, http.StatusNotFound, USER_NOT_FOUND, "")
		return
	}

	ResponseWithData(c, user)
}
