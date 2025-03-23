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
	FindAll(page, size int, tagID uint, direction, company string) ([]domain.Resume, int64, error)
	Update(resume *domain.Resume) error
	Delete(id uint) error
	IncrementViewCount(id uint) error
	IncrementDownloadCount(id uint) error

	// 标签相关
	FindAllTags(tagType string) ([]domain.Tag, error)
	CreateTag(tag *domain.Tag) error
	FindOrCreateTag(name, tagType string) (*domain.Tag, error)

	// Offer相关
	CreateOffer(offer *domain.Offer) error
	DeleteOffersByResumeID(resumeID uint) error
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
	if err := r.db.Preload("Tags").Preload("Offers").Where("id = ?", id).First(&resume).Error; err != nil {
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
	if err := r.db.Preload("Tags").Preload("Offers").Where("user_id = ?", userID).Find(&resumes).Error; err != nil {
		return nil, err
	}
	return resumes, nil
}

// FindAll 查询所有简历，支持分页和筛选
func (r *resumeRepository) FindAll(page, size int, tagID uint, direction, company string) ([]domain.Resume, int64, error) {
	var resumes []domain.Resume
	var total int64

	offset := (page - 1) * size

	// 构建查询
	query := r.db.Model(&domain.Resume{})

	// 标签筛选
	if tagID > 0 {
		query = query.Joins("JOIN resume_tags ON resumes.id = resume_tags.resume_id").
			Where("resume_tags.tag_id = ?", tagID)
	}

	// 方向筛选
	if direction != "" {
		query = query.Joins("JOIN resume_tags rt ON resumes.id = rt.resume_id").
			Joins("JOIN tags t ON rt.tag_id = t.id").
			Where("t.type = 'direction' AND t.name LIKE ?", "%"+direction+"%")
	}

	// 公司筛选
	if company != "" {
		query = query.Joins("JOIN offers o ON resumes.id = o.resume_id").
			Where("o.company LIKE ?", "%"+company+"%")
	}

	// 计算总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 查询数据
	if err := query.Preload("Tags").Preload("Offers").
		Offset(offset).Limit(size).
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
	// 使用事务确保原子性
	return r.db.Transaction(func(tx *gorm.DB) error {
		// 删除关联的offer记录
		if err := tx.Where("resume_id = ?", id).Delete(&domain.Offer{}).Error; err != nil {
			return err
		}

		// 删除标签关联
		if err := tx.Where("resume_id = ?", id).Delete(&domain.ResumeTag{}).Error; err != nil {
			return err
		}

		// 删除简历记录
		if err := tx.Delete(&domain.Resume{}, id).Error; err != nil {
			return err
		}

		return nil
	})
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

// FindAllTags 查询所有标签
func (r *resumeRepository) FindAllTags(tagType string) ([]domain.Tag, error) {
	var tags []domain.Tag
	query := r.db.Model(&domain.Tag{})

	if tagType != "" {
		query = query.Where("type = ?", tagType)
	}

	if err := query.Find(&tags).Error; err != nil {
		return nil, err
	}

	return tags, nil
}

// CreateTag 创建标签
func (r *resumeRepository) CreateTag(tag *domain.Tag) error {
	return r.db.Create(tag).Error
}

// FindOrCreateTag 查找或创建标签
func (r *resumeRepository) FindOrCreateTag(name, tagType string) (*domain.Tag, error) {
	var tag domain.Tag

	err := r.db.Where("name = ? AND type = ?", name, tagType).
		FirstOrCreate(&tag, domain.Tag{
			Name: name,
			Type: tagType,
		}).Error

	if err != nil {
		return nil, err
	}

	return &tag, nil
}

// CreateOffer 创建Offer记录
func (r *resumeRepository) CreateOffer(offer *domain.Offer) error {
	return r.db.Create(offer).Error
}

// DeleteOffersByResumeID 删除简历相关的所有Offer记录
func (r *resumeRepository) DeleteOffersByResumeID(resumeID uint) error {
	return r.db.Where("resume_id = ?", resumeID).Delete(&domain.Offer{}).Error
}
