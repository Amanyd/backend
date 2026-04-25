package tus

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/eulerbutcooler/wingman-backend/internal/config"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"
	"go.uber.org/zap"
)

func NewTUSHandler(cfg config.MinIOConfig, logger *zap.Logger) (*handler.Handler, error) {
	scheme := "http"
	if cfg.MinIOUseSSL {
		scheme = "https"
	}
	fullEndpoint := fmt.Sprintf("%s://%s", scheme, cfg.MinIOEndpoint)

	s3Client := s3.New(s3.Options{
		BaseEndpoint: aws.String(fullEndpoint),
		Region:       "us-east-1",
		Credentials:  awscreds.NewStaticCredentialsProvider(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		UsePathStyle: true,
	})

	store := s3store.New(cfg.MinIOBucket, s3Client)

	composer := handler.NewStoreComposer()
	store.UseIn(composer)

	h, err := handler.NewHandler(handler.Config{
		BasePath:      "/api/v1/files/upload/",
		StoreComposer: composer,
	})
	if err != nil {
		return nil, fmt.Errorf("tus handler: %w", err)
	}

	go func() {
		for {
			event, ok := <-h.CompleteUploads
			if !ok {
				return
			}
			logger.Debug("tus upload complete",
				zap.String("id", event.Upload.ID),
				zap.Int64("size", event.Upload.Size),
			)
		}
	}()

	go func() {
		for {
			event, ok := <-h.CreatedUploads
			if !ok {
				return
			}
			logger.Debug("tus upload created",
				zap.String("id", event.Upload.ID),
				zap.Int64("size", event.Upload.Size),
			)
		}
	}()

	return h, nil
}

func ExtractMetadata(info handler.FileInfo) map[string]string {
	return info.MetaData
}

func MakeObjectKey(courseID, fileID, fileName string) string {
	return fmt.Sprintf("courses/%s/%s/%s", courseID, fileID, fileName)
}

type contextKey string

const uploadInfoKey contextKey = "tusUploadInfo"

func WithUploadInfo(ctx context.Context, info handler.FileInfo) context.Context {
	return context.WithValue(ctx, uploadInfoKey, info)
}

func UploadInfoFromCtx(ctx context.Context) (handler.FileInfo, bool) {
	info, ok := ctx.Value(uploadInfoKey).(handler.FileInfo)
	return info, ok
}
