package tests

import (
	"bytes"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"claude-imgbed/internal/image"
	"io"
	"mime/multipart"
	"net/textproto"
	"os"
	"path/filepath"
	"testing"
)

// Helper function to create a test image
func createTestImage(width, height int, format string) ([]byte, error) {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a gradient
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / width),
				G: uint8((y * 255) / height),
				B: 128,
				A: 255,
			})
		}
	}

	buf := new(bytes.Buffer)
	var err error
	switch format {
	case "jpeg", "jpg":
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	case "png":
		err = png.Encode(buf, img)
	default:
		err = jpeg.Encode(buf, img, &jpeg.Options{Quality: 90})
	}

	return buf.Bytes(), err
}

// Helper function to create multipart.FileHeader from bytes
func createFileHeader(filename, contentType string, data []byte) *multipart.FileHeader {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	h := make(textproto.MIMEHeader)
	h.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	h.Set("Content-Type", contentType)

	part, _ := writer.CreatePart(h)
	part.Write(data)
	writer.Close()

	reader := multipart.NewReader(body, writer.Boundary())
	form, _ := reader.ReadForm(10 << 20)

	if files, ok := form.File["file"]; ok && len(files) > 0 {
		return files[0]
	}
	return nil
}

// Test Validator - File Size Validation
func TestValidator_FileSizeLimit(t *testing.T) {
	validator := image.NewValidator(1024*1024, []string{"image/jpeg", "image/png"}) // 1MB limit

	// Create a 2MB image (should fail)
	largeData := make([]byte, 2*1024*1024)
	fileHeader := createFileHeader("large.jpg", "image/jpeg", largeData)

	err := validator.Validate(fileHeader)
	if err != image.ErrFileTooLarge {
		t.Errorf("Expected ErrFileTooLarge, got %v", err)
	}

	// Create a small image (should pass size check)
	smallData, _ := createTestImage(100, 100, "jpeg")
	fileHeader = createFileHeader("small.jpg", "image/jpeg", smallData)

	// This might fail on magic number, but should pass size check
	err = validator.Validate(fileHeader)
	if err == image.ErrFileTooLarge {
		t.Errorf("Small file should not trigger ErrFileTooLarge")
	}
}

// Test Validator - Format Validation
func TestValidator_FormatValidation(t *testing.T) {
	validator := image.NewValidator(5*1024*1024, []string{"image/jpeg", "image/png"})

	tests := []struct {
		name        string
		contentType string
		shouldFail  bool
	}{
		{"Valid JPEG", "image/jpeg", false},
		{"Valid PNG", "image/png", false},
		{"Invalid GIF", "image/gif", true},
		{"Invalid WebP", "image/webp", true},
		{"Invalid PDF", "application/pdf", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, _ := createTestImage(100, 100, "jpeg")
			fileHeader := createFileHeader("test.jpg", tt.contentType, data)

			err := validator.Validate(fileHeader)
			if tt.shouldFail && err != image.ErrInvalidFormat {
				t.Errorf("Expected ErrInvalidFormat for %s, got %v", tt.contentType, err)
			}
		})
	}
}

// Test Validator - Magic Number Check
func TestValidator_MagicNumberCheck(t *testing.T) {
	validator := image.NewValidator(5*1024*1024, []string{"image/jpeg", "image/png"})

	tests := []struct {
		name        string
		contentType string
		magicBytes  []byte
		shouldPass  bool
	}{
		{
			name:        "Valid JPEG magic",
			contentType: "image/jpeg",
			magicBytes:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			shouldPass:  true,
		},
		{
			name:        "Invalid JPEG magic",
			contentType: "image/jpeg",
			magicBytes:  []byte{0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
			shouldPass:  false,
		},
		{
			name:        "Valid PNG magic",
			contentType: "image/png",
			magicBytes:  []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A},
			shouldPass:  true,
		},
		{
			name:        "Fake PNG (wrong magic)",
			contentType: "image/png",
			magicBytes:  []byte{0xFF, 0xD8, 0xFF, 0xE0, 0x00, 0x10},
			shouldPass:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create fake file with specific magic bytes
			data := make([]byte, 1024)
			copy(data, tt.magicBytes)

			fileHeader := createFileHeader("test.jpg", tt.contentType, data)
			err := validator.Validate(fileHeader)

			if tt.shouldPass && err == image.ErrInvalidMagic {
				t.Errorf("Valid magic number should not fail: %v", err)
			}
			if !tt.shouldPass && err != image.ErrInvalidMagic {
				t.Errorf("Invalid magic number should fail with ErrInvalidMagic, got: %v", err)
			}
		})
	}
}

// Test Validator - GetExtension
func TestValidator_GetExtension(t *testing.T) {
	validator := image.NewValidator(5*1024*1024, []string{"image/jpeg"})

	tests := []struct {
		contentType string
		expected    string
	}{
		{"image/jpeg", ".jpg"},
		{"image/png", ".png"},
		{"image/gif", ".gif"},
		{"image/webp", ".webp"},
		{"application/pdf", ""},
	}

	for _, tt := range tests {
		t.Run(tt.contentType, func(t *testing.T) {
			ext := validator.GetExtension(tt.contentType)
			if ext != tt.expected {
				t.Errorf("Expected %s, got %s", tt.expected, ext)
			}
		})
	}
}

// Test Processor - Image Resizing
func TestProcessor_ImageResizing(t *testing.T) {
	processor := image.NewProcessor(2000, 90, 300)

	tests := []struct {
		name           string
		width          int
		height         int
		expectResize   bool
		expectedMaxDim int
	}{
		{"Small image no resize", 800, 600, false, 800},
		{"Large width resize", 3000, 2000, true, 2000},
		{"Large height resize", 2000, 3000, true, 2000},
		{"Exact limit no resize", 2000, 1500, false, 2000},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imgData, err := createTestImage(tt.width, tt.height, "jpeg")
			if err != nil {
				t.Fatalf("Failed to create test image: %v", err)
			}

			fileHeader := createFileHeader("test.jpg", "image/jpeg", imgData)
			processed, err := processor.Process(fileHeader)

			if err != nil {
				t.Fatalf("Failed to process image: %v", err)
			}

			// Check dimensions
			if processed.Width > 2000 || processed.Height > 2000 {
				t.Errorf("Processed image exceeds max dimension: %dx%d", processed.Width, processed.Height)
			}

			// Check thumbnail exists and is within limits
			thumbBounds := processed.Thumbnail.Bounds()
			thumbWidth := thumbBounds.Dx()
			thumbHeight := thumbBounds.Dy()

			if thumbWidth > 300 || thumbHeight > 300 {
				t.Errorf("Thumbnail exceeds max size: %dx%d", thumbWidth, thumbHeight)
			}
		})
	}
}

// Test Processor - Thumbnail Generation
func TestProcessor_ThumbnailGeneration(t *testing.T) {
	processor := image.NewProcessor(2000, 90, 300)

	imgData, _ := createTestImage(1920, 1080, "jpeg")
	fileHeader := createFileHeader("test.jpg", "image/jpeg", imgData)

	processed, err := processor.Process(fileHeader)
	if err != nil {
		t.Fatalf("Failed to process image: %v", err)
	}

	// Verify thumbnail exists
	if processed.Thumbnail == nil {
		t.Fatal("Thumbnail was not generated")
	}

	// Verify thumbnail dimensions
	bounds := processed.Thumbnail.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	if width > 300 || height > 300 {
		t.Errorf("Thumbnail size exceeds limit: %dx%d", width, height)
	}

	// Verify aspect ratio is maintained
	originalRatio := float64(1920) / float64(1080)
	thumbnailRatio := float64(width) / float64(height)

	if abs(originalRatio-thumbnailRatio) > 0.01 {
		t.Errorf("Thumbnail aspect ratio not maintained: original=%.2f, thumbnail=%.2f", originalRatio, thumbnailRatio)
	}
}

func abs(x float64) float64 {
	if x < 0 {
		return -x
	}
	return x
}

// Test Storage - Directory Creation
func TestStorage_DirectoryCreation(t *testing.T) {
	tempDir := t.TempDir()
	storage := image.NewStorage(tempDir, 90)

	// Create a test processed image
	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &image.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Verify directory structure was created
	if _, err := os.Stat(filepath.Dir(result.FullPath)); os.IsNotExist(err) {
		t.Error("Directory structure was not created")
	}

	// Verify files exist
	if _, err := os.Stat(result.FullPath); os.IsNotExist(err) {
		t.Error("Main image file was not created")
	}

	// Verify thumbnail exists
	thumbPath := filepath.Join(filepath.Dir(result.FullPath), filepath.Base(result.ThumbnailPath))
	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Error("Thumbnail file was not created")
	}
}

// Test Storage - Filename Format
func TestStorage_FilenameFormat(t *testing.T) {
	tempDir := t.TempDir()
	storage := image.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))
	processed := &image.ProcessedImage{
		Image:     img,
		Thumbnail: img,
		Width:     100,
		Height:    100,
	}

	result, err := storage.Save(processed, ".jpg")
	if err != nil {
		t.Fatalf("Failed to save image: %v", err)
	}

	// Verify filename format: YYYYMMDDhhmmss_uuid.ext
	filename := result.Filename
	if len(filename) < 24 { // 14 (timestamp) + 1 (_) + 8 (uuid) + 4 (.jpg)
		t.Errorf("Filename format incorrect: %s", filename)
	}

	// Verify extension
	if filepath.Ext(filename) != ".jpg" {
		t.Errorf("Expected .jpg extension, got %s", filepath.Ext(filename))
	}
}

// Test Storage - Cleanup on Thumbnail Failure
func TestStorage_CleanupOnFailure(t *testing.T) {
	tempDir := t.TempDir()
	storage := image.NewStorage(tempDir, 90)

	img := image.NewRGBA(image.Rect(0, 0, 100, 100))

	// Create processed image with nil thumbnail to trigger failure
	processed := &image.ProcessedImage{
		Image:     img,
		Thumbnail: nil, // This should cause thumbnail save to fail
		Width:     100,
		Height:    100,
	}

	_, err := storage.Save(processed, ".jpg")
	if err == nil {
		t.Error("Expected error when thumbnail is nil")
	}
}

// Test Storage - GetImagePath
func TestStorage_GetImagePath(t *testing.T) {
	storage := image.NewStorage("/uploads", 90)

	path := storage.GetImagePath("2026", "03", "test.jpg")
	expected := filepath.Join("/uploads", "2026", "03", "test.jpg")

	if path != expected {
		t.Errorf("Expected %s, got %s", expected, path)
	}
}

// Benchmark tests
func BenchmarkProcessor_Process(b *testing.B) {
	processor := image.NewProcessor(2000, 90, 300)
	imgData, _ := createTestImage(1920, 1080, "jpeg")
	fileHeader := createFileHeader("test.jpg", "image/jpeg", imgData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		processor.Process(fileHeader)
	}
}

func BenchmarkValidator_Validate(b *testing.B) {
	validator := image.NewValidator(5*1024*1024, []string{"image/jpeg", "image/png"})
	imgData, _ := createTestImage(800, 600, "jpeg")
	fileHeader := createFileHeader("test.jpg", "image/jpeg", imgData)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		validator.Validate(fileHeader)
	}
}
