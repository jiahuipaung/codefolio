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

// UploadResumeRequest 上传简历请求
type UploadResumeRequest struct {
	Title       string               `form:"title" binding:"required"`
	Description string               `form:"description"`
	Tags        []string             `form:"tags[]"`
	Directions  []string             `form:"directions[]"`
	Offers      []service.OfferInput `form:"offers"`
}

// UpdateResumeRequest 更新简历请求
type UpdateResumeRequest struct {
	Title       string               `json:"title"`
	Description string               `json:"description"`
	Tags        []string             `json:"tags"`
	Directions  []string             `json:"directions"`
	Offers      []service.OfferInput `json:"offers"`
}

// ResumeResponse 简历响应
type ResumeResponse struct {
	ID            uint            `json:"id"`
	Title         string          `json:"title"`
	Description   string          `json:"description"`
	FileName      string          `json:"file_name"`
	FileSize      int64           `json:"file_size"`
	FileURL       string          `json:"file_url"`
	ViewCount     int             `json:"view_count"`
	DownloadCount int             `json:"download_count"`
	Tags          []TagResponse   `json:"tags"`
	Directions    []TagResponse   `json:"directions"`
	Offers        []OfferResponse `json:"offers"`
	CreatedAt     string          `json:"created_at"`
	UpdatedAt     string          `json:"updated_at"`
}

// TagResponse 标签响应
type TagResponse struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
}

// OfferResponse Offer响应
type OfferResponse struct {
	ID        uint   `json:"id"`
	Company   string `json:"company"`
	Position  string `json:"position"`
	OfferDate string `json:"offer_date"`
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
func (h *ResumeHandler) convertToResumeResponse(c *gin.Context, resume *service.Resume) ResumeResponse {
	// 分离技术栈标签和方向标签
	var tags []TagResponse
	var directions []TagResponse

	for _, tag := range resume.Tags {
		tagResp := TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		}

		if tag.Type == "tech_stack" {
			tags = append(tags, tagResp)
		} else if tag.Type == "direction" {
			directions = append(directions, tagResp)
		}
	}

	// 转换Offer
	var offers []OfferResponse
	for _, offer := range resume.Offers {
		offers = append(offers, OfferResponse{
			ID:        offer.ID,
			Company:   offer.Company,
			Position:  offer.Position,
			OfferDate: offer.OfferDate.Format("2006-01-02"),
		})
	}

	// 获取文件URL
	fileURL := h.resumeService.GetResumeFileURL(c, resume)

	return ResumeResponse{
		ID:            resume.ID,
		Title:         resume.Title,
		Description:   resume.Description,
		FileName:      resume.FileName,
		FileSize:      resume.FileSize,
		FileURL:       fileURL,
		ViewCount:     resume.ViewCount,
		DownloadCount: resume.DownloadCount,
		Tags:          tags,
		Directions:    directions,
		Offers:        offers,
		CreatedAt:     resume.CreatedAt.Format("2006-01-02 15:04:05"),
		UpdatedAt:     resume.UpdatedAt.Format("2006-01-02 15:04:05"),
	}
}

// UploadResume 上传简历
// @Summary 上传简历
// @Description 上传简历文件
// @Tags 简历
// @Accept multipart/form-data
// @Produce json
// @Param file formData file true "简历文件(PDF)"
// @Param title formData string true "简历标题"
// @Param description formData string false "简历描述"
// @Param tags[] formData []string false "技术栈标签"
// @Param directions[] formData []string false "技术方向标签"
// @Param offers formData []service.OfferInput false "公司Offer记录"
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
		req.Title,
		req.Description,
		file,
		req.Tags,
		req.Directions,
		req.Offers,
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
	resp := h.convertToResumeResponse(c, resume)

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
	resp := h.convertToResumeResponse(c, resume)

	common.ResponseWithData(c, resp)
}

// GetResumes 获取简历列表
// @Summary 获取简历列表
// @Description 获取所有简历列表，支持分页和筛选
// @Tags 简历
// @Produce json
// @Param page query int false "页码，默认1"
// @Param size query int false "每页数量，默认10"
// @Param tag_id query int false "按标签ID筛选"
// @Param direction query string false "按技术方向筛选"
// @Param company query string false "按公司名称筛选"
// @Success 200 {object} common.Response{data=[]ResumeResponse}
// @Failure 400,500 {object} common.Response
// @Router /api/v1/resumes [get]
func (h *ResumeHandler) GetResumes(c *gin.Context) {
	// 获取分页参数
	page, size := GetPagingParams(c)

	// 获取筛选参数
	tagIDStr := c.DefaultQuery("tag_id", "0")
	tagID, _ := strconv.ParseUint(tagIDStr, 10, 32)

	direction := c.DefaultQuery("direction", "")
	company := c.DefaultQuery("company", "")

	// 获取当前用户ID
	userID := getCurrentUserID(c)

	// 获取简历列表
	resumes, total, err := h.resumeService.GetAllResumes(page, size, uint(tagID), direction, company, userID)
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
		respList = append(respList, h.convertToResumeResponse(c, &resume))
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
		respList = append(respList, h.convertToResumeResponse(c, &resume))
	}

	common.ResponseWithData(c, respList)
}

// UpdateResume 更新简历信息
// @Summary 更新简历信息
// @Description 更新简历基本信息、标签和Offer
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
		req.Title,
		req.Description,
		req.Tags,
		req.Directions,
		req.Offers,
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
	resp := h.convertToResumeResponse(c, resume)

	common.ResponseWithData(c, resp)
}

// UpdateResumeFile 更新简历文件
// @Summary 更新简历文件
// @Description 更新简历文件
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
	resp := h.convertToResumeResponse(c, resume)

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
	c.FileAttachment(resume.FilePath, resume.FileName)
}

// GetTags 获取所有标签
// @Summary 获取所有标签
// @Description 获取所有技术栈或方向标签
// @Tags 标签
// @Produce json
// @Param type query string false "标签类型(tech_stack或direction)"
// @Success 200 {object} common.Response{data=[]TagResponse}
// @Failure 500 {object} common.Response
// @Router /api/v1/resume-tags [get]
func (h *ResumeHandler) GetTags(c *gin.Context) {
	tagType := c.DefaultQuery("type", "")

	// 获取标签
	tags, err := h.resumeService.GetAllTags(tagType)
	if err != nil {
		util.GetLogger().Error("获取标签失败", zap.Error(err))
		common.ResponseWithError(c, common.CodeInternalError, http.StatusInternalServerError)
		return
	}

	// 转换为响应结构
	var respList []TagResponse
	for _, tag := range tags {
		respList = append(respList, TagResponse{
			ID:   tag.ID,
			Name: tag.Name,
		})
	}

	common.ResponseWithData(c, respList)
}

// ServeResumeFile 提供简历文件服务
// @Summary 提供简历文件服务
// @Description 静态文件服务，提供简历文件访问
// @Tags 文件
// @Produce octet-stream
// @Param path path string true "文件路径"
// @Success 200 {file} file "简历文件"
// @Router /api/v1/files/{path} [get]
func (h *ResumeHandler) ServeResumeFile(c *gin.Context) {
	filePath := c.Param("path")
	if filePath == "" {
		c.Status(http.StatusNotFound)
		return
	}

	// 构建完整路径
	fullPath := util.UploadDir + "/" + filePath

	// 提供文件
	c.File(fullPath)
}
