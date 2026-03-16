package media

import (
	"bytes"
	"errors"
	"image"
	"image/jpeg"
	"image/png"
	"net/http"
	"strings"

	xdraw "golang.org/x/image/draw"
	_ "golang.org/x/image/webp"
)

var (
	ErrMalformedImage          = errors.New("malformed image")
	ErrImageDimensionsExceeded = errors.New("image dimensions exceeded")
)

const (
	DefaultThumbnailMaxDimension = 512
	thumbnailJPEGQuality         = 82
)

type ValidationLimits struct {
	MaxDimension int
	MaxPixels    int
}

func ValidateImage(content []byte, allowedMIMEs map[string]struct{}, limits ValidationLimits) (string, image.Config, error) {
	mimeType := sniffMimeType(content)
	if _, ok := allowedMIMEs[mimeType]; !ok {
		return "", image.Config{}, http.ErrNotSupported
	}

	cfg, _, err := image.DecodeConfig(bytes.NewReader(content))
	if err != nil {
		return "", image.Config{}, ErrMalformedImage
	}
	if cfg.Width <= 0 || cfg.Height <= 0 {
		return "", image.Config{}, ErrMalformedImage
	}
	if limits.MaxDimension > 0 && (cfg.Width > limits.MaxDimension || cfg.Height > limits.MaxDimension) {
		return "", image.Config{}, ErrImageDimensionsExceeded
	}
	if limits.MaxPixels > 0 {
		pixels := int64(cfg.Width) * int64(cfg.Height)
		if pixels > int64(limits.MaxPixels) {
			return "", image.Config{}, ErrImageDimensionsExceeded
		}
	}

	return mimeType, cfg, nil
}

func GenerateThumbnail(content []byte, maxDimension int) ([]byte, string, error) {
	src, _, err := image.Decode(bytes.NewReader(content))
	if err != nil {
		return nil, "", ErrMalformedImage
	}

	bounds := src.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width <= 0 || height <= 0 {
		return nil, "", ErrMalformedImage
	}

	targetWidth, targetHeight := thumbnailDimensions(width, height, maxDimension)
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	xdraw.ApproxBiLinear.Scale(dst, dst.Bounds(), src, bounds, xdraw.Over, nil)

	var encoded bytes.Buffer
	if imageIsOpaque(src) {
		if err := jpeg.Encode(&encoded, dst, &jpeg.Options{Quality: thumbnailJPEGQuality}); err != nil {
			return nil, "", err
		}
		return encoded.Bytes(), "image/jpeg", nil
	}

	if err := png.Encode(&encoded, dst); err != nil {
		return nil, "", err
	}
	return encoded.Bytes(), "image/png", nil
}

func ThumbnailFilename(originalFilename string, mimeType string) string {
	trimmed := strings.TrimSpace(originalFilename)
	base := trimmed
	if dot := strings.LastIndex(trimmed, "."); dot > 0 {
		base = trimmed[:dot]
	}
	if base == "" {
		base = "upload"
	}

	return base + "-thumb" + extensionForMimeType(mimeType)
}

func sniffMimeType(content []byte) string {
	if len(content) == 0 {
		return "application/octet-stream"
	}
	if len(content) > 512 {
		return http.DetectContentType(content[:512])
	}
	return http.DetectContentType(content)
}

func thumbnailDimensions(width int, height int, maxDimension int) (int, int) {
	if maxDimension <= 0 || (width <= maxDimension && height <= maxDimension) {
		return width, height
	}

	if width >= height {
		targetWidth := maxDimension
		targetHeight := height * maxDimension / width
		if targetHeight == 0 {
			targetHeight = 1
		}
		return targetWidth, targetHeight
	}

	targetHeight := maxDimension
	targetWidth := width * maxDimension / height
	if targetWidth == 0 {
		targetWidth = 1
	}
	return targetWidth, targetHeight
}

func imageIsOpaque(img image.Image) bool {
	type opaque interface {
		Opaque() bool
	}
	if typed, ok := img.(opaque); ok {
		return typed.Opaque()
	}
	return false
}

func extensionForMimeType(mimeType string) string {
	switch mimeType {
	case "image/png":
		return ".png"
	case "image/webp":
		return ".webp"
	default:
		return ".jpg"
	}
}
