package domain

import (
	"time"
)

// Role 定义岗位类型
type Role int

//const (
//	Frontend  Role = iota + 1 // 1
//	Backend                   // 2
//	Algorithm                 // 3
//	Product                   // 4
//	Operation                 // 5
//)

// Level 定义经历等级
type Level int

//const (
//	Intern            Level = iota + 1 // 1
//	CampusRecruitment                  // 2
//	SocialRecruitment                  // 3
//)

// Company 定义公司类型
type Company int

//const (
//	Tencent     Company = iota + 1 // 1
//	Alibaba                        // 2
//	Meituan                        // 3
//	ByteDance                      // 4
//	JD                             // 5
//	Baidu                          // 6
//	Kuaishou                       // 7
//	Netease                        // 8
//	Pinduoduo                      // 9
//	Didi                           // 10
//	Huawei                         // 11
//	Bilibili                       // 12
//	Xiaohongshu                    // 13
//)

// Resume 简历信息
type Resume struct {
	ID          uint      `json:"id" gorm:"primaryKey"`
	UserID      uint      `json:"user_id"`
	ImageURL    string    `json:"image_url"`                       // 简历图片URL, 由前端上传的PDF文件转换为图片后存储在服务器上的URL
	Role        Role      `json:"role"`                            // 应聘职位
	Level       Level     `json:"level"`                           // 经历等级：实习生/应届生/社招
	University  string    `json:"university"`                      // 毕业院校
	PassCompany []Company `json:"pass_company" gorm:"type:text[]"` // 面试通过的公司
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}
