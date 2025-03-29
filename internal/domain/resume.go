package domain

import (
	"time"
)

// Resume 简历信息
type Resume struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	ImageURL    string    `json:"image_url"`    // 简历图片URL, 由前端上传的PDF文件转换为图片后存储在服务器上的URL
	Role        string    `json:"role"`         // 应聘职位
	Level       string    `json:"level"`        // 经历等级：实习生/应届生/社招
	University  string    `json:"university"`   // 毕业院校
	PassCompany []string  `json:"pass_company"` // 面试通过的公司
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
