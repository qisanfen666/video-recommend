package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"video-recommend/config"
	"video-recommend/internal/handler"
	"video-recommend/internal/middleware"
	"video-recommend/pkg/cache"
	"video-recommend/pkg/database"
	"video-recommend/pkg/jwt"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitMySQL(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer database.CloseMySQL()

	if err := cache.InitRedis(&cfg.Redis); err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer cache.CloseRedis()

	jwt.InitJWT(&cfg.JWT)

	gin.SetMode(cfg.App.Mode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(middleware.CORS())

	setupRoutes(r)

	go func() {
		addr := cfg.App.Addr()
		log.Printf("Server starting on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")
}

func setupRoutes(r *gin.Engine) {
	userHandler := handler.NewUserHandler()
	videoHandler := handler.NewVideoHandler()
	recoHandler := handler.NewRecoHandler()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	api := r.Group("/api/v1")
	{
		auth := api.Group("/auth")
		{
			auth.POST("/register", userHandler.Register)
			auth.POST("/login", userHandler.Login)
		}

		videos := api.Group("/videos")
		{
			videos.GET("", videoHandler.List)
			videos.GET("/:id", videoHandler.Get)
			videos.POST("", middleware.Auth(), videoHandler.Create)
			videos.PUT("/:id", middleware.Auth(), videoHandler.Update)
			videos.DELETE("/:id", middleware.Auth(), middleware.AdminOnly(), videoHandler.Delete)
			videos.POST("/batch", middleware.Auth(), videoHandler.BatchCreate)
			videos.POST("/init", middleware.Auth(), videoHandler.InitSampleVideos)
			videos.DELETE("", middleware.Auth(), middleware.AdminOnly(), videoHandler.DeleteAll)
		}

		api.GET("/hot", recoHandler.HotRank)

		protected := api.Group("")
		protected.Use(middleware.Auth())
		{
			protected.GET("/profile", userHandler.GetProfile)
			protected.PUT("/profile", userHandler.UpdateProfile)
			protected.GET("/users", middleware.AdminOnly(), userHandler.List)

			protected.POST("/videos/:id/behavior", videoHandler.RecordBehavior)
			protected.GET("/videos/history", videoHandler.GetHistory)

			protected.POST("/recommend", recoHandler.Recommend)
			protected.POST("/similar-users", recoHandler.SimilarUsers)

			protected.DELETE("/cache/:type", recoHandler.InvalidateCache)
		}
	}

	fmt.Println("[Routes] API routes registered")
}
