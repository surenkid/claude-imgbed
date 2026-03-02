package main

import (
	"claude-imgbed"
	"claude-imgbed/internal/api"
	"claude-imgbed/internal/config"
	"claude-imgbed/internal/models"
	"fmt"
	"log"
	"os"
)

func main() {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Create uploads directory if it doesn't exist
	if err := os.MkdirAll(cfg.Upload.StoragePath, 0755); err != nil {
		log.Fatalf("Failed to create uploads directory: %v", err)
	}

	// Initialize recent uploads cache
	recentUploads := models.NewRecentUploads(cfg.Cache.RecentUploadsSize)

	// Setup router
	router := api.SetupRouter(cfg, recentUploads, claude_imgbed.StaticFS)

	// Start server
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting server on %s", addr)
	log.Printf("Auth token: %s", cfg.Auth.Token)
	log.Printf("Upload path: %s", cfg.Upload.StoragePath)
	log.Printf("Max file size: %d bytes (%.2f MB)", cfg.Upload.MaxSize, float64(cfg.Upload.MaxSize)/1024/1024)

	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
