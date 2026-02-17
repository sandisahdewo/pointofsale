package storage

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOConfig struct {
	Endpoint  string
	AccessKey string
	SecretKey string
	Bucket    string
	UseSSL    bool
	PublicURL string
}

type MinIOImageStorage struct {
	client        *minio.Client
	bucket        string
	publicBaseURL string
}

func NewMinIOImageStorage(cfg MinIOConfig) (*MinIOImageStorage, error) {
	endpoint := strings.TrimSpace(cfg.Endpoint)
	accessKey := strings.TrimSpace(cfg.AccessKey)
	secretKey := strings.TrimSpace(cfg.SecretKey)
	bucket := strings.TrimSpace(cfg.Bucket)
	publicURL := strings.TrimSpace(cfg.PublicURL)

	if endpoint == "" {
		return nil, fmt.Errorf("minio endpoint is required")
	}
	if accessKey == "" {
		return nil, fmt.Errorf("minio access key is required")
	}
	if secretKey == "" {
		return nil, fmt.Errorf("minio secret key is required")
	}
	if bucket == "" {
		return nil, fmt.Errorf("minio bucket is required")
	}
	if publicURL == "" {
		return nil, fmt.Errorf("minio public url is required")
	}

	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("create minio client: %w", err)
	}

	storage := &MinIOImageStorage{
		client:        client,
		bucket:        bucket,
		publicBaseURL: strings.TrimRight(publicURL, "/"),
	}

	if err := storage.ensureBucket(context.Background()); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *MinIOImageStorage) UploadImage(ctx context.Context, objectKey string, data []byte, contentType string) (string, error) {
	key := strings.TrimLeft(strings.TrimSpace(objectKey), "/")
	if key == "" {
		return "", fmt.Errorf("object key is required")
	}
	if len(data) == 0 {
		return "", fmt.Errorf("image payload is empty")
	}

	_, err := s.client.PutObject(ctx, s.bucket, key, bytes.NewReader(data), int64(len(data)), minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("upload object to minio: %w", err)
	}

	return fmt.Sprintf("%s/%s/%s", s.publicBaseURL, s.bucket, key), nil
}

func (s *MinIOImageStorage) ensureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return fmt.Errorf("check minio bucket: %w", err)
	}
	if !exists {
		if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("create minio bucket: %w", err)
		}
	}

	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, s.bucket)
	if err := s.client.SetBucketPolicy(ctx, s.bucket, policy); err != nil {
		return fmt.Errorf("set minio bucket policy: %w", err)
	}

	return nil
}
