package handler

import (
	"video-recommend/internal/domain"
	"video-recommend/internal/middleware"
	"video-recommend/internal/service"
	"video-recommend/pkg/response"

	"github.com/gin-gonic/gin"
)

type VideoHandler struct {
	videoSvc *service.VideoService
}

func NewVideoHandler() *VideoHandler {
	return &VideoHandler{
		videoSvc: service.NewVideoService(),
	}
}

func (h *VideoHandler) Create(c *gin.Context) {
	var req domain.VideoCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	video, err := h.videoSvc.Create(&req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, video)
}

func (h *VideoHandler) Get(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	video, err := h.videoSvc.GetByID(uri.ID)
	if err != nil {
		if err == service.ErrVideoNotFound {
			response.NotFound(c, "video not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, video)
}

func (h *VideoHandler) Update(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req domain.VideoUpdateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	video, err := h.videoSvc.Update(uri.ID, &req)
	if err != nil {
		if err == service.ErrVideoNotFound {
			response.NotFound(c, "video not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, video)
}

func (h *VideoHandler) Delete(c *gin.Context) {
	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if err := h.videoSvc.Delete(uri.ID); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "deleted", nil)
}

func (h *VideoHandler) List(c *gin.Context) {
	var req domain.VideoListRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	videos, total, err := h.videoSvc.List(&req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 {
		pageSize = 20
	}

	response.PageSuccess(c, total, page, pageSize, videos)
}

func (h *VideoHandler) RecordBehavior(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var uri struct {
		ID int64 `uri:"id" binding:"required"`
	}
	if err := c.ShouldBindUri(&uri); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	var req domain.BehaviorCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	req.VideoID = uri.ID

	behaviorSvc := service.NewBehaviorService()
	behavior, err := behaviorSvc.Record(userID, &req, c.ClientIP(), c.GetHeader("User-Agent"))
	if err != nil {
		if err == service.ErrVideoNotFound {
			response.NotFound(c, "video not found")
			return
		}
		response.InternalServerError(c, err.Error())
		return
	}

	response.Success(c, behavior)
}

func (h *VideoHandler) GetHistory(c *gin.Context) {
	userID := middleware.GetUserID(c)

	var req domain.UserHistoryRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	if req.Page < 1 {
		req.Page = 1
	}
	if req.PageSize < 1 {
		req.PageSize = 20
	}

	behaviorSvc := service.NewBehaviorService()
	history, total, err := behaviorSvc.GetUserHistory(userID, &req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.PageSuccess(c, total, req.Page, req.PageSize, history)
}

func (h *VideoHandler) BatchCreate(c *gin.Context) {
	var req domain.VideoBatchCreateRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, err.Error())
		return
	}

	videos, err := h.videoSvc.BatchCreate(&req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.Created(c, videos)
}

func (h *VideoHandler) InitSampleVideos(c *gin.Context) {
	sampleVideos := []domain.VideoCreateRequest{
		{
			Title:       "Go语言从入门到精通",
			CoverURL:    "https://picsum.photos/200/150?random=1",
			VideoURL:    "https://example.com/video1.mp4",
			Category:    "education",
			Tags:        []string{"programming", "go", "tutorial"},
			Duration:    3600,
			Description: "完整的Go语言学习教程，适合零基础学员",
		},
		{
			Title:       "分布式系统架构实战",
			CoverURL:    "https://picsum.photos/200/150?random=2",
			VideoURL:    "https://example.com/video2.mp4",
			Category:    "education",
			Tags:        []string{"architecture", "distributed", "system"},
			Duration:    4800,
			Description: "深入解析分布式系统的设计与实现",
		},
		{
			Title:       "机器学习算法解析",
			CoverURL:    "https://picsum.photos/200/150?random=3",
			VideoURL:    "https://example.com/video3.mp4",
			Category:    "education",
			Tags:        []string{"machine learning", "ai", "algorithms"},
			Duration:    5400,
			Description: "机器学习核心算法的详细讲解",
		},
		{
			Title:       "美食探店 - 城市美食地图",
			CoverURL:    "https://picsum.photos/200/150?random=4",
			VideoURL:    "https://example.com/video4.mp4",
			Category:    "food",
			Tags:        []string{"food", "travel", "restaurant"},
			Duration:    1200,
			Description: "探索城市里的美食小店",
		},
		{
			Title:       "家常菜教程 - 红烧肉",
			CoverURL:    "https://picsum.photos/200/150?random=5",
			VideoURL:    "https://example.com/video5.mp4",
			Category:    "food",
			Tags:        []string{"cooking", "recipe", "pork"},
			Duration:    1800,
			Description: "详细讲解家常红烧肉的做法",
		},
		{
			Title:       "游戏实况 - 王者荣耀",
			CoverURL:    "https://picsum.photos/200/150?random=6",
			VideoURL:    "https://example.com/video6.mp4",
			Category:    "game",
			Tags:        []string{"game", "moba", "stream"},
			Duration:    3000,
			Description: "王者荣耀精彩对局实况直播",
		},
		{
			Title:       "独立游戏测评",
			CoverURL:    "https://picsum.photos/200/150?random=7",
			VideoURL:    "https://example.com/video7.mp4",
			Category:    "game",
			Tags:        []string{"game", "indie", "review"},
			Duration:    2400,
			Description: "最新独立游戏的深度测评",
		},
		{
			Title:       "健身教程 - 腹肌训练",
			CoverURL:    "https://picsum.photos/200/150?random=8",
			VideoURL:    "https://example.com/video8.mp4",
			Category:    "sports",
			Tags:        []string{"fitness", "abs", "workout"},
			Duration:    1800,
			Description: "15分钟腹肌训练计划",
		},
		{
			Title:       "马拉松训练指南",
			CoverURL:    "https://picsum.photos/200/150?random=9",
			VideoURL:    "https://example.com/video9.mp4",
			Category:    "sports",
			Tags:        []string{"running", "marathon", "training"},
			Duration:    3600,
			Description: "完整的马拉松训练计划分享",
		},
		{
			Title:       "音乐MV - 流行金曲",
			CoverURL:    "https://picsum.photos/200/150?random=10",
			VideoURL:    "https://example.com/video10.mp4",
			Category:    "music",
			Tags:        []string{"music", "mv", "pop"},
			Duration:    240,
			Description: "2024年最火流行歌曲MV合集",
		},
	}

	req := &domain.VideoBatchCreateRequest{Videos: sampleVideos}
	videos, err := h.videoSvc.BatchCreate(req)
	if err != nil {
		response.InternalServerError(c, err.Error())
		return
	}

	response.SuccessWithMessage(c, "successfully initialized 10 sample videos", videos)
}

func (h *VideoHandler) DeleteAll(c *gin.Context) {
	if err := h.videoSvc.DeleteAll(); err != nil {
		response.InternalServerError(c, err.Error())
		return
	}
	response.SuccessWithMessage(c, "all videos deleted", nil)
}
