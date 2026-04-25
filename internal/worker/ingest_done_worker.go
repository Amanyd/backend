package worker

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/infra/nats"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go/jetstream"
	"go.uber.org/zap"
)

type IngestDoneWorkerDeps struct {
	Files   port.FileRepository
	Lessons port.LessonRepository
	Quizzes port.QuizRepository
	Queue   port.MessageQueue
}

var difficulties = []domain.Difficulty{domain.DifficultyEasy, domain.DifficultyMedium, domain.DifficultyHard}

func StartIngestDoneWorker(ctx context.Context, js jetstream.JetStream, deps IngestDoneWorkerDeps, log *zap.Logger) error {
	cons, err := nats.CreateOrUpdateConsumer(ctx, js, nats.StreamIngestDone, nats.DurableIngestDone, nats.SubjectIngestDone)
	if err != nil {
		return err
	}

	log.Info("ingest_done_worker started")

	return nats.ConsumeLoop(ctx, cons, func(msg jetstream.Msg) {
		if err := handleIngestDone(ctx, msg, deps, log); err != nil {
			log.Error("ingest_done_worker handle failed", zap.Error(err))
			msg.Nak()
			return
		}
		msg.Ack()
	})
}

type ingestDonePayload struct {
	Status  string `json:"status"`
	FileID  string `json:"file_id"`
}

func handleIngestDone(ctx context.Context, msg jetstream.Msg, deps IngestDoneWorkerDeps, log *zap.Logger) error {
	var payload ingestDonePayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		log.Warn("ingest_done bad json, dropping", zap.Error(err))
		msg.Ack()
		return nil
	}

	fileID, err := uuid.Parse(payload.FileID)
	if err != nil {
		log.Warn("ingest_done bad file_id, dropping", zap.String("file_id", payload.FileID))
		msg.Ack()
		return nil
	}

	if payload.Status != "success" {
		log.Info("ingest_done failed", zap.String("file_id", payload.FileID))
		return deps.Files.UpdateIngestStatus(ctx, fileID, domain.IngestFailed)
	}

	if err := deps.Files.UpdateIngestStatus(ctx, fileID, domain.IngestReady); err != nil {
		return fmt.Errorf("update file status: %w", err)
	}

	file, err := deps.Files.GetByID(ctx, fileID)
	if err != nil {
		return fmt.Errorf("get file: %w", err)
	}

	lesson, err := deps.Lessons.GetByID(ctx, file.LessonID)
	if err != nil {
		return fmt.Errorf("get lesson: %w", err)
	}

	allReady, err := deps.Files.AllReadyForCourse(ctx, lesson.CourseID)
	if err != nil {
		return fmt.Errorf("check all ready: %w", err)
	}
	if !allReady {
		log.Info("ingest_done not all files ready yet", zap.String("course_id", lesson.CourseID.String()))
		return nil
	}

	log.Info("all files ready, triggering quiz generation", zap.String("course_id", lesson.CourseID.String()))

	for _, diff := range difficulties {
		quiz := &domain.Quiz{
			CourseID:   lesson.CourseID,
			Difficulty: diff,
			Status:     domain.QuizGenerating,
		}
		if err := deps.Quizzes.CreateQuiz(ctx, quiz); err != nil {
			return fmt.Errorf("create quiz %s: %w", diff, err)
		}

		payload, err := json.Marshal(map[string]any{
			"course_id":    lesson.CourseID.String(),
			"difficulty":   string(diff),
			"limit_chunks": 20,
		})
		if err != nil {
			return fmt.Errorf("marshal quiz request: %w", err)
		}

		if err := deps.Queue.Publish(ctx, nats.SubjectQuizRequest, payload); err != nil {
			return fmt.Errorf("publish quiz request %s: %w", diff, err)
		}
	}

	return nil
}
