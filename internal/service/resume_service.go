package service

import (
	"codefolio/internal/domain"
	"codefolio/internal/repository"
	"codefolio/internal/util"
	"errors"
	"mime/multipart"
	"time"

	"github.com/gin-gonic/gin"
)

// 简历相关错误
var (
	ErrResumeNotFound    = errors.New("简历不存在")
	ErrNotResumeOwner    = errors.New("非简历所有者，无权操作")
	ErrViewLimitExceeded = errors.New("已超过简历查看限制")
)

// ResumeService 简历服务接口
type ResumeService interface {
	// 简历基本操作
	CreateResume(c *gin.Context, userID uint, title, description string, file *multipart.FileHeader, tags []string, directions []string, offers []OfferInput) (*domain.Resume, error)
	GetResumeByID(c *gin.Context, id uint, currentUserID uint) (*domain.Resume, error)
	GetUserResumes(userID uint) ([]domain.Resume, error)
	GetAllResumes(page, size int, tagID uint, direction, company string, currentUserID uint) ([]domain.Resume, int64, error)
	UpdateResume(resumeID, userID uint, title, description string, tags []string, directions []string, offers []OfferInput) (*domain.Resume, error)
	UpdateResumeFile(c *gin.Context, resumeID, userID uint, file *multipart.FileHeader) (*domain.Resume, error)
	DeleteResume(resumeID, userID uint) error

	// 标签相关
	GetAllTags(tagType string) ([]domain.Tag, error)

	// 文件相关
	GetResumeFileURL(c *gin.Context, resume *domain.Resume) string
	DownloadResume(c *gin.Context, resumeID, userID uint) (*domain.Resume, error)

	// 访问控制
	CanViewResume(userID uint) bool
}

// OfferInput Offer输入数据
type OfferInput struct {
	Company  string `json:"company"`
	Position string `json:"position"`
	Date     string `json:"date"`
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
	return &resumeService{
		resumeRepo:          resumeRepo,
		userRepo:            userRepo,
		anonymousViewLimit:  anonymousViewLimit,
		registeredViewLimit: registeredViewLimit,
	}
}

// CreateResume 创建简历
func (s *resumeService) CreateResume(c *gin.Context, userID uint, title, description string, file *multipart.FileHeader, tagNames []string, directionNames []string, offers []OfferInput) (*domain.Resume, error) {
	// 保存文件
	fileResult, err := util.SaveUploadedFile(c, file, userID)
	if err != nil {
		return nil, err
	}

	// 创建简历记录
	resume := &domain.Resume{
		UserID:      userID,
		Title:       title,
		Description: description,
		FilePath:    fileResult.FilePath,
		FileName:    fileResult.FileName,
		FileType:    fileResult.FileType,
		FileSize:    fileResult.FileSize,
	}

	// 使用事务确保数据一致性
	err = s.resumeRepo.Create(resume)
	if err != nil {
		// 如果保存数据库失败，删除已上传的文件
		_ = util.DeleteFile(fileResult.FilePath)
		return nil, err
	}

	// 处理标签
	for _, tagName := range tagNames {
		tag, err := s.resumeRepo.FindOrCreateTag(tagName, "tech_stack")
		if err != nil {
			continue
		}
		resume.Tags = append(resume.Tags, *tag)
	}

	// 处理方向
	for _, dirName := range directionNames {
		dir, err := s.resumeRepo.FindOrCreateTag(dirName, "direction")
		if err != nil {
			continue
		}
		resume.Tags = append(resume.Tags, *dir)
	}

	// 处理Offer
	for _, offerInput := range offers {
		offerDate, _ := time.Parse("2006-01-02", offerInput.Date)
		offer := &domain.Offer{
			ResumeID:  resume.ID,
			Company:   offerInput.Company,
			Position:  offerInput.Position,
			OfferDate: offerDate,
		}
		_ = s.resumeRepo.CreateOffer(offer)
		resume.Offers = append(resume.Offers, *offer)
	}

	return resume, nil
}

// GetResumeByID 根据ID获取简历
func (s *resumeService) GetResumeByID(c *gin.Context, id uint, currentUserID uint) (*domain.Resume, error) {
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
func (s *resumeService) GetAllResumes(page, size int, tagID uint, direction, company string, currentUserID uint) ([]domain.Resume, int64, error) {
	// 检查访问权限
	if !s.CanViewResume(currentUserID) {
		return nil, 0, ErrViewLimitExceeded
	}

	return s.resumeRepo.FindAll(page, size, tagID, direction, company)
}

// UpdateResume 更新简历信息（不包括文件）
func (s *resumeService) UpdateResume(resumeID, userID uint, title, description string, tagNames []string, directionNames []string, offers []OfferInput) (*domain.Resume, error) {
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
	resume.Title = title
	resume.Description = description

	// 保存基本信息
	if err := s.resumeRepo.Update(resume); err != nil {
		return nil, err
	}

	// 更新标签（先删除所有关联，再重新添加）
	resume.Tags = []domain.Tag{}

	// 处理技术栈标签
	for _, tagName := range tagNames {
		tag, err := s.resumeRepo.FindOrCreateTag(tagName, "tech_stack")
		if err != nil {
			continue
		}
		resume.Tags = append(resume.Tags, *tag)
	}

	// 处理方向标签
	for _, dirName := range directionNames {
		dir, err := s.resumeRepo.FindOrCreateTag(dirName, "direction")
		if err != nil {
			continue
		}
		resume.Tags = append(resume.Tags, *dir)
	}

	// 更新Offer（先删除所有关联，再重新添加）
	if err := s.resumeRepo.DeleteOffersByResumeID(resumeID); err != nil {
		return nil, err
	}

	resume.Offers = []domain.Offer{}

	// 处理Offer
	for _, offerInput := range offers {
		offerDate, _ := time.Parse("2006-01-02", offerInput.Date)
		offer := &domain.Offer{
			ResumeID:  resume.ID,
			Company:   offerInput.Company,
			Position:  offerInput.Position,
			OfferDate: offerDate,
		}
		if err := s.resumeRepo.CreateOffer(offer); err != nil {
			continue
		}
		resume.Offers = append(resume.Offers, *offer)
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
	oldFilePath := resume.FilePath

	// 更新简历信息
	resume.FilePath = fileResult.FilePath
	resume.FileName = fileResult.FileName
	resume.FileType = fileResult.FileType
	resume.FileSize = fileResult.FileSize

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
	_ = util.DeleteFile(resume.FilePath)

	// 删除数据库记录
	return s.resumeRepo.Delete(resumeID)
}

// GetAllTags 获取所有标签
func (s *resumeService) GetAllTags(tagType string) ([]domain.Tag, error) {
	return s.resumeRepo.FindAllTags(tagType)
}

// GetResumeFileURL 获取简历文件URL
func (s *resumeService) GetResumeFileURL(c *gin.Context, resume *domain.Resume) string {
	return util.GetFileURL(c, resume.FilePath)
}

// DownloadResume 下载简历
func (s *resumeService) DownloadResume(c *gin.Context, resumeID, userID uint) (*domain.Resume, error) {
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
