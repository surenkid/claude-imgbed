package image

import (
	"fmt"
	"image"
	"os"
	"path/filepath"
	"time"

	"github.com/disintegration/imaging"
	"github.com/google/uuid"
)

type Storage struct {
	basePath string
	quality  int
}

func NewStorage(basePath string, quality int) *Storage {
	return &Storage{
		basePath: basePath,
		quality:  quality,
	}
}

type StorageResult struct {
	Filename      string
	RelativePath  string
	ThumbnailPath string
	FullPath      string
}

func (s *Storage) Save(img *ProcessedImage, extension string) (*StorageResult, error) {
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	// Create directory structure: uploads/YYYY/MM/
	dirPath := filepath.Join(s.basePath, year, month)
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Generate filename: YYYYMMDDhhmmss_uuid.ext
	timestamp := now.Format("20060102150405")
	uuidStr := uuid.New().String()[:8]
	filename := fmt.Sprintf("%s_%s%s", timestamp, uuidStr, extension)

	// Full paths
	fullPath := filepath.Join(dirPath, filename)
	thumbnailFilename := fmt.Sprintf("%s_%s_thumb%s", timestamp, uuidStr, extension)
	thumbnailPath := filepath.Join(dirPath, thumbnailFilename)

	// Save main image
	if err := s.saveImage(img.Image, fullPath, extension); err != nil {
		return nil, fmt.Errorf("failed to save image: %w", err)
	}

	// Save thumbnail
	if err := s.saveImage(img.Thumbnail, thumbnailPath, extension); err != nil {
		// Clean up main image if thumbnail fails
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to save thumbnail: %w", err)
	}

	// Relative paths for URL generation
	relativePath := filepath.Join(year, month, filename)
	relativeThumbnail := filepath.Join(year, month, thumbnailFilename)

	return &StorageResult{
		Filename:      filename,
		RelativePath:  relativePath,
		ThumbnailPath: relativeThumbnail,
		FullPath:      fullPath,
	}, nil
}

func (s *Storage) saveImage(img image.Image, path string, extension string) error {
	// Determine format and save with quality
	switch extension {
	case ".jpg", ".jpeg":
		return imaging.Save(img, path, imaging.JPEGQuality(s.quality))
	case ".png":
		return imaging.Save(img, path)
	case ".gif":
		return imaging.Save(img, path)
	case ".webp":
		// imaging library doesn't support WebP encoding directly
		// For WebP, we'll save as JPEG with high quality
		return imaging.Save(img, path, imaging.JPEGQuality(s.quality))
	default:
		return imaging.Save(img, path)
	}
}

func (s *Storage) GetImagePath(year, month, filename string) string {
	return filepath.Join(s.basePath, year, month, filename)
}
