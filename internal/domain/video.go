package domain

import (
	"time"
)

type Video struct {
	ID          int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	Title       string     `gorm:"size:200;not null;index" json:"title"`
	CoverURL    string     `gorm:"size:255" json:"cover_url"`
	VideoURL    string     `gorm:"size:255" json:"video_url"`
	Category    string     `gorm:"size:50;index" json:"category"`
	Tags        string     `gorm:"type:text" json:"tags"`
	Duration    int        `gorm:"default:0" json:"duration"`
	Description string     `gorm:"type:text" json:"description"`
	Heat        float64    `gorm:"default:0;index" json:"heat"`
	ViewCount   int64      `gorm:"default:0" json:"view_count"`
	LikeCount   int64      `gorm:"default:0" json:"like_count"`
	Status      int        `gorm:"default:1" json:"status"`
	UploadTime  *time.Time `json:"upload_time,omitempty"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func (Video) TableName() string {
	return "videos"
}

type VideoCreateRequest struct {
	Title       string   `json:"title" binding:"required,min=1,max=200"`
	CoverURL    string   `json:"cover_url"`
	VideoURL    string   `json:"video_url" binding:"required"`
	Category    string   `json:"category" binding:"required"`
	Tags        []string `json:"tags"`
	Duration    int      `json:"duration" binding:"min=0"`
	Description string   `json:"description"`
}

type VideoUpdateRequest struct {
	Title       string   `json:"title" binding:"omitempty,max=200"`
	CoverURL    string   `json:"cover_url"`
	VideoURL    string   `json:"video_url"`
	Category    string   `json:"category"`
	Tags        []string `json:"tags"`
	Duration    int      `json:"duration" binding:"omitempty,min=0"`
	Description string   `json:"description"`
	Status      int      `json:"status" binding:"omitempty,oneof=1 2 3"`
}

type VideoListRequest struct {
	Page     int    `form:"page" binding:"min=0"`
	PageSize int    `form:"page_size" binding:"min=0,max=100"`
	Category string `form:"category"`
	Keyword  string `form:"keyword"`
	Sort     string `form:"sort" binding:"omitempty,oneof=heat view_count upload_time"`
	Order    string `form:"order" binding:"omitempty,oneof=asc desc"`
}

type HotVideo struct {
	VideoID   int64   `json:"video_id"`
	Title     string  `json:"title"`
	CoverURL  string  `json:"cover_url"`
	Category  string  `json:"category"`
	Heat      float64 `json:"heat"`
	ViewCount int64   `json:"view_count"`
	Rank      int     `json:"rank"`
}

type VideoBatchCreateRequest struct {
	Videos []VideoCreateRequest `json:"videos" binding:"required,dive"`
}
