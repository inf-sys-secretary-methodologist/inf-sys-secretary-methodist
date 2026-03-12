// Package storage provides S3/MinIO storage client implementation.
package storage

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
)

// S3Client provides methods for interacting with S3/MinIO storage
type S3Client struct {
	client         *minio.Client
	publicClient   *minio.Client // Client configured with public endpoint for presigned URLs
	bucketName     string
	maxSize        int64
	publicEndpoint string // External endpoint for presigned URLs
	useSSL         bool
	publicUseSSL   bool // SSL for public URLs (via reverse proxy)
}

// FileInfo contains metadata about uploaded file
type FileInfo struct {
	Key         string    `json:"key"`
	Size        int64     `json:"size"`
	ContentType string    `json:"content_type"`
	ETag        string    `json:"etag"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

// NewS3Client creates a new S3/MinIO client
func NewS3Client(cfg config.S3Config) (*S3Client, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create S3 client: %w", err)
	}

	// Create a second client for public endpoint (used for presigned URLs)
	// This is needed because presigned URLs include the host in the signature
	var publicClient *minio.Client
	if cfg.PublicEndpoint != "" && cfg.PublicEndpoint != cfg.Endpoint {
		publicClient, err = minio.New(cfg.PublicEndpoint, &minio.Options{
			Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
			Secure: cfg.PublicUseSSL, // Use separate SSL setting for public URLs
			Region: cfg.Region,
		})
		if err != nil {
			// If public client fails, we'll fall back to internal client
			publicClient = nil
		}
	}

	return &S3Client{
		client:         client,
		publicClient:   publicClient,
		bucketName:     cfg.BucketName,
		maxSize:        cfg.MaxFileSize,
		publicEndpoint: cfg.PublicEndpoint,
		useSSL:         cfg.UseSSL,
		publicUseSSL:   cfg.PublicUseSSL,
	}, nil
}

// EnsureBucket creates the bucket if it doesn't exist
func (s *S3Client) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucketName)
	if err != nil {
		return fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = s.client.MakeBucket(ctx, s.bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return fmt.Errorf("failed to create bucket: %w", err)
		}
	}

	return nil
}

// Upload uploads a file to S3/MinIO storage
func (s *S3Client) Upload(ctx context.Context, key string, reader io.Reader, size int64, contentType string) (*FileInfo, error) {
	if size > s.maxSize {
		return nil, fmt.Errorf("file size %d exceeds maximum allowed size %d", size, s.maxSize)
	}

	info, err := s.client.PutObject(ctx, s.bucketName, key, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &FileInfo{
		Key:         info.Key,
		Size:        info.Size,
		ContentType: contentType,
		ETag:        info.ETag,
		UploadedAt:  time.Now(),
	}, nil
}

// Download downloads a file from S3/MinIO storage
func (s *S3Client) Download(ctx context.Context, key string) (io.ReadCloser, *FileInfo, error) {
	obj, err := s.client.GetObject(ctx, s.bucketName, key, minio.GetObjectOptions{})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get object: %w", err)
	}

	stat, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, nil, fmt.Errorf("failed to get object stat: %w", err)
	}

	return obj, &FileInfo{
		Key:         stat.Key,
		Size:        stat.Size,
		ContentType: stat.ContentType,
		ETag:        stat.ETag,
		UploadedAt:  stat.LastModified,
	}, nil
}

// Delete removes a file from S3/MinIO storage
func (s *S3Client) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucketName, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete object: %w", err)
	}
	return nil
}

// Exists checks if a file exists in storage
func (s *S3Client) Exists(ctx context.Context, key string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucketName, key, minio.StatObjectOptions{})
	if err != nil {
		errResponse := minio.ToErrorResponse(err)
		if errResponse.Code == "NoSuchKey" {
			return false, nil
		}
		return false, fmt.Errorf("failed to check object existence: %w", err)
	}
	return true, nil
}

// GetPresignedURL generates a presigned URL for downloading a file
func (s *S3Client) GetPresignedURL(ctx context.Context, key string, expires time.Duration) (string, error) {
	// Use public client if available (generates URLs with correct host signature)
	clientToUse := s.client
	if s.publicClient != nil {
		clientToUse = s.publicClient
	}

	presignedURL, err := clientToUse.PresignedGetObject(ctx, s.bucketName, key, expires, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}

	return presignedURL.String(), nil
}

// GenerateKey generates a unique storage key for a document
func GenerateKey(documentID int64, fileName string) string {
	timestamp := time.Now().UnixNano()
	ext := path.Ext(fileName)
	return fmt.Sprintf("documents/%d/%d%s", documentID, timestamp, ext)
}

// GenerateTempKey generates a temporary storage key before document is created
func GenerateTempKey(userID int64, fileName string) string {
	timestamp := time.Now().UnixNano()
	ext := path.Ext(fileName)
	return fmt.Sprintf("temp/%d/%d%s", userID, timestamp, ext)
}

// Ping checks if the S3 connection is healthy
func (s *S3Client) Ping(ctx context.Context) error {
	_, err := s.client.BucketExists(ctx, s.bucketName)
	return err
}

// Close closes the S3 client (no-op for minio client, but good for interface consistency)
func (s *S3Client) Close() error {
	return nil
}

// MaxFileSize returns the maximum allowed file size
func (s *S3Client) MaxFileSize() int64 {
	return s.maxSize
}

// BucketName returns the bucket name
func (s *S3Client) BucketName() string {
	return s.bucketName
}
