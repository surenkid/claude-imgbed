package models

import (
	"sync"
	"time"
)

type UploadRecord struct {
	URL         string    `json:"url"`
	Thumbnail   string    `json:"thumbnail"`
	Filename    string    `json:"filename"`
	Size        int64     `json:"size"`
	Width       int       `json:"width"`
	Height      int       `json:"height"`
	UploadedAt  time.Time `json:"uploadedAt"`
}

type RecentUploads struct {
	mu      sync.RWMutex
	uploads []UploadRecord
	maxSize int
}

func NewRecentUploads(maxSize int) *RecentUploads {
	return &RecentUploads{
		uploads: make([]UploadRecord, 0, maxSize),
		maxSize: maxSize,
	}
}

func (r *RecentUploads) Add(record UploadRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Add to the beginning
	r.uploads = append([]UploadRecord{record}, r.uploads...)

	// Keep only maxSize records
	if len(r.uploads) > r.maxSize {
		r.uploads = r.uploads[:r.maxSize]
	}
}

func (r *RecentUploads) GetAll() []UploadRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy
	result := make([]UploadRecord, len(r.uploads))
	copy(result, r.uploads)
	return result
}
