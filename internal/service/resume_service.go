package service

import (
	"codefolio/internal/domain"
	"codefolio/internal/repository"
	"codefolio/internal/util"
	"errors"
	"fmt"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// 简历相关错误
var (
	ErrResumeNotFound    = errors.New("简历不存在")
	ErrNotResumeOwner    = errors.New("非简历所有者，无权操作")
	ErrViewLimitExceeded = errors.New("已超过简历查看限制")
	ErrFileNotFound      = errors.New("文件不存在或已过期")
)

// FileResult 文件处理结果
type FileResult struct {
	FilePath string // 文件路径
	FileKey  string // 文件唯一标识
}

// TempFileInfo 临时文件信息
type TempFileInfo struct {
	UserID    uint      // 上传用户ID
	FilePath  string    // 文件路径
	CreatedAt time.Time // 创建时间
}

// 临时文件缓存，用于存储已上传但尚未关联到简历的文件
var tempFiles = make(map[string]TempFileInfo)

// ResumeService 简历服务接口
type ResumeService interface {
	// 简历基本操作
	CreateResume(c *gin.Context, userID uint, file *multipart.FileHeader, role, level, university string, passCompany []string) (*domain.Resume, error)
	GetResumeByID(c *gin.Context, id uint, currentUserID uint) (*domain.Resume, error)
	GetUserResumes(userID uint) ([]domain.Resume, error)
	GetAllResumes(page, size int, role, level, university string, currentUserID uint) ([]domain.Resume, int64, error)
	UpdateResume(resumeID, userID uint, role, level, university string, passCompany []string) (*domain.Resume, error)
	UpdateResumeFile(c *gin.Context, resumeID, userID uint, file *multipart.FileHeader) (*domain.Resume, error)
	DeleteResume(resumeID, userID uint) error

	// 文件相关
	UploadAndConvertPDF(c *gin.Context, userID uint, file *multipart.FileHeader) (*FileResult, error)
	CreateResumeWithFileKey(userID uint, fileKey, role, level, university string, passCompany []string) (*domain.Resume, error)
	GetResumeFileURL(c *gin.Context, resume *domain.Resume) string
	DownloadResume(c *gin.Context, resumeID, userID uint) (*domain.Resume, error)

	// 访问控制
	CanViewResume(userID uint) bool
}

// resumeService 简历服务实现
type resumeService struct {
	resumeRepo repository.ResumeRepository
	userRepo   repository.UserRepository

	// 未登录用户可浏览的简历数量
	anonymousViewLimit int
	// 未上传简历的注册用户可浏览的简历数量
	registeredViewLimit int
}

// NewResumeService 创建简历服务实例
func NewResumeService(resumeRepo repository.ResumeRepository, userRepo repository.UserRepository, anonymousViewLimit, registeredViewLimit int) ResumeService {
	// 启动临时文件清理goroutine
	go cleanupTempFiles()

	return &resumeService{
		resumeRepo:          resumeRepo,
		userRepo:            userRepo,
		anonymousViewLimit:  anonymousViewLimit,
		registeredViewLimit: registeredViewLimit,
	}
}

// 定期清理过期的临时文件（超过30分钟未使用的文件）
func cleanupTempFiles() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		now := time.Now()
		for key, info := range tempFiles {
			// 超过30分钟的文件视为过期
			if now.Sub(info.CreatedAt) > 30*time.Minute {
				_ = util.DeleteFile(info.FilePath)
				delete(tempFiles, key)
			}
		}
	}
}

// UploadAndConvertPDF 上传并转换PDF文件为图片（第一步）
func (s *resumeService) UploadAndConvertPDF(c *gin.Context, userID uint, file *multipart.FileHeader) (*FileResult, error) {
	// 上传并转换PDF为图片
	uploadResult, err := util.SaveUploadedPDF(c, file, userID)
	if err != nil {
		return nil, err
	}

	// 生成文件唯一标识
	fileKey := uuid.New().String()

	// 记录物理文件路径（用于调试）
	filePath := uploadResult.FilePath

	// 构建直接访问的URL路径
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}

	host := c.Request.Host
	if host == "" {
		host = "localhost:8080"
	}

	// 使用新增的直接静态文件访问路径
	imageURL := fmt.Sprintf("%s://%s%s", scheme, host, filePath)

	util.GetLogger().Info("PDF转图片成功",
		zap.String("filePath", filePath),
		zap.String("imageURL", imageURL))

	// 存入临时文件缓存
	tempFiles[fileKey] = TempFileInfo{
		UserID:    userID,
		FilePath:  uploadResult.FilePath,
		CreatedAt: time.Now(),
	}

	// 返回结果包含图片URL和文件标识
	return &FileResult{
		FilePath: imageURL, // 返回完整的URL，而不是文件路径
		FileKey:  fileKey,
	}, nil
}

// CreateResumeWithFileKey 使用文件标识创建简历（第二步）
func (s *resumeService) CreateResumeWithFileKey(userID uint, fileKey, role, level, university string, passCompany []string) (*domain.Resume, error) {
	// 获取临时文件信息
	fileInfo, exists := tempFiles[fileKey]
	if !exists {
		return nil, ErrFileNotFound
	}

	// 验证用户身份
	if fileInfo.UserID != userID {
		return nil, ErrNotResumeOwner
	}

	// 创建简历记录
	resume := &domain.Resume{
		UserID:      userID,
		ImageURL:    fileInfo.FilePath,
		Role:        role,
		Level:       level,
		University:  university,
		PassCompany: passCompany,
	}

	// 保存到数据库
	err := s.resumeRepo.Create(resume)
	if err != nil {
		return nil, err
	}

	// 从临时文件缓存中移除
	delete(tempFiles, fileKey)

	return resume, nil
}

// CreateResume 创建简历（一次性操作，保留兼容性）
func (s *resumeService) CreateResume(c *gin.Context, userID uint, file *multipart.FileHeader, role, level, university string, passCompany []string) (*domain.Resume, error) {
	// 保存文件并转换为图片
	fileResult, err := util.SaveUploadedFile(c, file, userID)
	if err != nil {
		return nil, err
	}

	// 创建简历记录
	resume := &domain.Resume{
		UserID:      userID,
		ImageURL:    fileResult.FilePath,
		Role:        role,
		Level:       level,
		University:  university,
		PassCompany: passCompany,
	}

	// 使用事务确保数据一致性
	err = s.resumeRepo.Create(resume)
	if err != nil {
		// 如果保存数据库失败，删除已上传的文件
		_ = util.DeleteFile(fileResult.FilePath)
		return nil, err
	}

	return resume, nil
}

// GetResumeByID 根据ID获取简历
func (s *resumeService) GetResumeByID(_ *gin.Context, id uint, currentUserID uint) (*domain.Resume, error) {
	// 检查是否可以访问简历
	if !s.CanViewResume(currentUserID) {
		return nil, ErrViewLimitExceeded
	}

	// 获取简历
	resume, err := s.resumeRepo.FindByID(id)
	if err != nil || resume == nil {
		return nil, ErrResumeNotFound
	}

	// 如果不是简历所有者，增加查看次数
	if currentUserID != resume.UserID {
		_ = s.resumeRepo.IncrementViewCount(id)
	}

	return resume, nil
}

// GetUserResumes 获取用户的所有简历
func (s *resumeService) GetUserResumes(userID uint) ([]domain.Resume, error) {
	return s.resumeRepo.FindByUser(userID)
}

// GetAllResumes 获取所有简历（分页）
func (s *resumeService) GetAllResumes(page, size int, role, level, university string, currentUserID uint) ([]domain.Resume, int64, error) {
	// 检查访问权限
	if !s.CanViewResume(currentUserID) {
		return nil, 0, ErrViewLimitExceeded
	}

	return s.resumeRepo.FindAll(page, size, role, level, university)
}

// UpdateResume 更新简历信息（不包括文件）
func (s *resumeService) UpdateResume(resumeID, userID uint, role, level, university string, passCompany []string) (*domain.Resume, error) {
	// 获取简历
	resume, err := s.resumeRepo.FindByID(resumeID)
	if err != nil || resume == nil {
		return nil, ErrResumeNotFound
	}

	// 检查是否是简历所有者
	if resume.UserID != userID {
		return nil, ErrNotResumeOwner
	}

	// 更新基本信息
	resume.Role = role
	resume.Level = level
	resume.University = university
	resume.PassCompany = passCompany

	// 保存基本信息
	if err := s.resumeRepo.Update(resume); err != nil {
		return nil, err
	}

	return resume, nil
}

// UpdateResumeFile 更新简历文件
func (s *resumeService) UpdateResumeFile(c *gin.Context, resumeID, userID uint, file *multipart.FileHeader) (*domain.Resume, error) {
	// 获取简历
	resume, err := s.resumeRepo.FindByID(resumeID)
	if err != nil || resume == nil {
		return nil, ErrResumeNotFound
	}

	// 检查是否是简历所有者
	if resume.UserID != userID {
		return nil, ErrNotResumeOwner
	}

	// 保存新文件
	fileResult, err := util.SaveUploadedFile(c, file, userID)
	if err != nil {
		return nil, err
	}

	// 保存旧文件路径，以便失败时恢复
	oldFilePath := resume.ImageURL

	// 更新简历信息
	resume.ImageURL = fileResult.FilePath

	// 保存到数据库
	if err := s.resumeRepo.Update(resume); err != nil {
		// 如果更新失败，删除新上传的文件
		_ = util.DeleteFile(fileResult.FilePath)
		return nil, err
	}

	// 更新成功后，删除旧文件
	_ = util.DeleteFile(oldFilePath)

	return resume, nil
}

// DeleteResume 删除简历
func (s *resumeService) DeleteResume(resumeID, userID uint) error {
	// 获取简历
	resume, err := s.resumeRepo.FindByID(resumeID)
	if err != nil || resume == nil {
		return ErrResumeNotFound
	}

	// 检查是否是简历所有者
	if resume.UserID != userID {
		return ErrNotResumeOwner
	}

	// 删除文件
	_ = util.DeleteFile(resume.ImageURL)

	// 删除数据库记录
	return s.resumeRepo.Delete(resumeID)
}

// GetResumeFileURL 获取简历文件URL
func (s *resumeService) GetResumeFileURL(c *gin.Context, resume *domain.Resume) string {
	return util.GetFileURL(c, resume.ImageURL)
}

// DownloadResume 下载简历
func (s *resumeService) DownloadResume(_ *gin.Context, resumeID, userID uint) (*domain.Resume, error) {
	// 检查是否可以访问简历
	if !s.CanViewResume(userID) {
		return nil, ErrViewLimitExceeded
	}

	// 获取简历
	resume, err := s.resumeRepo.FindByID(resumeID)
	if err != nil || resume == nil {
		return nil, ErrResumeNotFound
	}

	// 如果不是简历所有者，增加下载次数
	if userID != resume.UserID {
		_ = s.resumeRepo.IncrementDownloadCount(resumeID)
	}

	return resume, nil
}

// CanViewResume 检查用户是否可以查看简历
func (s *resumeService) CanViewResume(userID uint) bool {
	// 未登录用户
	if userID == 0 {
		// 可以查看有限数量的简历
		// 这里需要从session中获取已查看数量，但简化实现
		return true // 开发阶段允许访问
	}

	// 已登录用户，检查是否上传过简历
	userResumes, err := s.resumeRepo.FindByUser(userID)
	if err != nil || len(userResumes) == 0 {
		// 没有上传过简历，限制访问数量
		// 这里需要从数据库中获取已查看数量，但简化实现
		return true // 开发阶段允许访问
	}

	// 已上传简历的用户可以无限制查看
	return true
}
