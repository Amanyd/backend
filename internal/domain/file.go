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
	ID           uuid.UUID
	LessonID     uuid.UUID
	FileName     string
	FileType     FileType
	MinioKey     string
	IngestStatus IngestStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
}
