package repository

import (
	"errors"
	"video-recommend/internal/domain"
	"video-recommend/pkg/database"

	"gorm.io/gorm"
)

type BehaviorRepository struct{}

func NewBehaviorRepository() *BehaviorRepository {
	return &BehaviorRepository{}
}

func (r *BehaviorRepository) Create(behavior *domain.Behavior) error {
	return database.DB.Create(behavior).Error
}

func (r *BehaviorRepository) GetByID(id int64) (*domain.Behavior, error) {
	var behavior domain.Behavior
	err := database.DB.First(&behavior, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &behavior, nil
}

func (r *BehaviorRepository) GetUserHistory(userID int64, req *domain.UserHistoryRequest) ([]domain.Behavior, int64, error) {
	var behaviors []domain.Behavior
	var total int64

	query := database.DB.Model(&domain.Behavior{}).Where("user_id = ?", userID)

	if req.Action != "" {
		query = query.Where("action = ?", req.Action)
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

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("timestamp desc").Find(&behaviors).Error; err != nil {
		return nil, 0, err
	}

	return behaviors, total, nil
}

func (r *BehaviorRepository) GetUserVideos(userID int64, actions []string) ([]int64, error) {
	var videoIDs []int64
	query := database.DB.Model(&domain.Behavior{}).
		Where("user_id = ?", userID).
		Where("action IN ?", actions).
		Select("DISTINCT video_id")

	if err := query.Pluck("video_id", &videoIDs).Error; err != nil {
		return nil, err
	}
	return videoIDs, nil
}

func (r *BehaviorRepository) GetVideoUsers(videoID int64, actions []string) ([]int64, error) {
	var userIDs []int64
	query := database.DB.Model(&domain.Behavior{}).
		Where("video_id = ?", videoID).
		Where("action IN ?", actions).
		Select("DISTINCT user_id")

	if err := query.Pluck("user_id", &userIDs).Error; err != nil {
		return nil, err
	}
	return userIDs, nil
}

func (r *BehaviorRepository) GetUserVideoMap(userIDs []int64, actions []string) (map[int64]map[int64]bool, error) {
	var behaviors []domain.Behavior
	err := database.DB.Model(&domain.Behavior{}).
		Where("user_id IN ?", userIDs).
		Where("action IN ?", actions).
		Find(&behaviors).Error
	if err != nil {
		return nil, err
	}

	result := make(map[int64]map[int64]bool)
	for _, b := range behaviors {
		if result[b.UserID] == nil {
			result[b.UserID] = make(map[int64]bool)
		}
		result[b.UserID][b.VideoID] = true
	}
	return result, nil
}

func (r *BehaviorRepository) GetVideoUserCount(videoID int64, actions []string) (int64, error) {
	var count int64
	err := database.DB.Model(&domain.Behavior{}).
		Where("video_id = ?", videoID).
		Where("action IN ?", actions).
		Count(&count).Error
	return count, err
}

func (r *BehaviorRepository) GetAll() ([]domain.Behavior, error) {
	var behaviors []domain.Behavior
	err := database.DB.Find(&behaviors).Error
	return behaviors, err
}

func (r *BehaviorRepository) DeleteByUserID(userID int64) error {
	return database.DB.Where("user_id = ?", userID).Delete(&domain.Behavior{}).Error
}
