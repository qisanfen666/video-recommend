package service

import (
	"errors"
	"video-recommend/internal/domain"
	"video-recommend/internal/repository"
)

var (
	ErrVideoNotFound = errors.New("video not found")
)

type VideoService struct {
	videoRepo *repository.VideoRepository
}

func NewVideoService() *VideoService {
	return &VideoService{
		videoRepo: repository.NewVideoRepository(),
	}
}

func (s *VideoService) Create(req *domain.VideoCreateRequest) (*domain.Video, error) {
	video := &domain.Video{
		Title:       req.Title,
		CoverURL:    req.CoverURL,
		VideoURL:    req.VideoURL,
		Category:    req.Category,
		Duration:    req.Duration,
		Description: req.Description,
		Status:      1,
	}

	tagsStr := ""
	for i, tag := range req.Tags {
		if i > 0 {
			tagsStr += ","
		}
		tagsStr += tag
	}
	video.Tags = tagsStr

	if err := s.videoRepo.Create(video); err != nil {
		return nil, err
	}

	return video, nil
}

func (s *VideoService) GetByID(id int64) (*domain.Video, error) {
	video, err := s.videoRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}
	return video, nil
}

func (s *VideoService) Update(id int64, req *domain.VideoUpdateRequest) (*domain.Video, error) {
	video, err := s.videoRepo.GetByID(id)
	if err != nil {
		return nil, err
	}
	if video == nil {
		return nil, ErrVideoNotFound
	}

	if req.Title != "" {
		video.Title = req.Title
	}
	if req.CoverURL != "" {
		video.CoverURL = req.CoverURL
	}
	if req.VideoURL != "" {
		video.VideoURL = req.VideoURL
	}
	if req.Category != "" {
		video.Category = req.Category
	}
	if req.Duration > 0 {
		video.Duration = req.Duration
	}
	if req.Description != "" {
		video.Description = req.Description
	}
	if req.Status > 0 {
		video.Status = req.Status
	}
	if len(req.Tags) > 0 {
		tagsStr := ""
		for i, tag := range req.Tags {
			if i > 0 {
				tagsStr += ","
			}
			tagsStr += tag
		}
		video.Tags = tagsStr
	}

	if err := s.videoRepo.Update(video); err != nil {
		return nil, err
	}

	return video, nil
}

func (s *VideoService) Delete(id int64) error {
	return s.videoRepo.Delete(id)
}

func (s *VideoService) List(req *domain.VideoListRequest) ([]domain.Video, int64, error) {
	return s.videoRepo.List(req)
}

func (s *VideoService) IncrementViewCount(id int64) error {
	return s.videoRepo.IncrementViewCount(id)
}

func (s *VideoService) IncrementLikeCount(id int64) error {
	return s.videoRepo.IncrementLikeCount(id)
}

func (s *VideoService) UpdateHeat(id int64, heat float64) error {
	return s.videoRepo.UpdateHeat(id, heat)
}

func (s *VideoService) GetHotVideos(limit int, category string) ([]domain.Video, error) {
	return s.videoRepo.GetHotVideos(limit, category)
}

func (s *VideoService) GetVideosByIDs(ids []int64) ([]domain.Video, error) {
	return s.videoRepo.GetVideosByIDs(ids)
}

func (s *VideoService) GetAll() ([]domain.Video, error) {
	return s.videoRepo.GetAll()
}

func (s *VideoService) BatchCreate(req *domain.VideoBatchCreateRequest) ([]*domain.Video, error) {
	videos := make([]*domain.Video, 0, len(req.Videos))
	for _, videoReq := range req.Videos {
		video := &domain.Video{
			Title:       videoReq.Title,
			CoverURL:    videoReq.CoverURL,
			VideoURL:    videoReq.VideoURL,
			Category:    videoReq.Category,
			Duration:    videoReq.Duration,
			Description: videoReq.Description,
			Status:      1,
		}

		tagsStr := ""
		for i, tag := range videoReq.Tags {
			if i > 0 {
				tagsStr += ","
			}
			tagsStr += tag
		}
		video.Tags = tagsStr

		videos = append(videos, video)
	}

	if err := s.videoRepo.BatchCreate(videos); err != nil {
		return nil, err
	}

	return videos, nil
}

func (s *VideoService) DeleteAll() error {
	return s.videoRepo.DeleteAll()
}
