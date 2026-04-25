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

type QuizDoneWorkerDeps struct {
	Quizzes port.QuizRepository
}

func StartQuizDoneWorker(ctx context.Context, js jetstream.JetStream, deps QuizDoneWorkerDeps, log *zap.Logger) error {
	cons, err := nats.CreateOrUpdateConsumer(ctx, js, nats.StreamQuizDone, nats.DurableQuizDone, nats.SubjectQuizDone)
	if err != nil {
		return err
	}

	log.Info("quiz_done_worker started")

	return nats.ConsumeLoop(ctx, cons, func(msg jetstream.Msg) {
		if err := handleQuizDone(ctx, msg, deps, log); err != nil {
			log.Error("quiz_done_worker handle failed", zap.Error(err))
			msg.Nak()
			return
		}
		msg.Ack()
	})
}

type quizDonePayload struct {
	Status     string        `json:"status"`
	CourseID   string        `json:"course_id"`
	Difficulty string        `json:"difficulty"`
	Questions  []rawQuestion `json:"questions"`
}

type rawQuestion struct {
	Type     string          `json:"type"`
	Question string          `json:"question"`
	Choices  json.RawMessage `json:"choices"`
	Answer   string          `json:"answer"`
}

func handleQuizDone(ctx context.Context, msg jetstream.Msg, deps QuizDoneWorkerDeps, log *zap.Logger) error {
	var payload quizDonePayload
	if err := json.Unmarshal(msg.Data(), &payload); err != nil {
		log.Warn("quiz_done bad json, dropping", zap.Error(err))
		msg.Ack()
		return nil
	}

	courseID, err := uuid.Parse(payload.CourseID)
	if err != nil {
		log.Warn("quiz_done bad course_id, dropping", zap.String("course_id", payload.CourseID))
		msg.Ack()
		return nil
	}

	difficulty := domain.Difficulty(payload.Difficulty)

	quiz, err := deps.Quizzes.GetQuizByCourseAndDifficulty(ctx, courseID, difficulty)
	if err != nil {
		return fmt.Errorf("get quiz: %w", err)
	}

	if payload.Status != "success" {
		log.Info("quiz_done failed", zap.String("course_id", payload.CourseID), zap.String("difficulty", payload.Difficulty))
		return deps.Quizzes.UpdateQuizStatus(ctx, quiz.ID, domain.QuizFailed)
	}

	if err := deps.Quizzes.DeleteQuestionsByQuiz(ctx, quiz.ID); err != nil {
		return fmt.Errorf("delete old questions: %w", err)
	}

	questions := make([]domain.Question, len(payload.Questions))
	for i, rq := range payload.Questions {
		var choices []domain.Choice
		if len(rq.Choices) > 0 && string(rq.Choices) != "null" {
			if err := json.Unmarshal(rq.Choices, &choices); err != nil {
				return fmt.Errorf("unmarshal choices for question %d: %w", i, err)
			}
		}

		questions[i] = domain.Question{
			QuizID:   quiz.ID,
			Type:     domain.QuestionType(rq.Type),
			Question: rq.Question,
			Choices:  choices,
			Answer:   rq.Answer,
			OrderIdx: i,
		}
	}

	if err := deps.Quizzes.CreateQuestions(ctx, questions); err != nil {
		return fmt.Errorf("insert questions: %w", err)
	}

	log.Info("quiz_done success",
		zap.String("course_id", payload.CourseID),
		zap.String("difficulty", payload.Difficulty),
		zap.Int("questions", len(questions)),
	)

	return deps.Quizzes.UpdateQuizStatus(ctx, quiz.ID, domain.QuizReady)
}
