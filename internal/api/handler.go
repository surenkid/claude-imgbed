package api

import (
	"fmt"
	"claude-imgbed/internal/config"
	"claude-imgbed/internal/image"
	"claude-imgbed/internal/models"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/gin-gonic/gin"
)

type Handler struct {
	cfg            *config.Config
	validator      *image.Validator
	processor      *image.Processor
	storage        *image.Storage
	recentUploads  *models.RecentUploads
}

func NewHandler(cfg *config.Config, recentUploads *models.RecentUploads) *Handler {
	validator := image.NewValidator(cfg.Upload.MaxSize, cfg.Upload.AllowedTypes)
	processor := image.NewProcessor(cfg.Image.MaxDimension, cfg.Image.Quality, cfg.Image.ThumbnailSize)
	storage := image.NewStorage(cfg.Upload.StoragePath, cfg.Image.Quality)

	return &Handler{
		cfg:           cfg,
		validator:     validator,
		processor:     processor,
		storage:       storage,
		recentUploads: recentUploads,
	}
}

func (h *Handler) UploadImage(c *gin.Context) {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   "no file uploaded",
		})
		return
	}

	// Validate file
	if err := h.validator.Validate(file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"success": false,
			"error":   err.Error(),
		})
		return
	}

	// Process image (resize, compress, generate thumbnail)
	processed, err := h.processor.Process(file)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to process image",
		})
		return
	}

	// Get file extension
	contentType := file.Header.Get("Content-Type")
	extension := h.validator.GetExtension(contentType)

	// Save to storage
	result, err := h.storage.Save(processed, extension)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"success": false,
			"error":   "failed to save image",
		})
		return
	}

	// Build URLs (assuming the server is accessible at the request host)
	scheme := "http"
	if c.Request.TLS != nil {
		scheme = "https"
	}
	host := c.Request.Host
	imageURL := fmt.Sprintf("%s://%s/images/%s", scheme, host, result.RelativePath)
	thumbnailURL := fmt.Sprintf("%s://%s/images/%s", scheme, host, result.ThumbnailPath)

	// Create upload record
	record := models.UploadRecord{
		URL:        imageURL,
		Thumbnail:  thumbnailURL,
		Filename:   result.Filename,
		Size:       file.Size,
		Width:      processed.Width,
		Height:     processed.Height,
		UploadedAt: time.Now(),
	}

	// Add to recent uploads
	h.recentUploads.Add(record)

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    record,
	})
}

func (h *Handler) GetRecentUploads(c *gin.Context) {
	uploads := h.recentUploads.GetAll()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    uploads,
	})
}

func (h *Handler) ServeImage(c *gin.Context) {
	year := c.Param("year")
	month := c.Param("month")
	filename := c.Param("filename")

	// Get image path
	imagePath := h.storage.GetImagePath(year, month, filename)

	// Check if file exists
	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		c.JSON(http.StatusNotFound, gin.H{
			"success": false,
			"error":   "image not found",
		})
		return
	}

	// Set cache headers for CDN
	c.Header("Cache-Control", "public, max-age=31536000")
	c.Header("CDN-Cache-Control", "max-age=31536000")

	// Determine content type from extension
	ext := filepath.Ext(filename)
	contentType := "image/jpeg"
	switch ext {
	case ".png":
		contentType = "image/png"
	case ".gif":
		contentType = "image/gif"
	case ".webp":
		contentType = "image/webp"
	}

	c.Header("Content-Type", contentType)
	c.File(imagePath)
}

func (h *Handler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":    "ok",
		"timestamp": time.Now().Format(time.RFC3339),
	})
}
