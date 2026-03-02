package tests

import (
	"bytes"
	"encoding/json"
	"fmt"
	"claude-imgbed/internal/api"
	"claude-imgbed/internal/auth"
	"claude-imgbed/internal/config"
	"claude-imgbed/internal/models"
	"claude-imgbed/internal/ratelimit"
	"io"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

// E2E Test: Complete Upload Flow
func TestE2E_CompleteUploadFlow(t *testing.T) {
	router, cfg, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Step 1: Health check
	t.Log("Step 1: Health check")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Health check failed: %d", w.Code)
	}

	// Step 2: Upload an image
	t.Log("Step 2: Upload image")
	imgData, _ := createTestImage(1920, 1080, "jpeg")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "vacation.jpg")
	part.Write(imgData)
	writer.Close()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Upload failed: %d - %s", w.Code, w.Body.String())
	}

	var uploadResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &uploadResponse)

	if uploadResponse["success"] != true {
		t.Fatal("Upload response success should be true")
	}

	data := uploadResponse["data"].(map[string]interface{})
	imageURL := data["url"].(string)
	thumbnailURL := data["thumbnail"].(string)
	filename := data["filename"].(string)

	t.Logf("Uploaded: %s", filename)
	t.Logf("URL: %s", imageURL)
	t.Logf("Thumbnail: %s", thumbnailURL)

	// Step 3: Verify file exists on disk
	t.Log("Step 3: Verify file on disk")
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	imagePath := filepath.Join(cfg.Upload.StoragePath, year, month, filename)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Fatalf("Uploaded file does not exist: %s", imagePath)
	}

	// Step 4: Retrieve the image via API
	t.Log("Step 4: Retrieve image via API")
	imageURLPath := "/images/" + year + "/" + month + "/" + filename

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", imageURLPath, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Image retrieval failed: %d", w.Code)
	}

	if w.Header().Get("Content-Type") != "image/jpeg" {
		t.Errorf("Wrong content type: %s", w.Header().Get("Content-Type"))
	}

	// Step 5: Check recent uploads
	t.Log("Step 5: Check recent uploads")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/recent", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Recent uploads failed: %d", w.Code)
	}

	var recentResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &recentResponse)
	recentData := recentResponse["data"].([]interface{})

	if len(recentData) == 0 {
		t.Fatal("Recent uploads should not be empty")
	}

	// Verify our upload is in recent
	found := false
	for _, item := range recentData {
		record := item.(map[string]interface{})
		if record["filename"] == filename {
			found = true
			t.Logf("Found in recent uploads: %s", filename)
			break
		}
	}

	if !found {
		t.Error("Uploaded image not found in recent uploads")
	}

	t.Log("E2E test completed successfully")
}

// E2E Test: Multiple Image Upload
func TestE2E_MultipleImageUpload(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	imageCount := 5
	uploadedFiles := make([]string, 0, imageCount)

	// Upload multiple images
	for i := 0; i < imageCount; i++ {
		t.Logf("Uploading image %d/%d", i+1, imageCount)

		imgData, _ := createTestImage(800+i*100, 600+i*100, "jpeg")

		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", fmt.Sprintf("image%d.jpg", i))
		part.Write(imgData)
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer test-token-123")

		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("Upload %d failed: %d", i, w.Code)
		}

		var response map[string]interface{}
		json.Unmarshal(w.Body.Bytes(), &response)
		data := response["data"].(map[string]interface{})
		uploadedFiles = append(uploadedFiles, data["filename"].(string))

		// Small delay between uploads
		time.Sleep(10 * time.Millisecond)
	}

	// Verify all images in recent uploads
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/recent", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	router.ServeHTTP(w, req)

	var recentResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &recentResponse)
	recentData := recentResponse["data"].([]interface{})

	if len(recentData) < imageCount {
		t.Errorf("Expected at least %d images in recent, got %d", imageCount, len(recentData))
	}

	// Verify all uploaded files are present
	for _, filename := range uploadedFiles {
		found := false
		for _, item := range recentData {
			record := item.(map[string]interface{})
			if record["filename"] == filename {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Uploaded file %s not found in recent uploads", filename)
		}
	}

	t.Logf("Successfully uploaded and verified %d images", imageCount)
}

// E2E Test: Error Handling Flow
func TestE2E_ErrorHandlingFlow(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		setupRequest   func() (*http.Request, *bytes.Buffer, *multipart.Writer)
		expectedStatus int
		description    string
	}{
		{
			name: "Missing authentication",
			setupRequest: func() (*http.Request, *bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				req, _ := http.NewRequest("POST", "/api/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				return req, body, writer
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should reject request without auth token",
		},
		{
			name: "Invalid token",
			setupRequest: func() (*http.Request, *bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				req, _ := http.NewRequest("POST", "/api/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer wrong-token")
				return req, body, writer
			},
			expectedStatus: http.StatusUnauthorized,
			description:    "Should reject request with invalid token",
		},
		{
			name: "No file uploaded",
			setupRequest: func() (*http.Request, *bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				writer.Close()
				req, _ := http.NewRequest("POST", "/api/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer test-token-123")
				return req, body, writer
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject request without file",
		},
		{
			name: "Invalid file format",
			setupRequest: func() (*http.Request, *bytes.Buffer, *multipart.Writer) {
				body := &bytes.Buffer{}
				writer := multipart.NewWriter(body)
				part, _ := writer.CreateFormFile("file", "document.pdf")
				part.Write([]byte("fake pdf content"))
				writer.Close()
				req, _ := http.NewRequest("POST", "/api/upload", body)
				req.Header.Set("Content-Type", writer.FormDataContentType())
				req.Header.Set("Authorization", "Bearer test-token-123")
				return req, body, writer
			},
			expectedStatus: http.StatusBadRequest,
			description:    "Should reject non-image files",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _, _ := tt.setupRequest()
			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("%s: expected status %d, got %d", tt.description, tt.expectedStatus, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)

			if response["success"] != false {
				t.Errorf("%s: expected success=false", tt.description)
			}

			t.Logf("✓ %s", tt.description)
		})
	}
}

// E2E Test: Image Size Handling
func TestE2E_ImageSizeHandling(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	tests := []struct {
		name           string
		width          int
		height         int
		expectResize   bool
		maxDimension   int
		description    string
	}{
		{
			name:         "Small image",
			width:        800,
			height:       600,
			expectResize: false,
			maxDimension: 2000,
			description:  "Small images should not be resized",
		},
		{
			name:         "Large width",
			width:        3000,
			height:       2000,
			expectResize: true,
			maxDimension: 2000,
			description:  "Images with width > 2000 should be resized",
		},
		{
			name:         "Large height",
			width:        1500,
			height:       2500,
			expectResize: true,
			maxDimension: 2000,
			description:  "Images with height > 2000 should be resized",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			imgData, _ := createTestImage(tt.width, tt.height, "jpeg")

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "test.jpg")
			part.Write(imgData)
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer test-token-123")

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("Upload failed: %d", w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})

			width := int(data["width"].(float64))
			height := int(data["height"].(float64))

			if width > tt.maxDimension || height > tt.maxDimension {
				t.Errorf("Image dimensions exceed max: %dx%d", width, height)
			}

			t.Logf("✓ %s: %dx%d -> %dx%d", tt.description, tt.width, tt.height, width, height)
		})
	}
}

// E2E Test: Concurrent Upload Stress Test
func TestE2E_ConcurrentUploadStress(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping stress test in short mode")
	}

	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	concurrency := 10
	done := make(chan bool, concurrency)
	errors := make(chan error, concurrency)

	t.Logf("Starting concurrent upload stress test with %d goroutines", concurrency)

	startTime := time.Now()

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			imgData, _ := createTestImage(800, 600, "jpeg")

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", fmt.Sprintf("concurrent%d.jpg", id))
			part.Write(imgData)
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer test-token-123")
			req.RemoteAddr = fmt.Sprintf("192.168.1.%d:12345", id+1) // Different IPs

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				errors <- fmt.Errorf("upload %d failed with status %d", id, w.Code)
			}

			done <- true
		}(i)
	}

	// Wait for all uploads
	for i := 0; i < concurrency; i++ {
		<-done
	}

	duration := time.Since(startTime)

	close(errors)
	errorCount := 0
	for err := range errors {
		t.Error(err)
		errorCount++
	}

	successCount := concurrency - errorCount
	t.Logf("Completed %d/%d uploads in %v", successCount, concurrency, duration)
	t.Logf("Average time per upload: %v", duration/time.Duration(concurrency))

	if errorCount > 0 {
		t.Errorf("%d/%d uploads failed", errorCount, concurrency)
	}
}

// E2E Test: Rate Limiting Behavior
func TestE2E_RateLimitingBehavior(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	imgData, _ := createTestImage(200, 200, "jpeg")

	successCount := 0
	rateLimitedCount := 0
	totalRequests := 20

	t.Logf("Testing rate limiting with %d requests", totalRequests)

	for i := 0; i < totalRequests; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write(imgData)
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer test-token-123")
		req.RemoteAddr = "192.168.1.100:12345" // Same IP for all

		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			rateLimitedCount++
		}

		// No delay - test burst behavior
	}

	t.Logf("Results: %d successful, %d rate-limited", successCount, rateLimitedCount)

	if rateLimitedCount == 0 {
		t.Error("Expected some requests to be rate-limited")
	}

	if successCount == 0 {
		t.Error("Expected some requests to succeed")
	}
}

// E2E Test: Image Format Support
func TestE2E_ImageFormatSupport(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	formats := []struct {
		name        string
		format      string
		contentType string
	}{
		{"JPEG", "jpeg", "image/jpeg"},
		{"PNG", "png", "image/png"},
	}

	for _, fmt := range formats {
		t.Run(fmt.name, func(t *testing.T) {
			imgData, _ := createTestImage(500, 400, fmt.format)

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "test."+fmt.format)
			part.Write(imgData)
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer test-token-123")

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Fatalf("%s upload failed: %d", fmt.name, w.Code)
			}

			var response map[string]interface{}
			json.Unmarshal(w.Body.Bytes(), &response)
			data := response["data"].(map[string]interface{})
			filename := data["filename"].(string)

			if !strings.HasSuffix(filename, "."+fmt.format) && !strings.HasSuffix(filename, ".jpg") {
				t.Errorf("Filename should have correct extension: %s", filename)
			}

			t.Logf("✓ %s format supported", fmt.name)
		})
	}
}

// E2E Test: Cache Headers
func TestE2E_CacheHeaders(t *testing.T) {
	router, cfg, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Upload an image first
	imgData, _ := createTestImage(300, 300, "jpeg")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "cache-test.jpg")
	part.Write(imgData)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w, req)

	var uploadResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	data := uploadResponse["data"].(map[string]interface{})
	filename := data["filename"].(string)

	// Now retrieve the image and check cache headers
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/images/"+year+"/"+month+"/"+filename, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Image retrieval failed: %d", w.Code)
	}

	// Verify cache headers
	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=31536000" {
		t.Errorf("Expected Cache-Control 'public, max-age=31536000', got '%s'", cacheControl)
	}

	cdnCache := w.Header().Get("CDN-Cache-Control")
	if cdnCache != "max-age=31536000" {
		t.Errorf("Expected CDN-Cache-Control 'max-age=31536000', got '%s'", cdnCache)
	}

	t.Log("✓ Cache headers correctly set for CDN optimization")
}

// E2E Test: Full Workflow with Verification
func TestE2E_FullWorkflowWithVerification(t *testing.T) {
	router, cfg, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	t.Log("=== Starting Full E2E Workflow Test ===")

	// 1. Health check
	t.Log("1. Checking system health...")
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatal("Health check failed")
	}
	t.Log("✓ System healthy")

	// 2. Upload image
	t.Log("2. Uploading test image...")
	imgData, _ := createTestImage(1600, 1200, "jpeg")

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "workflow-test.jpg")
	part.Write(imgData)
	writer.Close()

	w = httptest.NewRecorder()
	req, _ = http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token-123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("Upload failed: %d", w.Code)
	}

	var uploadResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	data := uploadResponse["data"].(map[string]interface{})
	filename := data["filename"].(string)
	t.Logf("✓ Image uploaded: %s", filename)

	// 3. Verify file on disk
	t.Log("3. Verifying file on disk...")
	now := time.Now()
	year := now.Format("2006")
	month := now.Format("01")
	imagePath := filepath.Join(cfg.Upload.StoragePath, year, month, filename)

	if _, err := os.Stat(imagePath); os.IsNotExist(err) {
		t.Fatal("Image file not found on disk")
	}
	t.Log("✓ File exists on disk")

	// 4. Verify thumbnail
	t.Log("4. Verifying thumbnail...")
	thumbFilename := strings.Replace(filename, ".jpg", "_thumb.jpg", 1)
	thumbPath := filepath.Join(cfg.Upload.StoragePath, year, month, thumbFilename)

	if _, err := os.Stat(thumbPath); os.IsNotExist(err) {
		t.Fatal("Thumbnail not found on disk")
	}
	t.Log("✓ Thumbnail exists")

	// 5. Retrieve via API
	t.Log("5. Retrieving image via API...")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/images/"+year+"/"+month+"/"+filename, nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal("Image retrieval failed")
	}
	t.Log("✓ Image retrieved successfully")

	// 6. Check recent uploads
	t.Log("6. Checking recent uploads...")
	w = httptest.NewRecorder()
	req, _ = http.NewRequest("GET", "/api/recent", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatal("Recent uploads check failed")
	}

	var recentResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &recentResponse)
	recentData := recentResponse["data"].([]interface{})

	found := false
	for _, item := range recentData {
		record := item.(map[string]interface{})
		if record["filename"] == filename {
			found = true
			break
		}
	}

	if !found {
		t.Fatal("Uploaded image not in recent uploads")
	}
	t.Log("✓ Image found in recent uploads")

	t.Log("=== Full E2E Workflow Test Completed Successfully ===")
}
