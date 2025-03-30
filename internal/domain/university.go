package domain

// University 大学模型
type University struct {
	ID   uint   `json:"id" gorm:"primaryKey"`
	Name string `json:"name" gorm:"size:100;not null;unique"`
}

// UniversityRepository 大学仓库接口
type UniversityRepository interface {
	GetAll() ([]University, error)
}
