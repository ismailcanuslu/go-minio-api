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

type BucketInfo struct {
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
}

type MinIOService struct {
	client *minio.Client
}

func NewMinIOService(ctx context.Context, endpoint, accessKey, secretKey string, useSSL bool) (*MinIOService, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client init failed: %w", err)
	}

	// Ping MinIO with a lightweight call to verify connectivity.
	if _, err = client.ListBuckets(ctx); err != nil {
		return nil, fmt.Errorf("minio connectivity check failed: %w", err)
	}

	return &MinIOService{client: client}, nil
}

// EnsureBucket creates the bucket if it does not exist.
func (s *MinIOService) EnsureBucket(ctx context.Context, bucket string) error {
	exists, err := s.client.BucketExists(ctx, bucket)
	if err != nil {
		return fmt.Errorf("bucket check failed: %w", err)
	}
	if !exists {
		if err = s.client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return fmt.Errorf("bucket create failed: %w", err)
		}
	}
	return nil
}

// DeleteBucketWithObjects removes all objects then deletes the bucket.
func (s *MinIOService) DeleteBucketWithObjects(ctx context.Context, bucket string) error {
	objectCh := s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Recursive: true})
	for obj := range objectCh {
		if obj.Err != nil {
			return fmt.Errorf("list for delete failed: %w", obj.Err)
		}
		if err := s.client.RemoveObject(ctx, bucket, obj.Key, minio.RemoveObjectOptions{}); err != nil {
			return fmt.Errorf("remove object %s failed: %w", obj.Key, err)
		}
	}
	if err := s.client.RemoveBucket(ctx, bucket); err != nil {
		return fmt.Errorf("remove bucket failed: %w", err)
	}
	return nil
}

// ListBuckets returns all buckets visible to the configured credentials.
func (s *MinIOService) ListBuckets(ctx context.Context) ([]BucketInfo, error) {
	buckets, err := s.client.ListBuckets(ctx)
	if err != nil {
		return nil, fmt.Errorf("list buckets failed: %w", err)
	}
	result := make([]BucketInfo, 0, len(buckets))
	for _, b := range buckets {
		result = append(result, BucketInfo{Name: b.Name, CreatedAt: b.CreationDate})
	}
	return result, nil
}

func (s *MinIOService) ListObjects(ctx context.Context, bucket, prefix string, recursive bool) ([]ObjectInfo, error) {
	results := make([]ObjectInfo, 0)
	for object := range s.client.ListObjects(ctx, bucket, minio.ListObjectsOptions{Prefix: prefix, Recursive: recursive}) {
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

func (s *MinIOService) PutObject(ctx context.Context, bucket, key, contentType string, body io.Reader, size int64) (UploadInfo, error) {
	info, err := s.client.PutObject(ctx, bucket, key, body, size, minio.PutObjectOptions{ContentType: contentType})
	if err != nil {
		return UploadInfo{}, fmt.Errorf("put object failed: %w", err)
	}
	return UploadInfo{ETag: info.ETag, Size: info.Size}, nil
}

func (s *MinIOService) GetObject(ctx context.Context, bucket, key string) (*minio.Object, ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, bucket, key, minio.GetObjectOptions{})
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

func (s *MinIOService) DeleteObject(ctx context.Context, bucket, key string) error {
	if err := s.client.RemoveObject(ctx, bucket, key, minio.RemoveObjectOptions{}); err != nil {
		return fmt.Errorf("delete object failed: %w", err)
	}
	return nil
}

func (s *MinIOService) StatObject(ctx context.Context, bucket, key string) (ObjectInfo, error) {
	stat, err := s.client.StatObject(ctx, bucket, key, minio.StatObjectOptions{})
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

func (s *MinIOService) PresignGetObject(ctx context.Context, bucket, key string, expiry time.Duration) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, bucket, key, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("presign failed: %w", err)
	}
	return url.String(), nil
}
