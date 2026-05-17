package service

import (
	"container/heap"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"
	"video-recommend/config"
	"video-recommend/internal/domain"
	"video-recommend/internal/repository"
	"video-recommend/pkg/cache"
)

type RecommendationService struct {
	videoSvc     *VideoService
	behaviorRepo *repository.BehaviorRepository
}

func NewRecommendationService() *RecommendationService {
	return &RecommendationService{
		videoSvc:     NewVideoService(),
		behaviorRepo: repository.NewBehaviorRepository(),
	}
}

type recoItem struct {
	videoID int64
	score   float64
}

type recoHeap []recoItem

func (h recoHeap) Len() int           { return len(h) }
func (h recoHeap) Less(i, j int) bool { return h[i].score > h[j].score }
func (h recoHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *recoHeap) Push(x interface{}) {
	*h = append(*h, x.(recoItem))
}

func (h *recoHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func (s *RecommendationService) RecommendForUser(userID int64, count int) ([]domain.Recommendation, error) {
	ctx := context.Background()
	cacheKey := fmt.Sprintf(config.GlobalConfig.Cache.UserRecoKey, userID)

	cached, err := cache.RDB.Get(ctx, cacheKey).Result()
	if err == nil {
		var recos []domain.Recommendation
		if json.Unmarshal([]byte(cached), &recos) == nil && len(recos) >= count {
			return recos[:count], nil
		}
	}

	targetVideos, err := s.getUserWatchedVideos(userID)
	if err != nil {
		return nil, err
	}

	if len(targetVideos) == 0 {
		return s.getHotRecommendations(count, "")
	}

	similarUsers := s.findSimilarUsers(userID, targetVideos, 50)
	if len(similarUsers) == 0 {
		return s.getHotRecommendations(count, "")
	}

	candidates := s.buildCandidates(userID, targetVideos, similarUsers)

	recoHeap := &recoHeap{}
	heap.Init(recoHeap)

	for vid, score := range candidates {
		heap.Push(recoHeap, recoItem{videoID: vid, score: score})
		if recoHeap.Len() > count {
			heap.Pop(recoHeap)
		}
	}

	result := make([]domain.Recommendation, 0, recoHeap.Len())
	rank := 1
	for recoHeap.Len() > 0 {
		item := heap.Pop(recoHeap).(recoItem)
		video, err := s.videoSvc.GetByID(item.videoID)
		if err != nil || video == nil {
			continue
		}
		result = append(result, domain.Recommendation{
			VideoID:  video.ID,
			Title:    video.Title,
			CoverURL: video.CoverURL,
			Category: video.Category,
			Score:    item.score,
			Reason:   "similar_users",
			Rank:     rank,
		})
		rank++
	}

	if len(result) < count {
		hotRecos, _ := s.getHotRecommendations(count-len(result), "")
		result = append(result, hotRecos...)
	}

	if data, err := json.Marshal(result); err == nil {
		cache.RDB.Set(ctx, cacheKey, data, time.Duration(config.GlobalConfig.Cache.ExpireSeconds)*time.Second)
	}

	return result, nil
}

func (s *RecommendationService) getUserWatchedVideos(userID int64) (map[int64]bool, error) {
	videoIDs, err := s.behaviorRepo.GetUserVideos(userID, []string{domain.ActionWatch, domain.ActionKanwan})
	if err != nil {
		return nil, err
	}

	result := make(map[int64]bool)
	for _, vid := range videoIDs {
		result[vid] = true
	}
	return result, nil
}

func (s *RecommendationService) findSimilarUsers(userID int64, targetVideos map[int64]bool, k int) []domain.UserSim {
	candidates := make(map[int64]int)

	for vid := range targetVideos {
		userIDs, err := s.behaviorRepo.GetVideoUsers(vid, []string{domain.ActionWatch, domain.ActionKanwan, domain.ActionLike})
		if err != nil {
			continue
		}
		for _, uid := range userIDs {
			if uid != userID {
				candidates[uid]++
			}
		}
	}

	if len(candidates) == 0 {
		return nil
	}

	userVideoMap, err := s.behaviorRepo.GetUserVideoMap(getKeys(candidates), []string{domain.ActionWatch, domain.ActionKanwan})
	if err != nil {
		log.Printf("Error getting user video map: %v", err)
		return nil
	}

	simHeap := &simHeap{}
	heap.Init(simHeap)

	targetCount := len(targetVideos)
	for otherUser, common := range candidates {
		otherVideos := userVideoMap[otherUser]
		if len(otherVideos) == 0 {
			continue
		}

		union := targetCount + len(otherVideos) - common
		if union == 0 {
			continue
		}

		similarity := float64(common) / float64(union)

		heap.Push(simHeap, domain.UserSim{
			UserID:     otherUser,
			Similarity: similarity,
		})

		if simHeap.Len() > k {
			heap.Pop(simHeap)
		}
	}

	result := make([]domain.UserSim, simHeap.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(simHeap).(domain.UserSim)
	}
	return result
}

func (s *RecommendationService) buildCandidates(userID int64, targetVideos map[int64]bool, similarUsers []domain.UserSim) map[int64]float64 {
	candidates := make(map[int64]float64)

	for _, sim := range similarUsers {
		otherVideos, err := s.behaviorRepo.GetUserVideos(sim.UserID, []string{domain.ActionWatch, domain.ActionKanwan})
		if err != nil {
			continue
		}

		for _, vid := range otherVideos {
			if !targetVideos[vid] {
				candidates[vid] += sim.Similarity
			}
		}
	}

	return candidates
}

func (s *RecommendationService) getHotRecommendations(count int, category string) ([]domain.Recommendation, error) {
	ctx := context.Background()
	rankKey := config.GlobalConfig.Cache.HotRankKey

	if category == "" {
		cached, err := cache.RDB.Get(ctx, rankKey).Result()
		if err == nil {
			var recos []domain.Recommendation
			if json.Unmarshal([]byte(cached), &recos) == nil && len(recos) >= count {
				return recos[:count], nil
			}
		}
	}

	videos, err := s.videoSvc.GetHotVideos(count, category)
	if err != nil {
		return nil, err
	}

	result := make([]domain.Recommendation, 0, len(videos))
	for i, v := range videos {
		result = append(result, domain.Recommendation{
			VideoID:  v.ID,
			Title:    v.Title,
			CoverURL: v.CoverURL,
			Category: v.Category,
			Score:    v.Heat,
			Reason:   "hot",
			Rank:     i + 1,
		})
	}

	if category == "" && len(result) > 0 {
		if data, err := json.Marshal(result); err == nil {
			cache.RDB.Set(ctx, rankKey, data, 5*time.Minute)
		}
	}

	return result, nil
}

func (s *RecommendationService) FindSimilarUsers(userID int64, limit int) ([]domain.UserSim, error) {
	targetVideos, err := s.getUserWatchedVideos(userID)
	if err != nil {
		return nil, err
	}

	if len(targetVideos) == 0 {
		return []domain.UserSim{}, nil
	}

	return s.findSimilarUsers(userID, targetVideos, limit), nil
}

func (s *RecommendationService) GetHotRank(req *domain.HotRankRequest) (*domain.HotRankResponse, error) {
	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 100 {
		req.Limit = 100
	}

	videos, err := s.videoSvc.GetHotVideos(req.Limit, req.Category)
	if err != nil {
		return nil, err
	}

	ranks := make([]domain.HotVideo, 0, len(videos))
	for i, v := range videos {
		ranks = append(ranks, domain.HotVideo{
			VideoID:   v.ID,
			Title:     v.Title,
			CoverURL:  v.CoverURL,
			Category:  v.Category,
			Heat:      v.Heat,
			ViewCount: v.ViewCount,
			Rank:      i + 1,
		})
	}

	return &domain.HotRankResponse{
		Ranks:     ranks,
		UpdatedAt: time.Now().Format(time.RFC3339),
	}, nil
}

func (s *RecommendationService) InvalidateUserCache(userID int64) {
	ctx := context.Background()
	key := fmt.Sprintf(config.GlobalConfig.Cache.UserRecoKey, userID)
	cache.RDB.Del(ctx, key)
}

func (s *RecommendationService) InvalidateHotCache() {
	ctx := context.Background()
	cache.RDB.Del(ctx, config.GlobalConfig.Cache.HotRankKey)
}

type simHeap []domain.UserSim

func (h simHeap) Len() int           { return len(h) }
func (h simHeap) Less(i, j int) bool { return h[i].Similarity > h[j].Similarity }
func (h simHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *simHeap) Push(x interface{}) {
	*h = append(*h, x.(domain.UserSim))
}

func (h *simHeap) Pop() interface{} {
	old := *h
	n := len(old)
	item := old[n-1]
	*h = old[0 : n-1]
	return item
}

func getKeys(m map[int64]int) []int64 {
	keys := make([]int64, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func (s *RecommendationService) CacheWarmUp() error {
	ctx := context.Background()

	videos, err := s.videoSvc.GetHotVideos(100, "")
	if err != nil {
		return err
	}

	recos := make([]domain.Recommendation, 0, len(videos))
	for i, v := range videos {
		recos = append(recos, domain.Recommendation{
			VideoID:  v.ID,
			Title:    v.Title,
			CoverURL: v.CoverURL,
			Category: v.Category,
			Score:    v.Heat,
			Rank:     i + 1,
		})
	}

	if data, err := json.Marshal(recos); err == nil {
		cache.RDB.Set(ctx, config.GlobalConfig.Cache.HotRankKey, data, 5*time.Minute)
	}

	log.Println("[Cache] Warm up completed")
	return nil
}
