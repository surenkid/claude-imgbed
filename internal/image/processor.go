package image

import (
	"image"
	"mime/multipart"

	"github.com/disintegration/imaging"
)

type Processor struct {
	maxDimension  int
	quality       int
	thumbnailSize int
}

func NewProcessor(maxDimension, quality, thumbnailSize int) *Processor {
	return &Processor{
		maxDimension:  maxDimension,
		quality:       quality,
		thumbnailSize: thumbnailSize,
	}
}

type ProcessedImage struct {
	Image     image.Image
	Thumbnail image.Image
	Width     int
	Height    int
}

func (p *Processor) Process(file *multipart.FileHeader) (*ProcessedImage, error) {
	// Open the uploaded file
	src, err := file.Open()
	if err != nil {
		return nil, err
	}
	defer src.Close()

	// Decode image
	img, err := imaging.Decode(src)
	if err != nil {
		return nil, err
	}

	// Get original dimensions
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Resize if needed (if either dimension > maxDimension)
	if width > p.maxDimension || height > p.maxDimension {
		if width > height {
			img = imaging.Resize(img, p.maxDimension, 0, imaging.Lanczos)
		} else {
			img = imaging.Resize(img, 0, p.maxDimension, imaging.Lanczos)
		}
		bounds = img.Bounds()
		width = bounds.Dx()
		height = bounds.Dy()
	}

	// Generate thumbnail (fit within thumbnailSize x thumbnailSize)
	thumbnail := imaging.Fit(img, p.thumbnailSize, p.thumbnailSize, imaging.Lanczos)

	return &ProcessedImage{
		Image:     img,
		Thumbnail: thumbnail,
		Width:     width,
		Height:    height,
	}, nil
}
