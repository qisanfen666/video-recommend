package domain

import (
	"time"
)

type Behavior struct {
	ID        int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID    int64     `gorm:"index;not null" json:"user_id"`
	VideoID   int64     `gorm:"index;not null" json:"video_id"`
	Action    string    `gorm:"size:20;not null" json:"action"`
	WatchTime int       `gorm:"default:0" json:"watch_time"`
	Score     float64   `gorm:"default:0" json:"score"`
	IP        string    `gorm:"size:50" json:"ip"`
	Device    string    `gorm:"size:50" json:"device"`
	TimeStamp time.Time `gorm:"index" json:"timestamp"`
	CreatedAt time.Time `json:"created_at"`
}

func (Behavior) TableName() string {
	return "behaviors"
}

const (
	ActionClick  = "click"
	ActionWatch  = "watch"
	ActionKanwan = "kanwan"
	ActionLike   = "like"
	ActionShare  = "share"
	ActionFavor  = "favor"
)

type BehaviorCreateRequest struct {
	VideoID   int64  `json:"video_id" binding:"omitempty"`
	Action    string `json:"action" binding:"required,oneof=click watch kanwan like share favor"`
	WatchTime int    `json:"watch_time" binding:"min=0"`
}

type UserHistoryRequest struct {
	Page     int    `form:"page" binding:"min=1"`
	PageSize int    `form:"page_size" binding:"min=1,max=100"`
	Action   string `form:"action" binding:"omitempty,oneof=click watch kanwan like share favor"`
}

type UserHistoryResponse struct {
	VideoID   int64     `json:"video_id"`
	Title     string    `json:"title"`
	CoverURL  string    `json:"cover_url"`
	Action    string    `json:"action"`
	WatchTime int       `json:"watch_time"`
	Timestamp time.Time `json:"timestamp"`
}
