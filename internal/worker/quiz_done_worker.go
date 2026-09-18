package worker

import (
	"context"
	"encoding/json"
	"errors"
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
	Lessons port.LessonRepository
	Queue   port.MessageQueue
}

func StartQuizDoneWorker(ctx context.Context, js jetstream.JetStream, deps QuizDoneWorkerDeps, log *zap.Logger) error {
	cons, err := nats.CreateOrUpdateConsumer(ctx, js, nats.StreamQuizDone, nats.DurableQuizDone, nats.SubjectQuizDone)
	if err != nil {
		return err
	}

	log.Info("quiz_done_worker started")

	return nats.ConsumeLoop(ctx, cons, func(msg jetstream.Msg) {
		// If the parent context is already cancelled (shutdown), ack the message
		// so it doesn't get redelivered in an infinite error loop.
		select {
		case <-ctx.Done():
			msg.Ack()
			return
		default:
		}

		if err := handleQuizDone(context.Background(), msg, deps, log); err != nil {
			log.Error("quiz_done_worker handle failed", zap.Error(err))
			msg.Nak()
			return
		}
		msg.Ack()
	})
}

type quizDonePayload struct {
	Type       string        `json:"type"`
	Status     string        `json:"status"`
	CourseID   string        `json:"course_id"`
	LessonID   string        `json:"lesson_id,omitempty"`
	Difficulty string        `json:"difficulty"`
	Keywords   []string      `json:"keywords,omitempty"`
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

	var quiz *domain.Quiz
	var getErr error
	if payload.Type == "lesson" {
		lessonID, err := uuid.Parse(payload.LessonID)
		if err != nil {
			log.Warn("quiz_done bad lesson_id, dropping", zap.String("lesson_id", payload.LessonID))
			msg.Ack()
			return nil
		}

		if len(payload.Keywords) > 0 {
			b, err := json.Marshal(payload.Keywords)
			if err != nil {
				log.Error("failed to marshal keywords", zap.Error(err))
			} else {
				if err := deps.Lessons.UpdateKeywords(ctx, lessonID, b); err != nil {
					log.Error("failed to save keywords to lesson", zap.Error(err))
				}
			}
		}

		quiz, getErr = deps.Quizzes.GetQuizByLessonAndDifficulty(ctx, lessonID, difficulty)
	} else {
		quiz, getErr = deps.Quizzes.GetQuizByCourseAndDifficulty(ctx, courseID, difficulty)
	}

	if getErr != nil {
		if errors.Is(getErr, domain.ErrNotFound) {
			log.Warn("quiz_done quiz not found, dropping",
				zap.String("course_id", payload.CourseID),
				zap.String("type", payload.Type),
			)
			msg.Ack()
			return nil
		}
		return fmt.Errorf("get quiz: %w", getErr)
	}

	if payload.Status != "success" {
		log.Info("quiz_done failed", zap.String("course_id", payload.CourseID), zap.String("type", payload.Type))
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
		zap.String("type", payload.Type),
		zap.Int("questions", len(questions)),
	)

	if err := deps.Quizzes.UpdateQuizStatus(ctx, quiz.ID, domain.QuizReady); err != nil {
		return fmt.Errorf("update quiz status: %w", err)
	}

	if payload.Type == "lesson" {
		// Saga Trigger: Check if all lesson quizzes for this course are ready.
		allQuizzes, err := deps.Quizzes.ListQuizzesByCourse(ctx, courseID)
		if err != nil {
			return fmt.Errorf("list quizzes to check saga: %w", err)
		}

		allReady := true
		hasLessons := false
		for _, q := range allQuizzes {
			if q.LessonID != nil {
				hasLessons = true
				if q.Status != domain.QuizReady {
					allReady = false
					break
				}
			}
		}

		if hasLessons && allReady {
			log.Info("All lesson quizzes ready. Triggering course quizzes.", zap.String("course_id", courseID.String()))
			
			// Fetch all lessons for the course to get their keywords
			lessons, err := deps.Lessons.ListByCourse(ctx, courseID)
			if err != nil {
				log.Error("failed to list lessons for saga", zap.Error(err))
				// We can continue, just won't have keywords
			}
			
			var allKeywords []string
			for _, lesson := range lessons {
				if len(lesson.Keywords) > 0 {
					var kw []string
					if err := json.Unmarshal(lesson.Keywords, &kw); err == nil {
						allKeywords = append(allKeywords, kw...)
					}
				}
			}

			for _, diff := range []domain.Difficulty{domain.DifficultyEasy, domain.DifficultyMedium, domain.DifficultyHard} {
				q := &domain.Quiz{
					CourseID:   courseID,
					Difficulty: diff,
					Status:     domain.QuizGenerating,
				}
				if err := deps.Quizzes.CreateQuiz(ctx, q); err != nil {
					log.Error("Failed to create course quiz", zap.Error(err))
					continue
				}

				reqPayload, err := json.Marshal(map[string]any{
					"type":         "course",
					"course_id":    courseID.String(),
					"difficulty":   string(diff),
					"keywords":     allKeywords,
					"limit_chunks": 50,
				})
				if err != nil {
					log.Error("Failed to marshal course quiz request", zap.Error(err))
					continue
				}
				if err := deps.Queue.Publish(ctx, nats.SubjectQuizRequest, reqPayload); err != nil {
					log.Error("Failed to publish course quiz request", zap.Error(err))
					continue
				}
			}
		}
	}

	return nil
}
