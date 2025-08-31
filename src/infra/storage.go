package storage

import (
	"context"
	"io"
	"time"

	"api/infra/config"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Storage struct {
	client     *minio.Client
	bucketName string
}

type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
	Region          string
}

func NewStorage(endpoint, accessKey, secretKey, bucket, region string, useSSL bool) (*Storage, error) {
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}

	// Check if bucket exists and create if not
	ctx := context.Background()
	exists, err := client.BucketExists(ctx, bucket)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = client.MakeBucket(ctx, bucket, minio.MakeBucketOptions{
			Region: region,
		})
		if err != nil {
			return nil, err
		}
	}

	return &Storage{
		client:     client,
		bucketName: bucket,
	}, nil
}

func (s *Storage) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) error {
	_, err := s.client.PutObject(ctx, s.bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *Storage) Download(ctx context.Context, objectName string) (io.ReadCloser, error) {
	return s.client.GetObject(ctx, s.bucketName, objectName, minio.GetObjectOptions{})
}

func (s *Storage) Delete(ctx context.Context, objectName string) error {
	return s.client.RemoveObject(ctx, s.bucketName, objectName, minio.RemoveObjectOptions{})
}

func (s *Storage) GetPresignedURL(ctx context.Context, objectName string, expiry int) (string, error) {
	url, err := s.client.PresignedGetObject(ctx, s.bucketName, objectName, time.Duration(expiry)*time.Second, nil)
	if err != nil {
		return "", err
	}
	return url.String(), nil
}