package main

import (
	"fmt"
	"log"
	"video-recommend/config"
	"video-recommend/internal/domain"
	"video-recommend/pkg/database"

	"gorm.io/gorm"
)

func main() {
	fmt.Println("Starting database migration...")

	cfg, err := config.LoadConfig("config/config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := database.InitMySQL(&cfg.Database); err != nil {
		log.Fatalf("Failed to connect to MySQL: %v", err)
	}
	defer database.CloseMySQL()

	if err := migrate(); err != nil {
		log.Fatalf("Migration failed: %v", err)
	}

	fmt.Println("Migration completed successfully!")
}

func migrate() error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(
			&domain.User{},
			&domain.Video{},
			&domain.Behavior{},
		); err != nil {
			return fmt.Errorf("auto migrate failed: %w", err)
		}

		if err := addIndexes(tx); err != nil {
			return fmt.Errorf("add indexes failed: %w", err)
		}

		return nil
	})
}

func addIndexes(tx *gorm.DB) error {
	if !tx.Migrator().HasIndex(&domain.User{}, "idx_users_username_email") {
		if err := tx.Exec("CREATE UNIQUE INDEX idx_users_username_email ON users(username, email)").Error; err != nil {
			log.Printf("Warning: failed to create composite index: %v", err)
		}
	}

	if !tx.Migrator().HasIndex(&domain.Video{}, "idx_videos_category_heat") {
		if err := tx.Exec("CREATE INDEX idx_videos_category_heat ON videos(category, heat DESC)").Error; err != nil {
			log.Printf("Warning: failed to create composite index: %v", err)
		}
	}

	if !tx.Migrator().HasIndex(&domain.Behavior{}, "idx_behaviors_user_video") {
		if err := tx.Exec("CREATE INDEX idx_behaviors_user_video ON behaviors(user_id, video_id)").Error; err != nil {
			log.Printf("Warning: failed to create composite index: %v", err)
		}
	}

	return nil
}
