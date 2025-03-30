package handler

import (
	"codefolio/internal/common"
	"codefolio/internal/domain"
	"codefolio/internal/service"
	"codefolio/internal/util"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
)

// ResumeHandler 简历处理器
type ResumeHandler struct {
	resumeService service.ResumeService
}

// NewResumeHandler 创建简历处理器
func NewResumeHandler(resumeService service.ResumeService) *ResumeHandler {
	return &ResumeHandler{
		resumeService: resumeService,
	}
}

// UploadPDFRequest PDF上传请求
type UploadPDFRequest struct {
	// 无需附加字段，仅包含文件
}

// UploadPDFResponse PDF上传响应
type UploadPDFResponse struct {
	ImageURL string `json:"image_url"` // 转换后的图片URL
	FileKey  string `json:"file_key"`  // 文件标识，用于后续创建简历时关联
}

// CreateResumeRequest 创建简历请求
type CreateResumeRequest struct {
	FileKey     string   `json:"file_key" binding:"required"`   // 上传PDF时返回的文件标识
	Role        string   `json:"role" binding:"required"`       // 应聘职位
	Level       string   `json:"level" binding:"required"`      // 经历等级
	University  string   `json:"university" binding:"required"` // 毕业院校
	PassCompany []string `json:"pass_company"`                  // 面试通过的公司
}

// UploadResumeRequest 上传简历请求（兼容旧接口）
type UploadResumeRequest struct {
	Role        string   `form:"role" binding:"required"`       // 应聘职位
	Level       string   `form:"level" binding:"required"`      // 经历等级
	University  string   `form:"university" binding:"required"` // 毕业院校
	PassCompany []string `form:"pass_company[]"`                // 面试通过的公司
}

// UpdateResumeRequest 更新简历请求
type UpdateResumeRequest struct {
	Role        string   `json:"role"`
	Level       string   `json:"level"`
	University  string   `json:"university"`
	PassCompany []string `json:"pass_company"`
}

// ResumeResponse 简历响应
type ResumeResponse struct {
	ID          uint     `json:"id"`
	UserID      uint     `json:"user_id"`
	ImageURL    string   `json:"image_url"`
	Role        string   `json:"role"`
	Level       string   `json:"level"`
	University  string   `json:"university"`
	PassCompany []string `json:"pass_company"`
	CreatedAt   string   `json:"created_at"`
	UpdatedAt   string   `json:"updated_at"`
}

// GetPagingParams 获取分页参数
func GetPagingParams(c *gin.Context) (page, size int) {
	pageStr := c.DefaultQuery("page", "1")
	sizeStr := c.DefaultQuery("size", "10")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	size, err = strconv.Atoi(sizeStr)
	if err != nil || size < 1 || size > 100 {
		size = 10
	}

	return page, size
}

// getCurrentUserID 获取当前用户ID
func getCurrentUserID(c *gin.Context) uint {
	userID, exists := c.Get("userID")
	if !exists {
		return 0
	}

	// 将userID转换为uint
	userIDStr, ok := userID.(string)
	if !ok {
		return 0
	}

	id, err := strconv.ParseUint(userIDStr, 10, 32)
	if err != nil {
		return 0
	}

	return uint(id)
}

// convertToResumeResponse 转换为简历响应
func (h *ResumeHandler) convertToResumeResponse(resume *domain.Resume) ResumeResponse {
	return ResumeResponse{
		ID:          resume.ID,
		UserID:      resume.UserID,
		ImageURL:    resume.ImageURL,
		Role:        resume.Role,
		Level:       resume.Level,
		University:  resume.University,
		PassCompany: resume.PassCompany,
		CreatedAt:   resume.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:   resume.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// UploadPDF 上传简历PDF文件（第一步）
// @Summary 上传简历PDF文件
// @Description 仅上传简历PDF文件并转换为图片，返回图片URL和文件标识，供前端预览和后续创建简历使用
// @Tags 简历
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "简历文件(PDF)"
// @Success 200 {object} common.Response{data=UploadPDFResponse}
// @Failure 400,500 {object} common.Response
// @Router /api/v1/resumes/upload-pdf [post]
func (h *ResumeHandler) UploadPDF(c *gin.Context) {
	// 获取当前用户ID，开发阶段允许匿名上传，默认用户ID为1
	userID := getCurrentUserID(c)
	if userID == 0 {
		// 开发阶段，使用默认用户ID
		userID = 1
	}

	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		util.GetLogger().Error("获取上传文件失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 上传并转换PDF
	fileResult, err := h.resumeService.UploadAndConvertPDF(c, userID, file)
	if err != nil {
		switch err {
		case util.ErrFileTooLarge:
			common.ResponseWithError(c, common.CodeInvalidParams)
		case util.ErrInvalidFileType:
			common.ResponseWithError(c, common.CodeInvalidParams)
		case util.ErrCommandNotFound:
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
			util.GetLogger().Error("PDF转图片工具不可用", zap.Error(err))
		case util.ErrConvertPDFFailed:
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
			util.GetLogger().Error("PDF转图片失败", zap.Error(err))
		default:
			util.GetLogger().Error("上传PDF失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 构建响应 - fileResult.FilePath 现在已经是完整的URL
	resp := UploadPDFResponse{
		ImageURL: fileResult.FilePath,
		FileKey:  fileResult.FileKey,
	}

	common.ResponseWithData(c, resp)
}

// CreateResume 创建简历（第二步）
// @Summary 创建简历
// @Description 使用已上传的PDF文件创建简历（需要先调用上传PDF接口）
// @Tags 简历
// @Accept json
// @Produce json
// @Param request body CreateResumeRequest true "创建简历请求"
// @Success 200 {object} common.Response{data=ResumeResponse}
// @Failure 400,401,500 {object} common.Response
// @Router /api/v1/resumes/create [post]
func (h *ResumeHandler) CreateResume(c *gin.Context) {
	// 获取当前用户ID，开发阶段允许匿名创建，默认用户ID为1
	userID := getCurrentUserID(c)
	if userID == 0 {
		// 开发阶段，使用默认用户ID
		userID = 1
	}

	// 绑定参数
	var req CreateResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.GetLogger().Error("绑定请求参数失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 创建简历
	resume, err := h.resumeService.CreateResumeWithFileKey(
		userID,
		req.FileKey,
		req.Role,
		req.Level,
		req.University,
		req.PassCompany,
	)

	if err != nil {
		switch err {
		case service.ErrFileNotFound:
			common.ResponseWithError(c, common.CodeInvalidParams, http.StatusBadRequest)
		default:
			util.GetLogger().Error("创建简历失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 转换为响应结构
	resp := h.convertToResumeResponse(resume)

	common.ResponseWithData(c, resp)
}

// UploadResume 上传简历（兼容旧接口）
// @Summary 上传简历
// @Description 上传简历文件并转换为图片
// @Tags 简历
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "简历文件(PDF)"
// @Param role formData string true "应聘职位"
// @Param level formData string true "经历等级"
// @Param university formData string true "毕业院校"
// @Param pass_company[] formData []string false "面试通过的公司"
// @Success 200 {object} common.Response{data=ResumeResponse}
// @Failure 400,401,500 {object} common.Response
// @Router /api/v1/resumes [post]
// @Security BearerAuth
func (h *ResumeHandler) UploadResume(c *gin.Context) {
	// 获取当前用户ID
	userID := getCurrentUserID(c)
	if userID == 0 {
		common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
		return
	}

	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		util.GetLogger().Error("获取上传文件失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 绑定参数
	var req UploadResumeRequest
	if err := c.ShouldBind(&req); err != nil {
		util.GetLogger().Error("绑定请求参数失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 创建简历
	resume, err := h.resumeService.CreateResume(
		c,
		userID,
		file,
		req.Role,
		req.Level,
		req.University,
		req.PassCompany,
	)

	if err != nil {
		switch err {
		case util.ErrFileTooLarge:
			common.ResponseWithError(c, common.CodeInvalidParams)
		case util.ErrInvalidFileType:
			common.ResponseWithError(c, common.CodeInvalidParams)
		default:
			util.GetLogger().Error("创建简历失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 转换为响应结构
	resp := h.convertToResumeResponse(resume)

	common.ResponseWithData(c, resp)
}

// GetResume 获取简历详情
// @Summary 获取简历详情
// @Description 根据ID获取简历详情
// @Tags 简历
// @Produce json
// @Param id path int true "简历ID"
// @Success 200 {object} common.Response{data=ResumeResponse}
// @Failure 400,404,500 {object} common.Response
// @Router /api/v1/resumes/{id} [get]
func (h *ResumeHandler) GetResume(c *gin.Context) {
	// 获取简历ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 获取当前用户ID
	userID := getCurrentUserID(c)

	// 获取简历
	resume, err := h.resumeService.GetResumeByID(c, uint(id), userID)
	if err != nil {
		switch err {
		case service.ErrResumeNotFound:
			common.ResponseWithError(c, common.CodeDataNotFound)
		case service.ErrViewLimitExceeded:
			common.ResponseWithError(c, common.CodeOperationNotAllowed)
		default:
			util.GetLogger().Error("获取简历失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 转换为响应结构
	resp := h.convertToResumeResponse(resume)

	common.ResponseWithData(c, resp)
}

// GetResumes 获取简历列表
// @Summary 获取简历列表
// @Description 获取所有简历列表，支持分页和筛选
// @Tags 简历
// @Produce json
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Param role query string false "按职位筛选"
// @Param level query string false "按经历等级筛选"
// @Param university query string false "按毕业院校筛选"
// @Success 200 {object} common.Response{data=[]ResumeResponse}
// @Failure 400,500 {object} common.Response
// @Router /api/v1/resumes [get]
func (h *ResumeHandler) GetResumes(c *gin.Context) {
	// 获取分页参数
	page, size := GetPagingParams(c)

	// 获取筛选参数
	role := c.DefaultQuery("role", "")
	level := c.DefaultQuery("level", "")
	university := c.DefaultQuery("university", "")

	// 获取当前用户ID
	userID := getCurrentUserID(c)

	// 获取简历列表
	resumes, total, err := h.resumeService.GetAllResumes(page, size, role, level, university, userID)
	if err != nil {
		switch err {
		case service.ErrViewLimitExceeded:
			common.ResponseWithError(c, common.CodeOperationNotAllowed)
		default:
			util.GetLogger().Error("获取简历列表失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 转换为响应结构
	var respList []ResumeResponse
	for _, resume := range resumes {
		respList = append(respList, h.convertToResumeResponse(&resume))
	}

	// 构建分页响应
	common.ResponseWithData(c, gin.H{
		"items": respList,
		"total": total,
		"page":  page,
		"size":  size,
	})
}

// GetUserResumes 获取用户简历列表
// @Summary 获取用户简历列表
// @Description 获取当前用户的所有简历
// @Tags 简历
// @Produce json
// @Success 200 {object} common.Response{data=[]ResumeResponse}
// @Failure 401,500 {object} common.Response
// @Router /api/v1/user/resumes [get]
// @Security BearerAuth
func (h *ResumeHandler) GetUserResumes(c *gin.Context) {
	// 获取当前用户ID
	userID := getCurrentUserID(c)
	if userID == 0 {
		common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
		return
	}

	// 获取用户简历
	resumes, err := h.resumeService.GetUserResumes(userID)
	if err != nil {
		util.GetLogger().Error("获取用户简历失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	// 转换为响应结构
	var respList []ResumeResponse
	for _, resume := range resumes {
		respList = append(respList, h.convertToResumeResponse(&resume))
	}

	common.ResponseWithData(c, respList)
}

// UpdateResume 更新简历信息
// @Summary 更新简历信息
// @Description 更新简历基本信息
// @Tags 简历
// @Accept json
// @Produce json
// @Param id path int true "简历ID"
// @Param request body UpdateResumeRequest true "更新请求"
// @Success 200 {object} common.Response{data=ResumeResponse}
// @Failure 400,401,403,404,500 {object} common.Response
// @Router /api/v1/resumes/{id} [put]
// @Security BearerAuth
func (h *ResumeHandler) UpdateResume(c *gin.Context) {
	// 获取简历ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 获取当前用户ID
	userID := getCurrentUserID(c)
	if userID == 0 {
		common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
		return
	}

	// 绑定参数
	var req UpdateResumeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		util.GetLogger().Error("绑定请求参数失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 更新简历
	resume, err := h.resumeService.UpdateResume(
		uint(id),
		userID,
		req.Role,
		req.Level,
		req.University,
		req.PassCompany,
	)

	if err != nil {
		switch err {
		case service.ErrResumeNotFound:
			common.ResponseWithError(c, common.CodeDataNotFound)
		case service.ErrNotResumeOwner:
			common.ResponseWithError(c, common.CodeForbidden, http.StatusForbidden)
		default:
			util.GetLogger().Error("更新简历失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 转换为响应结构
	resp := h.convertToResumeResponse(resume)

	common.ResponseWithData(c, resp)
}

// UpdateResumeFile 更新简历文件
// @Summary 更新简历文件
// @Description 更新简历文件并转换为图片
// @Tags 简历
// @Accept multipart/form-data
// @Produce json
// @Param id path int true "简历ID"
// @Param file formData file true "新简历文件(PDF)"
// @Success 200 {object} common.Response{data=ResumeResponse}
// @Failure 400,401,403,404,500 {object} common.Response
// @Router /api/v1/resumes/{id}/file [put]
// @Security BearerAuth
func (h *ResumeHandler) UpdateResumeFile(c *gin.Context) {
	// 获取简历ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 获取当前用户ID
	userID := getCurrentUserID(c)
	if userID == 0 {
		common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
		return
	}

	// 获取文件
	file, err := c.FormFile("file")
	if err != nil {
		util.GetLogger().Error("获取上传文件失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 更新文件
	resume, err := h.resumeService.UpdateResumeFile(c, uint(id), userID, file)
	if err != nil {
		switch err {
		case service.ErrResumeNotFound:
			common.ResponseWithError(c, common.CodeDataNotFound)
		case service.ErrNotResumeOwner:
			common.ResponseWithError(c, common.CodeForbidden, http.StatusForbidden)
		case util.ErrFileTooLarge:
			common.ResponseWithError(c, common.CodeInvalidParams)
		case util.ErrInvalidFileType:
			common.ResponseWithError(c, common.CodeInvalidParams)
		default:
			util.GetLogger().Error("更新简历文件失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 转换为响应结构
	resp := h.convertToResumeResponse(resume)

	common.ResponseWithData(c, resp)
}

// DeleteResume 删除简历
// @Summary 删除简历
// @Description 删除简历及相关文件
// @Tags 简历
// @Produce json
// @Param id path int true "简历ID"
// @Success 200 {object} common.Response
// @Failure 400,401,403,404,500 {object} common.Response
// @Router /api/v1/resumes/{id} [delete]
// @Security BearerAuth
func (h *ResumeHandler) DeleteResume(c *gin.Context) {
	// 获取简历ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 获取当前用户ID
	userID := getCurrentUserID(c)
	if userID == 0 {
		common.ResponseWithError(c, common.CodeUnauthorized, http.StatusUnauthorized)
		return
	}

	// 删除简历
	err = h.resumeService.DeleteResume(uint(id), userID)
	if err != nil {
		switch err {
		case service.ErrResumeNotFound:
			common.ResponseWithError(c, common.CodeDataNotFound)
		case service.ErrNotResumeOwner:
			common.ResponseWithError(c, common.CodeForbidden, http.StatusForbidden)
		default:
			util.GetLogger().Error("删除简历失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	common.ResponseSuccess(c)
}

// DownloadResume 下载简历
// @Summary 下载简历
// @Description 下载简历文件
// @Tags 简历
// @Produce octet-stream
// @Param id path int true "简历ID"
// @Success 200 {file} file "简历文件"
// @Failure 400,401,403,404,500 {object} common.Response
// @Router /api/v1/resumes/{id}/download [get]
func (h *ResumeHandler) DownloadResume(c *gin.Context) {
	// 获取简历ID
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		common.ResponseWithError(c, common.CodeInvalidParams)
		return
	}

	// 获取当前用户ID
	userID := getCurrentUserID(c)

	// 获取简历
	resume, err := h.resumeService.DownloadResume(c, uint(id), userID)
	if err != nil {
		switch err {
		case service.ErrResumeNotFound:
			common.ResponseWithError(c, common.CodeDataNotFound)
		case service.ErrViewLimitExceeded:
			common.ResponseWithError(c, common.CodeOperationNotAllowed)
		default:
			util.GetLogger().Error("下载简历失败", zap.Error(err))
			common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		}
		return
	}

	// 发送文件
	c.FileAttachment(resume.ImageURL, "resume.pdf")
}

// ServeResumeFile 提供简历文件服务
// @Summary 提供简历文件服务
// @Description 静态文件服务，提供简历文件和图片访问
// @Tags 文件
// @Produce octet-stream
// @Param path path string true "文件路径"
// @Success 200 {file} file "文件内容"
// @Router /api/v1/files/{path} [get]
func (h *ResumeHandler) ServeResumeFile(c *gin.Context) {
	filePath := c.Param("path")
	if filePath == "" {
		c.Status(http.StatusNotFound)
		return
	}

	// 构建完整路径
	fullPath := filepath.Join(util.UploadDir, filePath)

	// 检查文件是否存在
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		util.GetLogger().Error("文件不存在", zap.String("path", fullPath))
		c.Status(http.StatusNotFound)
		return
	}

	// 设置正确的Content-Type
	extension := filepath.Ext(fullPath)
	switch strings.ToLower(extension) {
	case ".jpg", ".jpeg":
		c.Header("Content-Type", "image/jpeg")
	case ".png":
		c.Header("Content-Type", "image/png")
	case ".pdf":
		c.Header("Content-Type", "application/pdf")
	}

	// 提供文件
	c.File(fullPath)
}
