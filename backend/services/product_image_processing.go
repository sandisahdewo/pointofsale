package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"path"
	"strings"
)

type ImageStorage interface {
	UploadImage(ctx context.Context, objectKey string, data []byte, contentType string) (string, error)
}

type decodedImagePayload struct {
	contentType string
	extension   string
	data        []byte
}

func (s *ProductService) resolveImageURL(rawValue string, objectKey string) (string, error) {
	trimmed := strings.TrimSpace(rawValue)
	if trimmed == "" {
		return "", nil
	}

	payload, isDataURL, err := parseImageDataURL(trimmed)
	if err != nil {
		return "", err
	}
	if !isDataURL {
		return trimmed, nil
	}
	if s.imageStorage == nil {
		return "", fmt.Errorf("image storage is not configured")
	}

	key := appendExtension(objectKey, payload.extension)
	uploadedURL, err := s.imageStorage.UploadImage(context.Background(), key, payload.data, payload.contentType)
	if err != nil {
		return "", fmt.Errorf("upload image: %w", err)
	}
	return uploadedURL, nil
}

func parseImageDataURL(value string) (*decodedImagePayload, bool, error) {
	lower := strings.ToLower(value)
	if !strings.HasPrefix(lower, "data:image/") {
		return nil, false, nil
	}

	commaIdx := strings.Index(value, ",")
	if commaIdx < 0 {
		return nil, true, fmt.Errorf("invalid image data url")
	}

	meta := value[len("data:"):commaIdx]
	payload := value[commaIdx+1:]
	if !strings.HasSuffix(strings.ToLower(meta), ";base64") {
		return nil, true, fmt.Errorf("image data url must use base64 encoding")
	}

	contentType := strings.TrimSpace(meta[:len(meta)-len(";base64")])
	if !strings.HasPrefix(strings.ToLower(contentType), "image/") {
		return nil, true, fmt.Errorf("image data url must use image mime type")
	}
	extension, err := imageExtensionFromContentType(contentType)
	if err != nil {
		return nil, true, err
	}

	decoded, err := base64.StdEncoding.DecodeString(payload)
	if err != nil {
		decoded, err = base64.RawStdEncoding.DecodeString(payload)
		if err != nil {
			return nil, true, fmt.Errorf("decode base64 image: %w", err)
		}
	}
	if len(decoded) == 0 {
		return nil, true, fmt.Errorf("image payload is empty")
	}

	return &decodedImagePayload{
		contentType: contentType,
		extension:   extension,
		data:        decoded,
	}, true, nil
}

func imageExtensionFromContentType(contentType string) (string, error) {
	switch strings.ToLower(strings.TrimSpace(contentType)) {
	case "image/jpeg":
		return "jpg", nil
	case "image/png":
		return "png", nil
	case "image/webp":
		return "webp", nil
	case "image/gif":
		return "gif", nil
	case "image/bmp":
		return "bmp", nil
	case "image/tiff":
		return "tiff", nil
	case "image/svg+xml":
		return "svg", nil
	case "image/avif":
		return "avif", nil
	default:
		return "", fmt.Errorf("unsupported image content type: %s", contentType)
	}
}

func appendExtension(objectKey, extension string) string {
	trimmed := strings.TrimSpace(objectKey)
	if trimmed == "" {
		return trimmed
	}
	if extension == "" {
		return trimmed
	}

	base := path.Base(trimmed)
	if strings.Contains(base, ".") {
		return trimmed
	}

	return trimmed + "." + extension
}
