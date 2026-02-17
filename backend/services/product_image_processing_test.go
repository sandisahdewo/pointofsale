package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const tinyPNGDataURL = "data:image/png;base64,iVBORw0KGgoAAAANSUhEUgAAAAEAAAABCAQAAAC1HAwCAAAAC0lEQVR42mP8/x8AAwMCAO8B9foAAAAASUVORK5CYII="

type uploadCall struct {
	objectKey   string
	contentType string
	data        []byte
}

type fakeImageStorage struct {
	returnedURL string
	uploadErr   error
	calls       []uploadCall
}

func (f *fakeImageStorage) UploadImage(_ context.Context, objectKey string, data []byte, contentType string) (string, error) {
	f.calls = append(f.calls, uploadCall{
		objectKey:   objectKey,
		contentType: contentType,
		data:        data,
	})
	if f.uploadErr != nil {
		return "", f.uploadErr
	}
	return f.returnedURL, nil
}

func TestResolveImageURL_NonDataURL_ReturnsOriginal(t *testing.T) {
	svc := &ProductService{}

	got, err := svc.resolveImageURL(" https://example.com/image.jpg ", "products/1/image")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/image.jpg", got)
}

func TestResolveImageURL_Base64Image_UploadsAndReturnsURL(t *testing.T) {
	storage := &fakeImageStorage{returnedURL: "http://localhost:9000/pos-images/products/1/image.png"}
	svc := &ProductService{imageStorage: storage}

	got, err := svc.resolveImageURL(tinyPNGDataURL, "products/1/image")
	require.NoError(t, err)
	assert.Equal(t, storage.returnedURL, got)
	require.Len(t, storage.calls, 1)
	assert.Equal(t, "products/1/image.png", storage.calls[0].objectKey)
	assert.Equal(t, "image/png", storage.calls[0].contentType)
	assert.NotEmpty(t, storage.calls[0].data)
}

func TestResolveImageURL_Base64ImageWithoutStorage_ReturnsError(t *testing.T) {
	svc := &ProductService{}

	_, err := svc.resolveImageURL(tinyPNGDataURL, "products/1/image")
	require.Error(t, err)
	assert.ErrorContains(t, err, "image storage is not configured")
}

func TestResolveImageURL_InvalidBase64_ReturnsError(t *testing.T) {
	storage := &fakeImageStorage{returnedURL: "http://localhost:9000/pos-images/products/1/image.png"}
	svc := &ProductService{imageStorage: storage}

	_, err := svc.resolveImageURL("data:image/png;base64,###", "products/1/image")
	require.Error(t, err)
	assert.ErrorContains(t, err, "decode base64 image")
	assert.Empty(t, storage.calls)
}

func TestResolveImageURL_UploadFailure_ReturnsError(t *testing.T) {
	storage := &fakeImageStorage{uploadErr: errors.New("upload failed")}
	svc := &ProductService{imageStorage: storage}

	_, err := svc.resolveImageURL(tinyPNGDataURL, "products/1/image")
	require.Error(t, err)
	assert.ErrorContains(t, err, "upload image")
	require.Len(t, storage.calls, 1)
}
