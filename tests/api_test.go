package tests

import (
	"bytes"
	"encoding/json"
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
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func setupTestRouter() (*gin.Engine, *config.Config, string) {
	gin.SetMode(gin.TestMode)

	// Create temp directory for uploads
	tempDir, _ := os.MkdirTemp("", "imgbed-test-*")

	cfg := &config.Config{
		Server: config.ServerConfig{
			Port: "8080",
			Host: "0.0.0.0",
		},
		Auth: config.AuthConfig{
			Token: "test-token-123",
		},
		Upload: config.UploadConfig{
			MaxSize:      5 * 1024 * 1024, // 5MB
			AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
			StoragePath:  tempDir,
		},
		Image: config.ImageConfig{
			MaxDimension:  2000,
			Quality:       90,
			ThumbnailSize: 300,
		},
		RateLimit: config.RateLimitConfig{
			RequestsPerMinute: 10,
			Burst:             5,
		},
		Cache: config.CacheConfig{
			RecentUploadsSize: 100,
		},
	}

	recentUploads := models.NewRecentUploads(cfg.Cache.RecentUploadsSize)
	router := api.SetupRouter(cfg, recentUploads)

	return router, cfg, tempDir
}

// Test Health Check Endpoint
func TestAPI_HealthCheck(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/health", nil)
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["status"] != "ok" {
		t.Errorf("Expected status 'ok', got %v", response["status"])
	}

	if _, exists := response["timestamp"]; !exists {
		t.Error("Response should include timestamp")
	}
}

// Test Upload - Missing Authentication
func TestAPI_Upload_MissingAuth(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	// No Authorization header

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Test Upload - Invalid Token
func TestAPI_Upload_InvalidToken(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer wrong-token")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Test Upload - No File
func TestAPI_Upload_NoFile(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != false {
		t.Error("Expected success to be false")
	}
}

// Test Upload - File Too Large
func TestAPI_Upload_FileTooLarge(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Create a file larger than 5MB
	largeData := make([]byte, 6*1024*1024)

	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "large.jpg")
	part.Write(largeData)
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// Test Upload - Invalid Format
func TestAPI_Upload_InvalidFormat(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Create a text file pretending to be an image
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)
	part, _ := writer.CreateFormFile("file", "fake.jpg")
	part.Write([]byte("This is not an image"))
	writer.Close()

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected status 400, got %d", w.Code)
	}
}

// Test Upload - Successful Upload
func TestAPI_Upload_Success(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Create a valid JPEG image
	imgData, _ := createTestImage(800, 600, "jpeg")

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
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("Expected success to be true")
	}

	data := response["data"].(map[string]interface{})
	if data["url"] == nil {
		t.Error("Response should include URL")
	}
	if data["thumbnail"] == nil {
		t.Error("Response should include thumbnail URL")
	}
	if data["filename"] == nil {
		t.Error("Response should include filename")
	}
	if data["width"] == nil || data["height"] == nil {
		t.Error("Response should include dimensions")
	}
}

// Test Upload - Multiple Formats
func TestAPI_Upload_MultipleFormats(t *testing.T) {
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
			router, _, tempDir := setupTestRouter()
			defer os.RemoveAll(tempDir)

			imgData, _ := createTestImage(400, 300, fmt.format)

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
				t.Errorf("Expected status 200 for %s, got %d", fmt.name, w.Code)
			}
		})
	}
}

// Test Recent Uploads - Unauthorized
func TestAPI_RecentUploads_Unauthorized(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/recent", nil)
	// No Authorization header

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Errorf("Expected status 401, got %d", w.Code)
	}
}

// Test Recent Uploads - Success
func TestAPI_RecentUploads_Success(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/recent", nil)
	req.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	if response["success"] != true {
		t.Error("Expected success to be true")
	}

	if response["data"] == nil {
		t.Error("Response should include data array")
	}
}

// Test Serve Image - Not Found
func TestAPI_ServeImage_NotFound(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/images/2026/03/nonexistent.jpg", nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Errorf("Expected status 404, got %d", w.Code)
	}
}

// Test Serve Image - Success
func TestAPI_ServeImage_Success(t *testing.T) {
	router, cfg, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Create a test image file
	year := "2026"
	month := "03"
	filename := "test.jpg"

	dirPath := filepath.Join(cfg.Upload.StoragePath, year, month)
	os.MkdirAll(dirPath, 0755)

	imgData, _ := createTestImage(100, 100, "jpeg")
	imgPath := filepath.Join(dirPath, filename)
	os.WriteFile(imgPath, imgData, 0644)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/images/"+year+"/"+month+"/"+filename, nil)

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Check cache headers
	cacheControl := w.Header().Get("Cache-Control")
	if cacheControl != "public, max-age=31536000" {
		t.Errorf("Expected Cache-Control header, got %s", cacheControl)
	}

	cdnCache := w.Header().Get("CDN-Cache-Control")
	if cdnCache != "max-age=31536000" {
		t.Errorf("Expected CDN-Cache-Control header, got %s", cdnCache)
	}

	contentType := w.Header().Get("Content-Type")
	if contentType != "image/jpeg" {
		t.Errorf("Expected Content-Type image/jpeg, got %s", contentType)
	}
}

// Test Rate Limiting
func TestAPI_RateLimit(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	imgData, _ := createTestImage(100, 100, "jpeg")

	// Make multiple requests quickly
	successCount := 0
	rateLimitCount := 0

	for i := 0; i < 15; i++ {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, _ := writer.CreateFormFile("file", "test.jpg")
		part.Write(imgData)
		writer.Close()

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/api/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())
		req.Header.Set("Authorization", "Bearer test-token-123")
		req.RemoteAddr = "192.168.1.1:12345" // Same IP for all requests

		router.ServeHTTP(w, req)

		if w.Code == http.StatusOK {
			successCount++
		} else if w.Code == http.StatusTooManyRequests {
			rateLimitCount++
		}
	}

	// Should have some rate limited requests
	if rateLimitCount == 0 {
		t.Error("Expected some requests to be rate limited")
	}

	t.Logf("Success: %d, Rate Limited: %d", successCount, rateLimitCount)
}

// Test Concurrent Uploads
func TestAPI_ConcurrentUploads(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	concurrency := 5
	done := make(chan bool, concurrency)

	for i := 0; i < concurrency; i++ {
		go func(id int) {
			imgData, _ := createTestImage(200, 200, "jpeg")

			body := &bytes.Buffer{}
			writer := multipart.NewWriter(body)
			part, _ := writer.CreateFormFile("file", "test.jpg")
			part.Write(imgData)
			writer.Close()

			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/upload", body)
			req.Header.Set("Content-Type", writer.FormDataContentType())
			req.Header.Set("Authorization", "Bearer test-token-123")
			req.RemoteAddr = "192.168.1." + string(rune(id+1)) + ":12345" // Different IPs

			router.ServeHTTP(w, req)

			if w.Code != http.StatusOK {
				t.Errorf("Concurrent upload %d failed with status %d", id, w.Code)
			}

			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < concurrency; i++ {
		<-done
	}
}

// Test Authentication Middleware
func TestAuth_ValidateToken(t *testing.T) {
	authenticator := auth.NewAuthenticator("secret-token")

	tests := []struct {
		name     string
		header   string
		expected bool
	}{
		{"Valid token", "Bearer secret-token", true},
		{"Invalid token", "Bearer wrong-token", false},
		{"Missing Bearer", "secret-token", false},
		{"Empty header", "", false},
		{"Only Bearer", "Bearer", false},
		{"Extra spaces", "Bearer  secret-token", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := authenticator.ValidateToken(tt.header)
			if result != tt.expected {
				t.Errorf("Expected %v, got %v for header: %s", tt.expected, result, tt.header)
			}
		})
	}
}

// Test Rate Limiter
func TestRateLimit_IPLimiter(t *testing.T) {
	limiter := ratelimit.NewIPRateLimiter(10, 5) // 10 per minute, burst 5

	ip := "192.168.1.1"

	// First burst should succeed
	for i := 0; i < 5; i++ {
		l := limiter.GetLimiter(ip)
		if !l.Allow() {
			t.Errorf("Request %d should be allowed in burst", i+1)
		}
	}

	// Next request should be rate limited (no time passed)
	l := limiter.GetLimiter(ip)
	if l.Allow() {
		t.Error("Request should be rate limited after burst")
	}

	// Different IP should have its own limit
	ip2 := "192.168.1.2"
	l2 := limiter.GetLimiter(ip2)
	if !l2.Allow() {
		t.Error("Different IP should have its own rate limit")
	}
}

// Test Upload and Retrieve Flow
func TestAPI_UploadAndRetrieveFlow(t *testing.T) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	// Step 1: Upload an image
	imgData, _ := createTestImage(500, 400, "jpeg")

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
		t.Fatalf("Upload failed with status %d", w.Code)
	}

	var uploadResponse map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &uploadResponse)
	data := uploadResponse["data"].(map[string]interface{})

	// Step 2: Check recent uploads
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/api/recent", nil)
	req2.Header.Set("Authorization", "Bearer test-token-123")

	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("Recent uploads failed with status %d", w2.Code)
	}

	var recentResponse map[string]interface{}
	json.Unmarshal(w2.Body.Bytes(), &recentResponse)
	recentData := recentResponse["data"].([]interface{})

	if len(recentData) == 0 {
		t.Error("Recent uploads should contain the uploaded image")
	}

	// Verify the uploaded image is in recent uploads
	found := false
	for _, item := range recentData {
		record := item.(map[string]interface{})
		if record["filename"] == data["filename"] {
			found = true
			break
		}
	}

	if !found {
		t.Error("Uploaded image not found in recent uploads")
	}
}

// Benchmark API Upload
func BenchmarkAPI_Upload(b *testing.B) {
	router, _, tempDir := setupTestRouter()
	defer os.RemoveAll(tempDir)

	imgData, _ := createTestImage(800, 600, "jpeg")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
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
	}
}
