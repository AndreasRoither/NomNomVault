package media

import (
	"bytes"
	"encoding/base64"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"net/http"
	"testing"
)

func TestValidateImageAcceptsCommonFormats(t *testing.T) {
	t.Parallel()

	allowed := map[string]struct{}{
		"image/png":  {},
		"image/jpeg": {},
		"image/webp": {},
	}
	limits := ValidationLimits{MaxDimension: 4096, MaxPixels: 12_000_000}

	cases := []struct {
		name     string
		content  []byte
		expected string
	}{
		{name: "png", content: validPNG(false), expected: "image/png"},
		{name: "jpeg", content: validJPEG(), expected: "image/jpeg"},
		{name: "webp", content: validWebP(), expected: "image/webp"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			mimeType, cfg, err := ValidateImage(tc.content, allowed, limits)
			if err != nil {
				t.Fatalf("expected valid image, got error: %v", err)
			}
			if mimeType != tc.expected {
				t.Fatalf("expected mime %q, got %q", tc.expected, mimeType)
			}
			if cfg.Width <= 0 || cfg.Height <= 0 {
				t.Fatalf("expected positive dimensions, got %dx%d", cfg.Width, cfg.Height)
			}
		})
	}
}

func TestValidateImageRejectsInvalidInputs(t *testing.T) {
	t.Parallel()

	allowed := map[string]struct{}{
		"image/png":  {},
		"image/jpeg": {},
		"image/webp": {},
	}

	if _, _, err := ValidateImage([]byte("not-an-image"), allowed, ValidationLimits{MaxDimension: 4096, MaxPixels: 12_000_000}); err != http.ErrNotSupported {
		t.Fatalf("expected unsupported mime, got %v", err)
	}

	if _, _, err := ValidateImage(corruptPNG(), allowed, ValidationLimits{MaxDimension: 4096, MaxPixels: 12_000_000}); err != ErrMalformedImage {
		t.Fatalf("expected malformed image, got %v", err)
	}

	if _, _, err := ValidateImage(validPNG(false), allowed, ValidationLimits{MaxDimension: 1, MaxPixels: 12_000_000}); err != ErrImageDimensionsExceeded {
		t.Fatalf("expected dimension limit error, got %v", err)
	}

	if _, _, err := ValidateImage(validPNG(false), allowed, ValidationLimits{MaxDimension: 4096, MaxPixels: 1}); err != ErrImageDimensionsExceeded {
		t.Fatalf("expected max pixel error, got %v", err)
	}
}

func TestGenerateThumbnailPreservesFormatRulesAndAspectRatio(t *testing.T) {
	t.Parallel()

	jpegThumbnail, jpegMIME, err := GenerateThumbnail(largeJPEG(), 512)
	if err != nil {
		t.Fatalf("generate jpeg thumbnail: %v", err)
	}
	if jpegMIME != "image/jpeg" {
		t.Fatalf("expected jpeg thumbnail mime, got %q", jpegMIME)
	}
	jpegCfg, _, err := image.DecodeConfig(bytes.NewReader(jpegThumbnail))
	if err != nil {
		t.Fatalf("decode jpeg thumbnail: %v", err)
	}
	if jpegCfg.Width != 512 || jpegCfg.Height != 384 {
		t.Fatalf("expected 512x384 jpeg thumbnail, got %dx%d", jpegCfg.Width, jpegCfg.Height)
	}

	pngThumbnail, pngMIME, err := GenerateThumbnail(largeTransparentPNG(), 512)
	if err != nil {
		t.Fatalf("generate png thumbnail: %v", err)
	}
	if pngMIME != "image/png" {
		t.Fatalf("expected png thumbnail mime, got %q", pngMIME)
	}
	pngCfg, _, err := image.DecodeConfig(bytes.NewReader(pngThumbnail))
	if err != nil {
		t.Fatalf("decode png thumbnail: %v", err)
	}
	if pngCfg.Width != 512 || pngCfg.Height != 384 {
		t.Fatalf("expected 512x384 png thumbnail, got %dx%d", pngCfg.Width, pngCfg.Height)
	}
}

func TestThumbnailFilename(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name     string
		input    string
		mimeType string
		expected string
	}{
		{name: "replace extension", input: "hero.png", mimeType: "image/jpeg", expected: "hero-thumb.jpg"},
		{name: "no extension", input: "hero", mimeType: "image/png", expected: "hero-thumb.png"},
		{name: "trim whitespace", input: "  hero.webp  ", mimeType: "image/png", expected: "hero-thumb.png"},
		{name: "empty name", input: "   ", mimeType: "image/jpeg", expected: "upload-thumb.jpg"},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if actual := ThumbnailFilename(tc.input, tc.mimeType); actual != tc.expected {
				t.Fatalf("expected %q, got %q", tc.expected, actual)
			}
		})
	}
}

func validPNG(transparent bool) []byte {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	alpha := uint8(255)
	if transparent {
		alpha = 120
	}
	img.Set(0, 0, color.RGBA{R: 220, G: 64, B: 64, A: alpha})
	img.Set(1, 0, color.RGBA{R: 64, G: 140, B: 220, A: alpha})
	img.Set(0, 1, color.RGBA{R: 64, G: 220, B: 96, A: alpha})
	img.Set(1, 1, color.RGBA{R: 240, G: 220, B: 64, A: alpha})

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

func validJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 2, 2))
	img.Set(0, 0, color.RGBA{R: 200, G: 70, B: 70, A: 255})
	img.Set(1, 0, color.RGBA{R: 70, G: 120, B: 220, A: 255})
	img.Set(0, 1, color.RGBA{R: 70, G: 210, B: 90, A: 255})
	img.Set(1, 1, color.RGBA{R: 230, G: 210, B: 60, A: 255})

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 90}); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

func largeJPEG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	for y := 0; y < 768; y++ {
		for x := 0; x < 1024; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / 1023),
				G: uint8((y * 255) / 767),
				B: 96,
				A: 255,
			})
		}
	}

	var encoded bytes.Buffer
	if err := jpeg.Encode(&encoded, img, &jpeg.Options{Quality: 90}); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

func largeTransparentPNG() []byte {
	img := image.NewRGBA(image.Rect(0, 0, 1024, 768))
	for y := 0; y < 768; y++ {
		for x := 0; x < 1024; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8((x * 255) / 1023),
				G: uint8((y * 255) / 767),
				B: 128,
				A: 120,
			})
		}
	}

	var encoded bytes.Buffer
	if err := png.Encode(&encoded, img); err != nil {
		panic(err)
	}
	return encoded.Bytes()
}

func corruptPNG() []byte {
	bytes := validPNG(false)
	if len(bytes) > 8 {
		return bytes[:8]
	}
	return bytes
}

func validWebP() []byte {
	encoded := "UklGRkYAAABXRUJQVlA4IDoAAADwAQCdASoCAAIAAgA0JYgCdLoAAwkG+4AA/pwpnW3zlW2e6r0jgkPVYC6Pe4jvoTY9qK+RSLF4AAAA"
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		panic(err)
	}
	return decoded
}
