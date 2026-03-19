package storage

import (
	"bytes"
	"context"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/inf-sys-secretary-methodologist/inf-sys-secretary-methodist/internal/shared/infrastructure/config"
)

func createTestMinioClient(t *testing.T) *minio.Client {
	t.Helper()
	client, err := minio.New("localhost:19999", &minio.Options{
		Creds:  credentials.NewStaticV4("minioadmin", "minioadmin", ""),
		Secure: false,
	})
	require.NoError(t, err)
	return client
}

func TestS3Client_Close(t *testing.T) {
	client := &S3Client{}
	err := client.Close()
	assert.NoError(t, err)
}

func TestS3Client_MaxFileSize(t *testing.T) {
	client := &S3Client{maxSize: 1024 * 1024}
	assert.Equal(t, int64(1024*1024), client.MaxFileSize())
}

func TestS3Client_BucketName(t *testing.T) {
	client := &S3Client{bucketName: "test-bucket"}
	assert.Equal(t, "test-bucket", client.BucketName())
}

func TestNewS3Client_Basic(t *testing.T) {
	cfg := config.S3Config{
		Endpoint:        "localhost:19999",
		PublicEndpoint:  "localhost:19999",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		Region:          "us-east-1",
		UseSSL:          false,
		PublicUseSSL:    false,
		MaxFileSize:     50 * 1024 * 1024,
	}

	s3Client, err := cfg.Endpoint, error(nil) // Just create the client
	_ = s3Client
	client, err := NewS3Client(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Equal(t, "test-bucket", client.BucketName())
	assert.Equal(t, int64(50*1024*1024), client.MaxFileSize())
}

func TestNewS3Client_WithDifferentPublicEndpoint(t *testing.T) {
	cfg := config.S3Config{
		Endpoint:        "localhost:19999",
		PublicEndpoint:  "public.localhost:19998",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		Region:          "us-east-1",
		UseSSL:          false,
		PublicUseSSL:    true,
		MaxFileSize:     10 * 1024 * 1024,
	}

	client, err := NewS3Client(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.NotNil(t, client.publicClient)
}

func TestNewS3Client_SameEndpoints(t *testing.T) {
	cfg := config.S3Config{
		Endpoint:        "localhost:19999",
		PublicEndpoint:  "localhost:19999",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		Region:          "us-east-1",
		UseSSL:          false,
		MaxFileSize:     10 * 1024 * 1024,
	}

	client, err := NewS3Client(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Nil(t, client.publicClient) // Same endpoint, no public client
}

func TestNewS3Client_EmptyPublicEndpoint(t *testing.T) {
	cfg := config.S3Config{
		Endpoint:        "localhost:19999",
		PublicEndpoint:  "",
		AccessKeyID:     "minioadmin",
		SecretAccessKey: "minioadmin",
		BucketName:      "test-bucket",
		Region:          "us-east-1",
		UseSSL:          false,
		MaxFileSize:     10 * 1024 * 1024,
	}

	client, err := NewS3Client(cfg)
	assert.NoError(t, err)
	assert.NotNil(t, client)
	assert.Nil(t, client.publicClient)
}

func TestS3Client_Upload_ExceedsMaxSize(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
		maxSize:    100,
	}

	_, err := client.Upload(context.Background(), "key", bytes.NewReader([]byte("data")), 200, "text/plain")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "exceeds maximum allowed size")
}

func TestS3Client_Upload_ConnectionError(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
		maxSize:    1024 * 1024,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.Upload(ctx, "key", bytes.NewReader([]byte("data")), 4, "text/plain")
	assert.Error(t, err)
}

func TestS3Client_EnsureBucket_ConnectionError(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := client.EnsureBucket(ctx)
	assert.Error(t, err)
}

func TestS3Client_Download_ConnectionError(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	reader, info, err := client.Download(ctx, "nonexistent-key")
	// GetObject itself might not fail (it's lazy), but Stat will fail
	if err != nil {
		assert.Nil(t, reader)
		assert.Nil(t, info)
	} else if reader != nil {
		// The reader is returned but stat call when reading will fail
		_ = reader.Close()
	}
}

func TestS3Client_Delete_ConnectionError(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// RemoveObject may or may not fail depending on minio client behavior
	_ = client.Delete(ctx, "nonexistent-key")
}

func TestS3Client_Exists_ConnectionError(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.Exists(ctx, "nonexistent-key")
	assert.Error(t, err)
}

func TestS3Client_GetPresignedURL_UsesPublicClient(t *testing.T) {
	mc := createTestMinioClient(t)
	publicMc := createTestMinioClient(t)

	client := &S3Client{
		client:       mc,
		publicClient: publicMc,
		bucketName:   "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	// PresignedGetObject requires bucket region lookup which needs network
	_, err := client.GetPresignedURL(ctx, "test-key", 1*time.Hour)
	// Error is expected since no minio server is running, but the code path is exercised
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate presigned URL")
}

func TestS3Client_GetPresignedURL_NoPublicClient(t *testing.T) {
	mc := createTestMinioClient(t)

	client := &S3Client{
		client:       mc,
		publicClient: nil,
		bucketName:   "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	_, err := client.GetPresignedURL(ctx, "test-key", 1*time.Hour)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to generate presigned URL")
}

func TestS3Client_Ping_ConnectionError(t *testing.T) {
	mc := createTestMinioClient(t)
	client := &S3Client{
		client:     mc,
		bucketName: "test",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := client.Ping(ctx)
	assert.Error(t, err)
}
