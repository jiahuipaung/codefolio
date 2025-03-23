package domain

import (
	"time"
)

// Resume 简历模型
type Resume struct {
	ID            uint      `json:"id" gorm:"primaryKey"`
	UserID        uint      `json:"user_id" gorm:"not null"`
	Title         string    `json:"title" gorm:"size:100;not null"`
	Description   string    `json:"description" gorm:"type:text"`
	FilePath      string    `json:"file_path" gorm:"size:255;not null"`
	FileName      string    `json:"file_name" gorm:"size:255;not null"`
	FileType      string    `json:"file_type" gorm:"size:50;not null;default:'application/pdf'"`
	FileSize      int64     `json:"file_size" gorm:"not null"`
	ViewCount     int       `json:"view_count" gorm:"default:0"`
	DownloadCount int       `json:"download_count" gorm:"default:0"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`

	// 关联
	Tags   []Tag   `json:"tags" gorm:"many2many:resume_tags;"`
	Offers []Offer `json:"offers" gorm:"foreignKey:ResumeID"`
	User   User    `json:"-" gorm:"foreignKey:UserID"`
}

// Tag 标签模型
type Tag struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"size:50;not null;uniqueIndex"`
	Type      string    `json:"type" gorm:"size:20;not null"` // 'tech_stack' 或 'direction'
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Resumes []Resume `json:"-" gorm:"many2many:resume_tags;"`
}

// Offer 公司Offer记录模型
type Offer struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	ResumeID  uint      `json:"resume_id" gorm:"not null"`
	Company   string    `json:"company" gorm:"size:100;not null"`
	Position  string    `json:"position" gorm:"size:100"`
	OfferDate time.Time `json:"offer_date"`
	CreatedAt time.Time `json:"created_at"`

	// 关联
	Resume Resume `json:"-" gorm:"foreignKey:ResumeID"`
}

// ResumeTag 简历与标签的关联模型
type ResumeTag struct {
	ResumeID uint `gorm:"primaryKey"`
	TagID    uint `gorm:"primaryKey"`
}
