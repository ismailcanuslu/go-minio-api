package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ObjectInfo struct {
	Key          string    `json:"key"`
	Size         int64     `json:"size"`
	ETag         string    `json:"etag"`
	LastModified time.Time `json:"lastModified"`
	ContentType  string    `json:"contentType,omitempty"`
}

type UploadInfo struct {
	ETag string `json:"etag"`
	Size int64  `json:"size"`
}

type MinIOService struct {
	client *minio.Client
	bucket string
}

func NewMinIOService(ctx context.Context, endpoint, accessKey, secretKey string, useSSL bool, bucket string) (*MinIOService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client init failed: %w", err)
	}

	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("bucket check failed: %w", err)
	}

	if !exists {
		if err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("bucket create failed: %w", err)
		}
	}

	return &MinIOService{client: client, bucket: bucket}, nil
}

func (s *MinIOService) Bucket() string {
	return s.bucket
}

func (s *MinIOService) ListObjects(ctx context.Context, prefix string, recursive bool) ([]ObjectInfo, error) {
	results := make([]ObjectInfo, 0)
	for object := range s.client.ListObjects(ctx, s.bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: recursive}) {
		if object.Err != nil {
			return nil, fmt.Errorf("list objects failed: %w", object.Err)
		}
		results = append(results, ObjectInfo{
			Key:          object.Key,
			Size:         object.Size,
			ETag:         object.ETag,
			LastModified: object.LastModified,
		})
	}
	return results, nil
}

func (s *MinIOService) PutObject(ctx context.Context, key, contentType string, body io.Reader, size int64) (UploadInfo, error) {
	info, err := s.client.PutObject(ctx, s.bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return UploadInfo{}, fmt.Errorf("put object failed: %w", err)
	}

	return UploadInfo{ETag: info.ETag, Size: info.Size}, nil
}

func (s *MinIOService) GetObject(ctx context.Context, key string) (*minio.Object, ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, ObjectInfo{}, fmt.Errorf("get object failed: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		obj.Close()
		return nil, ObjectInfo{}, fmt.Errorf("object not found or inaccessible: %w", err)
	}

	return obj, ObjectInfo{
		Key:          key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		LastModified: stat.LastModified,
		ContentType:  stat.ContentType,
	}, nil
}

func (s *MinIOService) DeleteObject(ctx context.Context, key string) error {
	if err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object failed: %w", err)
	}
	return nil
}

func (s *MinIOService) StatObject(ctx context.Context, key string) (ObjectInfo, error) {
	stat, err := s.client.StatObject(ctx, s.bucket, key, minio.StatObjectOptions{})
	if err != nil {
		return ObjectInfo{}, fmt.Errorf("stat object failed: %w", err)
	}

	return ObjectInfo{
		Key:          key,
		Size:         stat.Size,
		ETag:         stat.ETag,
		LastModified: stat.LastModified,
		ContentType:  stat.ContentType,
	}, nil
}

func (s *MinIOService) PresignGetObject(ctx context.Context, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return url.String(), nil
}
