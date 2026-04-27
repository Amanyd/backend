package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
)

type CourseService struct {
	courses port.CourseRepository
	lessons port.LessonRepository
	cache   port.Cache
}

func NewCourseService(courses port.CourseRepository, lessons port.LessonRepository, cache port.Cache) *CourseService {
	return &CourseService{courses: courses, lessons: lessons, cache: cache}
}

func (s *CourseService) Create(ctx context.Context, title, desc, rank string, instructorID uuid.UUID) (*domain.Course, error) {
	course := &domain.Course{
		Title:        title,
		Description:  desc,
		Rank:         rank,
		InstructorID: instructorID,
	}
	if err := s.courses.Create(ctx, course); err != nil {
		return nil, err
	}
	s.cache.Delete(ctx, "courses:rank:"+rank)
	s.cache.Delete(ctx, "courses:all")
	return course, nil
}

func (s *CourseService) GetByID(ctx context.Context, courseID uuid.UUID) (*domain.Course, error) {
	key := "course:" + courseID.String()

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var course domain.Course
		if json.Unmarshal([]byte(cached), &course) == nil {
			return &course, nil
		}
	}

	course, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(course); err == nil {
		s.cache.Set(ctx, key, string(data), 10*time.Minute)
	}
	return course, nil
}

func (s *CourseService) ListByRank(ctx context.Context, rank string) ([]domain.Course, error) {
	key := "courses:rank:" + rank

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var courses []domain.Course
		if json.Unmarshal([]byte(cached), &courses) == nil {
			return courses, nil
		}
	}

	courses, err := s.courses.ListByRank(ctx, rank)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(courses); err == nil {
		s.cache.Set(ctx, key, string(data), 5*time.Minute)
	}
	return courses, nil
}

func (s *CourseService) ListAll(ctx context.Context) ([]domain.Course, error) {
	key := "courses:all"

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var courses []domain.Course
		if json.Unmarshal([]byte(cached), &courses) == nil {
			return courses, nil
		}
	}

	courses, err := s.courses.ListAll(ctx)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(courses); err == nil {
		s.cache.Set(ctx, key, string(data), 5*time.Minute)
	}
	return courses, nil
}

func (s *CourseService) ListByInstructor(ctx context.Context, instructorID uuid.UUID) ([]domain.Course, error) {
	return s.courses.ListByInstructor(ctx, instructorID)
}

func (s *CourseService) Update(ctx context.Context, courseID, instructorID uuid.UUID, title, desc, rank string) (*domain.Course, error) {
	course, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		return nil, err
	}
	if course.InstructorID != instructorID {
		return nil, domain.ErrForbidden
	}

	oldRank := course.Rank
	course.Title = title
	course.Description = desc
	course.Rank = rank
	if err := s.courses.Update(ctx, course); err != nil {
		return nil, err
	}

	s.cache.Delete(ctx, "course:"+courseID.String())
	s.cache.Delete(ctx, "courses:rank:"+oldRank)
	s.cache.Delete(ctx, "courses:all")
	if rank != oldRank {
		s.cache.Delete(ctx, "courses:rank:"+rank)
	}
	return course, nil
}

func (s *CourseService) Delete(ctx context.Context, courseID, instructorID uuid.UUID) error {
	course, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		return err
	}
	if course.InstructorID != instructorID {
		return domain.ErrForbidden
	}
	if err := s.courses.Delete(ctx, courseID); err != nil {
		return err
	}
	s.cache.Delete(ctx, "course:"+courseID.String())
	s.cache.Delete(ctx, "courses:rank:"+course.Rank)
	s.cache.Delete(ctx, "courses:all")
	return nil
}

// Lesson methods

func (s *CourseService) CreateLesson(ctx context.Context, courseID, instructorID uuid.UUID, title string, orderIdx int) (*domain.Lesson, error) {
	if err := s.verifyOwnership(ctx, courseID, instructorID); err != nil {
		return nil, err
	}

	lesson := &domain.Lesson{
		CourseID:  courseID,
		Title:    title,
		OrderIdx: orderIdx,
	}
	if err := s.lessons.Create(ctx, lesson); err != nil {
		return nil, err
	}
	s.cache.Delete(ctx, "lessons:course:"+courseID.String())
	return lesson, nil
}

func (s *CourseService) ListLessons(ctx context.Context, courseID uuid.UUID) ([]domain.Lesson, error) {
	key := "lessons:course:" + courseID.String()

	if cached, err := s.cache.Get(ctx, key); err == nil {
		var lessons []domain.Lesson
		if json.Unmarshal([]byte(cached), &lessons) == nil {
			return lessons, nil
		}
	}

	lessons, err := s.lessons.ListByCourse(ctx, courseID)
	if err != nil {
		return nil, err
	}

	if data, err := json.Marshal(lessons); err == nil {
		s.cache.Set(ctx, key, string(data), 5*time.Minute)
	}
	return lessons, nil
}

func (s *CourseService) UpdateLesson(ctx context.Context, lessonID, instructorID uuid.UUID, title string, orderIdx int) (*domain.Lesson, error) {
	lesson, err := s.lessons.GetByID(ctx, lessonID)
	if err != nil {
		return nil, err
	}
	if err := s.verifyOwnership(ctx, lesson.CourseID, instructorID); err != nil {
		return nil, err
	}

	lesson.Title = title
	lesson.OrderIdx = orderIdx
	if err := s.lessons.Update(ctx, lesson); err != nil {
		return nil, err
	}
	s.cache.Delete(ctx, "lessons:course:"+lesson.CourseID.String())
	return lesson, nil
}

func (s *CourseService) DeleteLesson(ctx context.Context, lessonID, instructorID uuid.UUID) error {
	lesson, err := s.lessons.GetByID(ctx, lessonID)
	if err != nil {
		return err
	}
	if err := s.verifyOwnership(ctx, lesson.CourseID, instructorID); err != nil {
		return err
	}
	if err := s.lessons.Delete(ctx, lessonID); err != nil {
		return err
	}
	s.cache.Delete(ctx, "lessons:course:"+lesson.CourseID.String())
	return nil
}

func (s *CourseService) verifyOwnership(ctx context.Context, courseID, instructorID uuid.UUID) error {
	course, err := s.courses.GetByID(ctx, courseID)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return domain.ErrNotFound
		}
		return err
	}
	if course.InstructorID != instructorID {
		return domain.ErrForbidden
	}
	return nil
}
