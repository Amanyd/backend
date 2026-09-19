package minio

import (
	"context"
	"fmt"
	"net/url"
	"path/filepath"
	"time"

	"github.com/Amanyd/backend/internal/port"
	"github.com/minio/minio-go/v7"
)

type storage struct {
	client *minio.Client
	bucket string
}

func NewStorage(client *minio.Client, bucket string) port.ObjectStorage {
	return &storage{client: client, bucket: bucket}
}

func (s *storage) PresignUpload(ctx context.Context, key string, expiry time.Duration) (string, error) {
	u, err := s.client.PresignedPutObject(ctx, s.bucket, key, expiry)
	if err != nil {
		return "", fmt.Errorf("minio presign upload: %w", err)
	}
	return u.String(), nil
}

func (s *storage) PresignView(ctx context.Context, key string, expiry time.Duration) (string, error) {
	reqParams := make(url.Values)
	reqParams.Set("response-content-disposition", "inline")
	
	ext := filepath.Ext(key)
	if ext == ".pdf" {
		reqParams.Set("response-content-type", "application/pdf")
	} else if ext == ".docx" {
		reqParams.Set("response-content-type", "application/vnd.openxmlformats-officedocument.wordprocessingml.document")
	} else if ext == ".ppt" {
		reqParams.Set("response-content-type", "application/vnd.ms-powerpoint")
	} else if ext == ".pptx" {
		reqParams.Set("response-content-type", "application/vnd.openxmlformats-officedocument.presentationml.presentation")
	}

	u, err := s.client.PresignedGetObject(ctx, s.bucket, key, expiry, reqParams)
	if err != nil {
		return "", fmt.Errorf("minio presign view: %w", err)
	}
	return u.String(), nil
}

func (s *storage) Delete(ctx context.Context, key string) error {
	err := s.client.RemoveObject(ctx, s.bucket, key, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("minio delete object: %w", err)
	}
	return nil
}
