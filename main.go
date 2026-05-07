package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
	"video-recommend/data"
	"video-recommend/index"
	"video-recommend/internal"
	"video-recommend/live"
	"video-recommend/metrics"
	"video-recommend/reco"

	"github.com/gin-gonic/gin"
)

func main() {
	fmt.Println("[1] 程序启动:", time.Now().Format("15:04:05"))

	genReport, err := data.GenerateDataset(5000, 1000, 10000)
	if err != nil {
		panic(err)
	}
	log.Printf(
		"数据生成完成: 总耗时 %d ms | 落盘约 %.2f MiB (videos %d / users %d / behaviors %d 条)",
		genReport.TotalGenerationMs,
		float64(genReport.DiskBytes.Total)/(1024*1024),
		genReport.VideoCount, genReport.UserCount, genReport.BehaviorCount,
	)

	log.Println("正在加载数据到内存索引...")

	if err := internal.LoadAll(); err != nil {
		panic(err)
	}

	index.BuildAll()

	log.Println("数据加载完成")

	internal.Init()

	r := gin.Default()

	r.Use(metrics.Middleware())

	r.GET("/api/videos/:id", getVideoHandler)
	r.GET("/api/users/:id", getUserHandler)
	r.GET("/api/users/:id/history", getUserHistoryHandler)
	r.GET("/api/hot", getHotRankHandler)

	r.GET("/debug/stats", getStatsHandler)
	r.GET("/debug/recent", getRecentHandler)

	r.POST("/api/recommend", recommendHandler)
	r.POST("/api/similar-users", similarUsersHandler)
	r.POST("/api/admin/regenerate-data", regenerateDataHandler)
	r.POST("/api/admin/simulate-behaviors", simulateBehaviorsHandler)

	r.Static("/static", "./web")
	r.GET("/", func(c *gin.Context) {
		c.File("./web/index.html")
	})

	r.Run(":8081")
}

func regenerateDataHandler(c *gin.Context) {
	var req struct {
		VideoCount    int `json:"video_count"`
		UserCount     int `json:"user_count"`
		BehaviorCount int `json:"behavior_count"`
	}
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	genReport, err := data.GenerateDataset(req.VideoCount, req.UserCount, req.BehaviorCount)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var reload data.ReloadStat
	t0 := time.Now()
	if err := internal.LoadAll(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      err.Error(),
			"generation": genReport,
		})
		return
	}
	reload.LoadMs = time.Since(t0).Milliseconds()

	t1 := time.Now()
	index.BuildAll()
	reload.IndexMs = time.Since(t1).Milliseconds()

	t2 := time.Now()
	internal.Init()
	reload.InitMs = time.Since(t2).Milliseconds()
	reload.TotalMs = reload.LoadMs + reload.IndexMs + reload.InitMs

	c.JSON(http.StatusOK, gin.H{
		"generation": genReport,
		"reload":     reload,
	})
}

func simulateBehaviorsHandler(c *gin.Context) {
	var req live.SimulateRequest
	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}
	rep, err := live.SimulateLiveBehaviors(req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, rep)
}

func getVideoHandler(c *gin.Context) {
	id := c.Param("id")
	var videoID int64
	if _, err := fmt.Sscanf(id, "%d", &videoID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid video id"})
		return
	}

	video, ok := internal.GetVideo(videoID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "video not found"})
		return
	}

	c.JSON(http.StatusOK, video)
}

func getUserHandler(c *gin.Context) {
	id := c.Param("id")
	var userID int64
	if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	user, ok := internal.GetUser(userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user not found"})
		return
	}

	c.JSON(http.StatusOK, user)
}

func getUserHistoryHandler(c *gin.Context) {
	id := c.Param("id")
	var userID int64
	if _, err := fmt.Sscanf(id, "%d", &userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid user id"})
		return
	}

	history, ok := internal.GetBehaviors(userID)
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": "user history not found"})
		return
	}

	c.JSON(http.StatusOK, history)
}

func getHotRankHandler(c *gin.Context) {
	top20 := internal.GetTop20()
	c.JSON(http.StatusOK, gin.H{
		"count": len(top20),
		"list":  top20,
	})
}

func getStatsHandler(c *gin.Context) {
	stats := metrics.GetAllStats()
	c.JSON(http.StatusOK, gin.H{
		"interfaces": stats,
		"summary": gin.H{
			"total_interfaces": len(stats),
		},
	})
}

func getRecentHandler(c *gin.Context) {
	recent := metrics.GetRecent(10)
	c.JSON(http.StatusOK, gin.H{
		"recent_requests": recent,
	})
}

func recommendHandler(c *gin.Context) {
	var req struct {
		UserID   int64 `json:"user_id"`
		PageSize int   `json:"page_size"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request",
		})
		return
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}

	start := time.Now()
	result := reco.RecommendForUser(req.UserID, req.PageSize)
	duration := time.Since(start)

	videos := make([]gin.H, 0, len(result))
	for _, r := range result {
		if v, ok := internal.GetVideo(r.VideoID); ok {
			videos = append(videos, gin.H{
				"id":     v.ID,
				"title":  v.Title,
				"score":  r.Score,
				"reason": r.Reason,
			})
		}
	}

	c.JSON(200, gin.H{
		"user_id": req.UserID,
		"videos":  videos,
		"count":   len(videos),
		"time_ms": duration.Milliseconds(),
	})
}

func similarUsersHandler(c *gin.Context) {
	var req struct {
		UserID int64 `json:"user_id"`
		TopK   int   `json:"top_k"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{
			"error": "invalid request",
		})
		return
	}
	if req.TopK <= 0 {
		req.TopK = 10
	}

	start := time.Now()
	sims := reco.FindSimilarUsers(req.UserID, req.TopK)
	duration := time.Since(start)

	users := make([]gin.H, 0, len(sims))
	for _, s := range sims {
		if u, ok := internal.GetUser(s.UserID); ok {
			users = append(users, gin.H{
				"id":            u.ID,
				"name":          u.Name,
				"similarity":    s.Similarity,
				"common_videos": s.CommonVideos,
			})
		}
	}

	c.JSON(200, gin.H{
		"user_id": req.UserID,
		"users":   users,
		"count":   len(users),
		"time_ms": duration.Milliseconds(),
	})
}
