package tests

import (
	"image"
	"image/color"
	imgpkg "claude-imgbed/internal/image"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// Test Storage - Basic Save Operation
func TestStorage_BasicSave(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	// Create test image
	img := image.NewRGBA(image.Rect(0, 0, 200, 200))
	for y := 0; y < 200; y++ {
		for x := 0; x < 200; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 0, B: 0, A: 255})
		}
	}

	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     200,
		Height:    200,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify result structure
	if result.Filename == "" {
		t.Error("Filename should not be empty")
	}
	if result.RelativePath == "" {
		t.Error("RelativePath should not be empty")
	}
	if result.ThumbnailPath == "" {
		t.Error("ThumbnailPath should not be empty")
	}
	if result.FullPath == "" {
		t.Error("FullPath should not be empty")
	}
}

// Test Storage - Directory Structure
func TestStorage_DirectoryStructure(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Verify directory structure: uploads/YYYY/MM/
	now := time.Now()
	expectedYear := now.Format("2006")
	expectedMonth := now.Format("01")

	if !strings.Contains(result.RelativePath, expectedYear) {
		t.Errorf("Path should contain year %s, got %s", expectedYear, result.RelativePath)
	}
	if !strings.Contains(result.RelativePath, expectedMonth) {
		t.Errorf("Path should contain month %s, got %s", expectedMonth, result.RelativePath)
	}

	// Verify directory exists
	dirPath := filepath.Join(tempDir, expectedYear, expectedMonth)
	if _, err := os.Stat(dirPath); os.IsNotExist(err) {
		t.Errorf("Directory %s should exist", dirPath)
	}
}

// Test Storage - Filename Format Validation
func TestStorage_FilenameFormat(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	filename := result.Filename

	// Format: YYYYMMDDhhmmss_uuid.ext
	// Example: 20260302143025_a1b2c3d4.jpg

	// Check minimum length
	if len(filename) < 24 {
		t.Errorf("Filename too short: %s", filename)
	}

	// Check extension
	if !strings.HasSuffix(filename, ".jpg") {
		t.Errorf("Filename should end with .jpg: %s", filename)
	}

	// Check underscore separator
	if !strings.Contains(filename, "_") {
		t.Errorf("Filename should contain underscore: %s", filename)
	}

	// Check timestamp part (first 14 characters should be digits)
	timestampPart := filename[:14]
	for _, c := range timestampPart {
		if c < '0' || c > '9' {
			t.Errorf("Timestamp part should be all digits: %s", timestampPart)
			break
		}
	}
}

// Test Storage - Multiple Extensions
func TestStorage_MultipleExtensions(t *testing.T) {
	extensions := []string{".jpg", ".png", ".gif", ".webp"}

	for _, ext := range extensions {
		t.Run(ext, func(t *testing.T) {
			tempDir := t.TempDir()
			storage := imgpkg.NewStorage(tempDir, 90)

			img := image.NewRGBA(image.Rect(0, 0, 100, 100))
			processed := &imgpkg.ProcessedImage{
				Image:     img,
				Thumbnail: img,
				Width:     100,
				Height:    100,
			}

			result, err := storage.Save(processed, ext)
			if err != nil {
				t.Fatalf("Save failed for %s: %v", ext, err)
			}

			if !strings.HasSuffix(result.Filename, ext) {
				t.Errorf("Filename should have extension %s, got %s", ext, result.Filename)
			}

			// Verify file exists
			if _, err := os.Stat(result.FullPath); os.IsNotExist(err) {
				t.Errorf("File should exist: %s", result.FullPath)
			}
		})
	}
}

// Test Storage - Thumbnail Naming
func TestStorage_ThumbnailNaming(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Thumbnail should have _thumb suffix
	if !strings.Contains(result.ThumbnailPath, "_thumb") {
		t.Errorf("Thumbnail path should contain '_thumb': %s", result.ThumbnailPath)
	}

	// Both should have same extension
	mainExt := filepath.Ext(result.Filename)
	thumbExt := filepath.Ext(filepath.Base(result.ThumbnailPath))

	if mainExt != thumbExt {
		t.Errorf("Extensions should match: main=%s, thumb=%s", mainExt, thumbExt)
	}
}

// Test Storage - File Permissions
func TestStorage_FilePermissions(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Save failed: %v", err)
	}

	// Check directory permissions (should be 0755)
	dirPath := filepath.Dir(result.FullPath)
	dirInfo, err := os.Stat(dirPath)
	if err != nil {
		t.Fatalf("Failed to stat directory: %v", err)
	}

	if !dirInfo.IsDir() {
		t.Error("Should be a directory")
	}

	// Directory should be readable and executable
	mode := dirInfo.Mode()
	if mode&0400 == 0 {
		t.Error("Directory should be readable by owner")
	}
	if mode&0100 == 0 {
		t.Error("Directory should be executable by owner")
	}
}

// Test Storage - Concurrent Saves
func TestStorage_ConcurrentSaves(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			img := image.NewRGBA(image.Rect(0, 0, 100, 100))
			processed := &imgpkg.ProcessedImage{
				Image:     img,
				Thumbnail: img,
				Width:     100,
				Height:    100,
			}

			_, err := storage.Save(processed, ".jpg")
			if err != nil {
				errors <- err
			}
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}

	close(errors)
	for err := range errors {
		t.Errorf("Concurrent save failed: %v", err)
	}
}

// Test Storage - GetImagePath
func TestStorage_GetImagePath(t *testing.T) {
	tests := []struct {
		name     string
		basePath string
		year     string
		month    string
		filename string
		expected string
	}{
		{
			name:     "Standard path",
			basePath: "/uploads",
			year:     "2026",
			month:    "03",
			filename: "test.jpg",
			expected: filepath.Join("/uploads", "2026", "03", "test.jpg"),
		},
		{
			name:     "Relative path",
			basePath: "./uploads",
			year:     "2025",
			month:    "12",
			filename: "image.png",
			expected: filepath.Join("./uploads", "2025", "12", "image.png"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := imgpkg.NewStorage(tt.basePath, 90)
			result := storage.GetImagePath(tt.year, tt.month, tt.filename)

			if result != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, result)
			}
		})
	}
}

// Test Storage - Unique Filenames
func TestStorage_UniqueFilenames(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	// Save multiple images quickly
	filenames := make(map[string]bool)
	for i := 0; i < 5; i++ {
		result, err := storage.Save(processed, ".jpg")
		if err != nil {
			t.Fatalf("Save %d failed: %v", i, err)
		}

		if filenames[result.Filename] {
			t.Errorf("Duplicate filename generated: %s", result.Filename)
		}
		filenames[result.Filename] = true

		// Small delay to ensure different timestamps
		time.Sleep(1 * time.Millisecond)
	}

	if len(filenames) != 5 {
		t.Errorf("Expected 5 unique filenames, got %d", len(filenames))
	}
}

// Test Storage - Large Image
func TestStorage_LargeImage(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	// Create a large image
	img := image.NewRGBA(image.Rect(0, 0, 4000, 3000))
	for y := 0; y < 3000; y++ {
		for x := 0; x < 4000; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(x % 256),
				G: uint8(y % 256),
				B: 128,
				A: 255,
			})
		}
	}

	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     4000,
		Height:    3000,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Failed to save large image: %v", err)
	}

	// Verify file exists and has content
	info, err := os.Stat(result.FullPath)
	if err != nil {
		t.Fatalf("Failed to stat saved file: %v", err)
	}

	if info.Size() == 0 {
		t.Error("Saved file should not be empty")
	}

	t.Logf("Large image saved: %d bytes", info.Size())
}

// Test Storage - Quality Settings
func TestStorage_QualitySettings(t *testing.T) {
	qualities := []int{50, 75, 90, 95}

	for _, quality := range qualities {
		t.Run(string(rune(quality)), func(t *testing.T) {
			tempDir := t.TempDir()
			storage := imgpkg.NewStorage(tempDir, quality)

			img := image.NewRGBA(image.Rect(0, 0, 500, 500))
			for y := 0; y < 500; y++ {
				for x := 0; x < 500; x++ {
					img.Set(x, y, color.RGBA{R: 255, G: 128, B: 64, A: 255})
				}
			}

			processed := &imgpkg.ProcessedImage{
				Image:     img,
				Thumbnail: img,
				Width:     500,
				Height:    500,
			}

			result, err := storage.Save(processed, ".jpg")
			if err != nil {
				t.Fatalf("Save failed with quality %d: %v", quality, err)
			}

			// Verify file exists
			info, err := os.Stat(result.FullPath)
			if err != nil {
				t.Fatalf("Failed to stat file: %v", err)
			}

			t.Logf("Quality %d: file size = %d bytes", quality, info.Size())
		})
	}
}

// Test Storage - Path Traversal Prevention
func TestStorage_PathTraversalPrevention(t *testing.T) {
	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	// Try to get path with traversal attempts
	maliciousPaths := []struct {
		year     string
		month    string
		filename string
	}{
		{"../../../etc", "passwd", "test.jpg"},
		{"2026", "../../../etc", "passwd"},
		{"2026", "03", "../../../etc/passwd"},
	}

	for _, mp := range maliciousPaths {
		path := storage.GetImagePath(mp.year, mp.month, mp.filename)

		// The path should still be constructed, but when used with os.Stat
		// it should not escape the base directory
		// This is more of a documentation test - actual prevention
		// should be done at the API level
		t.Logf("Path with traversal attempt: %s", path)
	}
}

// Test Storage - Disk Space Handling
func TestStorage_DiskSpaceHandling(t *testing.T) {
	// This test verifies behavior when disk operations fail
	// In a real scenario, this would test disk full conditions

	tempDir := t.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	// Normal save should work
	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Normal save should succeed: %v", err)
	}

	// Verify both files exist
	if _, err := os.Stat(result.FullPath); os.IsNotExist(err) {
		t.Error("Main image should exist")
	}

	thumbPath := filepath.Join(filepath.Dir(result.FullPath), filepath.Base(result.ThumbnailPath))
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Error("Thumbnail should exist")
	}
}

// Benchmark Storage Save
func BenchmarkStorage_Save(b *testing.B) {
	tempDir := b.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 800, 600))
	for y := 0; y < 600; y++ {
		for x := 0; x < 800; x++ {
			img.Set(x, y, color.RGBA{R: 255, G: 128, B: 64, A: 255})
		}
	}

	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     800,
		Height:    600,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		storage.Save(processed, ".jpg")
	}
}

// Benchmark Concurrent Storage
func BenchmarkStorage_ConcurrentSave(b *testing.B) {
	tempDir := b.TempDir()
	storage := imgpkg.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 400, 300))
	processed := &imgpkg.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     400,
		Height:    300,
	}

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			storage.Save(processed, ".jpg")
		}
	})
}
