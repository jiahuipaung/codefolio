package repository

import (
	"codefolio/internal/domain"
	"errors"
	"gorm.io/gorm"
)

// ResumeRepository 简历仓库接口
type ResumeRepository interface {
	Create(resume *domain.Resume) error
	FindByID(id uint) (*domain.Resume, error)
	FindByUser(userID uint) ([]domain.Resume, error)
	FindAll(page, size int, role, level, university int) ([]domain.Resume, int64, error)
	Update(resume *domain.Resume) error
	Delete(id uint) error
	IncrementViewCount(id uint) error
	IncrementDownloadCount(id uint) error
}

// resumeRepository 简历仓库实现
type resumeRepository struct {
	db *gorm.DB
}

// NewResumeRepository 创建简历仓库实例
func NewResumeRepository(db *gorm.DB) ResumeRepository {
	return &resumeRepository{db: db}
}

// Create 创建简历
func (r *resumeRepository) Create(resume *domain.Resume) error {
	return r.db.Create(resume).Error
}

// FindByID 根据ID查找简历
func (r *resumeRepository) FindByID(id uint) (*domain.Resume, error) {
	var resume domain.Resume
	if err := r.db.Where("id = ?", id).First(&resume).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &resume, nil
}

// FindByUser 查找用户的所有简历
func (r *resumeRepository) FindByUser(userID uint) ([]domain.Resume, error) {
	var resumes []domain.Resume
	if err := r.db.Where("user_id = ?", userID).Find(&resumes).Error; err != nil {
		return nil, err
	}
	return resumes, nil
}

// FindAll 查询所有简历，支持分页和筛选
func (r *resumeRepository) FindAll(page, size int, role, level, university int) ([]domain.Resume, int64, error) {
	var resumes []domain.Resume
	var total int64

	offset := (page - 1) * size

	// 构建查询
	query := r.db.Model(&domain.Resume{})

	// 职位筛选
	if role > 0 {
		query = query.Where("role = ?", role)
	}

	// 经历等级筛选
	if level > 0 {
		query = query.Where("level = ?", level)
	}

	// 毕业院校筛选
	if university > 0 {
		query = query.Where("university = ?", university)
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if err := query.Offset(offset).Limit(size).
		Order("created_at DESC").
		Find(&resumes).Error; err != nil {
		return nil, 0, err
	}

	return resumes, total, nil
}

// Update 更新简历信息
func (r *resumeRepository) Update(resume *domain.Resume) error {
	return r.db.Save(resume).Error
}

// Delete 删除简历
func (r *resumeRepository) Delete(id uint) error {
	return r.db.Delete(&domain.Resume{}, id).Error
}

// IncrementViewCount 增加查看次数
func (r *resumeRepository) IncrementViewCount(id uint) error {
	return r.db.Model(&domain.Resume{}).
		Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).
		Error
}

// IncrementDownloadCount 增加下载次数
func (r *resumeRepository) IncrementDownloadCount(id uint) error {
	return r.db.Model(&domain.Resume{}).
		Where("id = ?", id).
		UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).
		Error
}
