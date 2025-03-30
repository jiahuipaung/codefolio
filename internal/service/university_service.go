package service

import "codefolio/internal/domain"

// UniversityService 大学服务接口
type UniversityService interface {
	GetAllUniversities() ([]domain.University, error)
}

// universityService 大学服务实现
type universityService struct {
	universityRepo domain.UniversityRepository
}

// NewUniversityService 创建大学服务实例
func NewUniversityService(universityRepo domain.UniversityRepository) UniversityService {
	return &universityService{
		universityRepo: universityRepo,
	}
}

// GetAllUniversities 获取所有大学
func (s *universityService) GetAllUniversities() ([]domain.University, error) {
	return s.universityRepo.GetAll()
}
