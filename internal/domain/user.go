package domain

import (
	"time"

	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Email         string    `gorm:"uniqueIndex;not null" json:"email"`
	Password      string    `gorm:"not null" json:"-"`
	FirstName     string    `json:"first_name"`
	LastName      string    `json:"last_name"`
	LastLoginAt   time.Time `json:"last_login_at"`
	IsActive      bool      `gorm:"default:true" json:"is_active"`
	EmailVerified bool      `gorm:"default:false" json:"email_verified"`
}

type UserRepository interface {
	Create(user *User) error
	FindByEmail(email string) (*User, error)
	FindByID(id uint) (*User, error)
	Update(user *User) error
}

type UserService interface {
	Register(email, password, firstName, lastName string) (*User, error)
	Login(email, password string) (string, error)
	GetUserByID(id uint) (*User, error)
	UpdateUser(user *User) error
}
