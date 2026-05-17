package domain

import (
	"time"
)

type User struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Username    string     `gorm:"uniqueIndex;size:50;not null" json:"username"`
	Password    string     `gorm:"size:255;not null" json:"-"`
	Nickname    string     `gorm:"size:100" json:"nickname"`
	Age         int        `gorm:"default:0" json:"age"`
	Gender      string     `gorm:"size:10;default:'unknown'" json:"gender"`
	Email       string     `gorm:"size:100" json:"email"`
	Avatar      string     `gorm:"size:255" json:"avatar"`
	Role        string     `gorm:"size:20;default:'user'" json:"role"`
	Preferences string     `gorm:"type:text" json:"preferences"`
	LastLoginAt *time.Time `json:"last_login_at,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (User) TableName() string {
	return "users"
}

type UserRegisterRequest struct {
	Username string `json:"username" binding:"required,min=3,max=50"`
	Password string `json:"password" binding:"required,min=6,max=100"`
	Email    string `json:"email" binding:"omitempty,email"`
}

type UserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

type UserLoginResponse struct {
	Token     string `json:"token"`
	ExpiresIn int64  `json:"expires_in"`
	User      *User  `json:"user"`
}

type UserProfileRequest struct {
	Nickname    string `json:"nickname" binding:"omitempty,max=100"`
	Age         int    `json:"age" binding:"omitempty,min=0,max=150"`
	Gender      string `json:"gender" binding:"omitempty,oneof=male female unknown"`
	Email       string `json:"email" binding:"omitempty,email"`
	Avatar      string `json:"avatar"`
	Preferences string `json:"preferences"`
}

type UserSim struct {
	UserID     int64   `json:"user_id"`
	Similarity float64 `json:"similarity"`
}
