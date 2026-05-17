package service

import (
	"context"
	"fmt"
	"time"
	"video-recommend/config"
	"video-recommend/internal/domain"
	"video-recommend/internal/repository"
	"video-recommend/pkg/cache"
)

type BehaviorService struct {
	behaviorRepo *repository.BehaviorRepository
	videoSvc     *VideoService
}

func NewBehaviorService() *BehaviorService {
	return &BehaviorService{
		behaviorRepo: repository.NewBehaviorRepository(),
		videoSvc:     NewVideoService(),
	}
}

func (s *BehaviorService) Record(userID int64, req *domain.BehaviorCreateRequest, ip, device string) (*domain.Behavior, error) {
	video, err := s.videoSvc.GetByID(req.VideoID)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}

	behavior := &domain.Behavior{
		UserID:    userID,
		VideoID:   req.VideoID,
		Action:    req.Action,
		WatchTime: req.WatchTime,
		IP:        ip,
		Device:    device,
		TimeStamp: time.Now(),
	}

	if err := s.behaviorRepo.Create(behavior); err != nil {
		return nil, err
	}

	go s.updateVideoStats(req.VideoID, req.Action)

	s.invalidateUserRecommendCache(userID)

	return behavior, nil
}

func (s *BehaviorService) updateVideoStats(videoID int64, action string) {
	switch action {
	case domain.ActionWatch, domain.ActionKanwan:
		s.videoSvc.IncrementViewCount(videoID)
	case domain.ActionLike:
		s.videoSvc.IncrementLikeCount(videoID)
	}
}

func (s *BehaviorService) invalidateUserRecommendCache(userID int64) {
	key := fmt.Sprintf(config.GlobalConfig.Cache.UserRecoKey, userID)
	cache.RDB.Del(context.Background(), key)
}

func (s *BehaviorService) GetUserHistory(userID int64, req *domain.UserHistoryRequest) ([]domain.UserHistoryResponse, int64, error) {
	behaviors, total, err := s.behaviorRepo.GetUserHistory(userID, req)
	if err != nil {
		return nil, 0, err
	}

	var resp []domain.UserHistoryResponse
	videoIDs := make([]int64, 0)
	for _, b := range behaviors {
		videoIDs = append(videoIDs, b.VideoID)
	}

	videos, err := s.videoSvc.GetVideosByIDs(videoIDs)
	if err != nil {
		return nil, 0, err
	}

	videoMap := make(map[int64]*domain.Video)
	for i := range videos {
		videoMap[videos[i].ID] = &videos[i]
	}

	for _, b := range behaviors {
		h := domain.UserHistoryResponse{
			VideoID:   b.VideoID,
			Action:    b.Action,
			WatchTime: b.WatchTime,
			Timestamp: b.TimeStamp,
		}
		if v, ok := videoMap[b.VideoID]; ok {
			h.Title = v.Title
			h.CoverURL = v.CoverURL
		}
		resp = append(resp, h)
	}

	return resp, total, nil
}

func (s *BehaviorService) GetUserVideos(userID int64) ([]int64, error) {
	return s.behaviorRepo.GetUserVideos(userID, []string{domain.ActionWatch, domain.ActionKanwan})
}

func (s *BehaviorService) GetVideoUsers(videoID int64) ([]int64, error) {
	return s.behaviorRepo.GetVideoUsers(videoID, []string{domain.ActionWatch, domain.ActionKanwan, domain.ActionLike})
}

func (s *BehaviorService) GetAllBehaviors() ([]domain.Behavior, error) {
	return s.behaviorRepo.GetAll()
}
