package tus

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	awscreds "github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/Amanyd/backend/internal/config"
	"github.com/Amanyd/backend/internal/domain"
	natsinfra "github.com/Amanyd/backend/internal/infra/nats"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
	"github.com/tus/tusd/v2/pkg/handler"
	"github.com/tus/tusd/v2/pkg/s3store"
	"go.uber.org/zap"
)

type TUSDeps struct {
	Files   port.FileRepository
	Lessons port.LessonRepository
	Courses port.CourseRepository
	Queue   port.MessageQueue
	Bucket  string
}

func NewTUSHandler(cfg config.MinIOConfig, deps TUSDeps, logger *zap.Logger) (*handler.Handler, error) {
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
		BasePath:                "/api/v1/files/tus/",
		StoreComposer:           composer,
		NotifyCompleteUploads:   true,
		NotifyCreatedUploads:    true,
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
			if err := handleUploadComplete(context.Background(), event, deps, logger); err != nil {
				logger.Error("tus complete hook failed", zap.Error(err))
			}
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

func handleUploadComplete(ctx context.Context, event handler.HookEvent, deps TUSDeps, log *zap.Logger) error {
	meta := event.Upload.MetaData

	lessonID, err := uuid.Parse(meta["lesson_id"])
	if err != nil {
		return fmt.Errorf("parse lesson_id: %w", err)
	}

	fileName := meta["file_name"]
	fileType := domain.FileType(meta["file_type"])
	instructorID, err := uuid.Parse(meta["instructor_id"])
	if err != nil {
		return fmt.Errorf("parse instructor_id: %w", err)
	}

	lesson, err := deps.Lessons.GetByID(ctx, lessonID)
	if err != nil {
		return fmt.Errorf("get lesson: %w", err)
	}

	course, err := deps.Courses.GetByID(ctx, lesson.CourseID)
	if err != nil {
		return fmt.Errorf("get course: %w", err)
	}
	if course.InstructorID != instructorID {
		return fmt.Errorf("instructor %s does not own course %s", instructorID, course.ID)
	}

	// The S3 store generates IDs in the format "objectKey+multipartUploadID".
	// We only need the objectKey portion to reference the object in MinIO.
	minioKey := event.Upload.ID
	if idx := strings.Index(minioKey, "+"); idx > 0 {
		minioKey = minioKey[:idx]
	}

	file := &domain.FileAsset{
		LessonID:     lessonID,
		FileName:     fileName,
		FileType:     fileType,
		MinioKey:     minioKey,
		IngestStatus: domain.IngestPending,
	}
	if err := deps.Files.Create(ctx, file); err != nil {
		return fmt.Errorf("create file: %w", err)
	}

	payload, err := json.Marshal(map[string]string{
		"bucket":     deps.Bucket,
		"key":        minioKey,
		"course_id":  course.ID.String(),
		"file_id":    file.ID.String(),
		"file_name":  fileName,
		"teacher_id": instructorID.String(),
	})
	if err != nil {
		return fmt.Errorf("marshal ingest payload: %w", err)
	}

	if err := deps.Queue.Publish(ctx, natsinfra.SubjectIngestRequest, payload); err != nil {
		return fmt.Errorf("publish ingest: %w", err)
	}

	if err := deps.Files.UpdateIngestStatus(ctx, file.ID, domain.IngestProcessing); err != nil {
		return fmt.Errorf("update status: %w", err)
	}

	log.Info("tus upload complete, ingest triggered",
		zap.String("file_id", file.ID.String()),
		zap.String("course_id", course.ID.String()),
	)
	return nil
}
