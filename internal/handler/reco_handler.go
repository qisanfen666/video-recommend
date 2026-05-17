package handler

import (
	"video-recommend/internal/domain"
	"video-recommend/internal/middleware"
	"video-recommend/internal/service"
	"video-recommend/pkg/response"

	"github.com/gin-gonic/gin"
)

type RecoHandler struct {
	recoSvc *service.RecommendationService
}

func NewRecoHandler() *RecoHandler {
	return &RecoHandler{
		recoSvc: service.NewRecommendationService(),
	}
}

func (h *RecoHandler) Recommend(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req domain.RecommendRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.VideoCount = 10
	}

	if req.VideoCount <= 0 {
		req.VideoCount = 10
	}
	if req.VideoCount > 100 {
		req.VideoCount = 100
	}

	recos, err := h.recoSvc.RecommendForUser(userID, req.VideoCount)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, domain.RecommendResponse{
		Recommendations: recos,
		Algorithm:       "collaborative_filtering",
		Count:           len(recos),
	})
}

func (h *RecoHandler) SimilarUsers(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req domain.UserSimilarityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		req.Limit = 20
	}

	if req.Limit <= 0 {
		req.Limit = 20
	}
	if req.Limit > 50 {
		req.Limit = 50
	}

	users, err := h.recoSvc.FindSimilarUsers(userID, req.Limit)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, users)
}

func (h *RecoHandler) HotRank(c *gin.Context) {
	var req domain.HotRankRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		req.Limit = 20
	}

	hotRank, err := h.recoSvc.GetHotRank(&req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, hotRank)
}

func (h *RecoHandler) InvalidateCache(c *gin.Context) {
	var uri struct {
		Type string `uri:"type" binding:"required,oneof=user hot"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	switch uri.Type {
	case "hot":
		h.recoSvc.InvalidateHotCache()
	case "user":
		userID := middleware.GetUserID(c)
		h.recoSvc.InvalidateUserCache(userID)
	}

	response.SuccessWithMessage(c, "cache invalidated", nil)
}
