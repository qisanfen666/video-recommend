package repository

import (
	"errors"
	"fmt"
	"video-recommend/internal/domain"
	"video-recommend/pkg/database"

	"gorm.io/gorm"
)

type VideoRepository struct{}

func NewVideoRepository() *VideoRepository {
	return &VideoRepository{}
}

func (r *VideoRepository) Create(video *domain.Video) error {
	return database.DB.Create(video).Error
}

func (r *VideoRepository) GetByID(id int64) (*domain.Video, error) {
	var video domain.Video
	err := database.DB.First(&video, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &video, nil
}

func (r *VideoRepository) Update(video *domain.Video) error {
	return database.DB.Save(video).Error
}

func (r *VideoRepository) Delete(id int64) error {
	return database.DB.Delete(&domain.Video{}, id).Error
}

func (r *VideoRepository) List(req *domain.VideoListRequest) ([]domain.Video, int64, error) {
	var videos []domain.Video
	var total int64

	query := database.DB.Model(&domain.Video{}).Where("status = ?", 1)

	if req.Category != "" {
		query = query.Where("category = ?", req.Category)
	}

	if req.Keyword != "" {
		keyword := fmt.Sprintf("%%%s%%", req.Keyword)
		query = query.Where("title LIKE ? OR description LIKE ?", keyword, keyword)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	orderBy := "id desc"
	if req.Sort != "" {
		order := "desc"
		if req.Order == "asc" {
			order = "asc"
		}
		orderBy = fmt.Sprintf("%s %s", req.Sort, order)
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order(orderBy).Find(&videos).Error; err != nil {
		return nil, 0, err
	}

	return videos, total, nil
}

func (r *VideoRepository) IncrementViewCount(id int64) error {
	return database.DB.Model(&domain.Video{}).Where("id = ?", id).
		UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
}

func (r *VideoRepository) IncrementLikeCount(id int64) error {
	return database.DB.Model(&domain.Video{}).Where("id = ?", id).
		UpdateColumn("like_count", gorm.Expr("like_count + ?", 1)).Error
}

func (r *VideoRepository) UpdateHeat(id int64, heat float64) error {
	return database.DB.Model(&domain.Video{}).Where("id = ?", id).
		Update("heat", heat).Error
}

func (r *VideoRepository) GetHotVideos(limit int, category string) ([]domain.Video, error) {
	var videos []domain.Video
	query := database.DB.Model(&domain.Video{}).Where("status = ?", 1)

	if category != "" {
		query = query.Where("category = ?", category)
	}

	if err := query.Order("heat desc").Limit(limit).Find(&videos).Error; err != nil {
		return nil, err
	}
	return videos, nil
}

func (r *VideoRepository) GetVideosByIDs(ids []int64) ([]domain.Video, error) {
	var videos []domain.Video
	if len(ids) == 0 {
		return videos, nil
	}
	err := database.DB.Where("id IN ?", ids).Find(&videos).Error
	return videos, err
}

func (r *VideoRepository) GetAll() ([]domain.Video, error) {
	var videos []domain.Video
	err := database.DB.Find(&videos).Error
	return videos, err
}

func (r *VideoRepository) BatchCreate(videos []*domain.Video) error {
	return database.DB.Create(videos).Error
}

func (r *VideoRepository) DeleteAll() error {
	return database.DB.Exec("TRUNCATE TABLE videos").Error
}
