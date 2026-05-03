package domain

import (
	"time"

	"github.com/google/uuid"
)

type IngestStatus string

const (
	IngestPending    IngestStatus = "pending"
	IngestProcessing IngestStatus = "processing"
	IngestReady      IngestStatus = "ready"
	IngestFailed     IngestStatus = "failed"
)

type FileType string

const (
	FileTypePDF  FileType = "pdf"
	FileTypePPT  FileType = "ppt"
	FileTypeDOCX FileType = "docx"
)

type FileAsset struct {
	ID           uuid.UUID    `json:"id"`
	LessonID     uuid.UUID    `json:"lesson_id"`
	FileName     string       `json:"file_name"`
	FileType     FileType     `json:"file_type"`
	MinioKey     string       `json:"minio_key"`
	IngestStatus IngestStatus `json:"ingest_status"`
	CreatedAt    time.Time    `json:"created_at"`
	UpdatedAt    time.Time    `json:"updated_at"`
}
