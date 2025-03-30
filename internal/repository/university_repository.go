package repository

import (
	"codefolio/internal/domain"
	"gorm.io/gorm"
)

// universityRepository 大学仓库实现
type universityRepository struct {
	db *gorm.DB
}

// NewUniversityRepository 创建大学仓库实例
func NewUniversityRepository(db *gorm.DB) domain.UniversityRepository {
	return &universityRepository{db: db}
}

// GetAll 获取所有大学
func (r *universityRepository) GetAll() ([]domain.University, error) {
	var universities []domain.University
	if err := r.db.Order("name").Find(&universities).Error; err != nil {
		return nil, err
	}
	return universities, nil
}
