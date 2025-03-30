package handler

import (
	"codefolio/internal/common"
	"codefolio/internal/service"
	"codefolio/internal/util"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// UniversityHandler 大学处理器
type UniversityHandler struct {
	universityService service.UniversityService
}

// NewUniversityHandler 创建大学处理器
func NewUniversityHandler(universityService service.UniversityService) *UniversityHandler {
	return &UniversityHandler{
		universityService: universityService,
	}
}

// GetUniversities 获取所有大学列表
// @Summary 获取所有大学列表
// @Description 返回所有大学的ID和名称
// @Tags 公共数据
// @Produce json
// @Success 200 {object} common.Response{data=[]domain.University}
// @Failure 500 {object} common.Response
// @Router /api/v1/universities [get]
func (h *UniversityHandler) GetUniversities(c *gin.Context) {
	// 获取所有大学
	universities, err := h.universityService.GetAllUniversities()
	if err != nil {
		util.GetLogger().Error("获取大学列表失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	common.ResponseWithData(c, universities)
}
