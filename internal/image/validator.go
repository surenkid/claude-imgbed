package image

import (
	"bytes"
	"errors"
	"mime/multipart"
)

var (
	ErrInvalidFormat = errors.New("invalid image format")
	ErrFileTooLarge  = errors.New("file too large")
	ErrInvalidMagic  = errors.New("invalid file magic number")
)

// Magic numbers for image formats
var magicNumbers = map[string][]byte{
	"image/jpeg": {0xFF, 0xD8, 0xFF},
	"image/png":  {0x89, 0x50, 0x4E, 0x47},
	"image/gif":  {0x47, 0x49, 0x46},
	"image/webp": {0x52, 0x49, 0x46, 0x46}, // RIFF header
}

type Validator struct {
	maxSize      int64
	allowedTypes map[string]bool
}

func NewValidator(maxSize int64, allowedTypes []string) *Validator {
	allowed := make(map[string]bool)
	for _, t := range allowedTypes {
		allowed[t] = true
	}
	return &Validator{
		maxSize:      maxSize,
		allowedTypes: allowed,
	}
}

func (v *Validator) Validate(file *multipart.FileHeader) error {
	// Check file size
	if file.Size > v.maxSize {
		return ErrFileTooLarge
	}

	// Check content type
	contentType := file.Header.Get("Content-Type")
	if !v.allowedTypes[contentType] {
		return ErrInvalidFormat
	}

	// Open file to check magic number
	f, err := file.Open()
	if err != nil {
		return err
	}
	defer f.Close()

	// Read first 12 bytes for magic number check
	header := make([]byte, 12)
	n, err := f.Read(header)
	if err != nil {
		return err
	}

	// Check magic number
	magic, exists := magicNumbers[contentType]
	if !exists {
		return ErrInvalidFormat
	}

	if n < len(magic) || !bytes.HasPrefix(header, magic) {
		return ErrInvalidMagic
	}

	// Additional check for WebP
	if contentType == "image/webp" {
		// WebP format: RIFF....WEBP
		if n < 12 || !bytes.Equal(header[8:12], []byte("WEBP")) {
			return ErrInvalidMagic
		}
	}

	return nil
}

func (v *Validator) GetExtension(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	default:
		return ""
	}
}
