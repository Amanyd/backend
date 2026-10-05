package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/Amanyd/backend/internal/domain"
	"github.com/Amanyd/backend/internal/infra/postgres/gen"
	"github.com/Amanyd/backend/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type analyticsRepo struct {
	pool *pgxpool.Pool
	q    *gen.Queries
}

func NewAnalyticsRepo(pool *pgxpool.Pool) port.AnalyticsRepo {
	return &analyticsRepo{pool: pool, q: gen.New(pool)}
}

func (r *analyticsRepo) RecordEvent(ctx context.Context, event *domain.Event) error {
	metadataJSON, err := json.Marshal(event.Metadata)
	if err != nil {
		return err
	}

	var courseID pgtype.UUID
	if event.CourseID != nil {
		courseID = pgtype.UUID{Bytes: *event.CourseID, Valid: true}
	}

	row, err := r.q.RecordEvent(ctx, gen.RecordEventParams{
		UserID:   event.UserID,
		CourseID: courseID,
		Type:     string(event.Type),
		Metadata: metadataJSON,
	})
	if err != nil {
		return err
	}
	event.ID = row.ID
	event.CreatedAt = row.CreatedAt
	return nil
}

func (r *analyticsRepo) GetCourseMetrics(ctx context.Context, courseID uuid.UUID) (*domain.Metric, error) {
	row, err := r.q.GetCourseMetrics(ctx, courseID)
	if err != nil {
		return nil, err
	}
	avgScore, err := toFloat64(row.AvgQuizScore)
	if err != nil {
		return nil, fmt.Errorf("avg_quiz_score: %w", err)
	}

	return &domain.Metric{
		CourseID:      row.CourseID,
		TotalStudents: int(row.TotalStudents),
		AvgQuizScore:  avgScore,
		TotalMessages: int(row.TotalMessages),
		TotalFiles:    int(row.TotalFiles),
	}, nil
}

func (r *analyticsRepo) GetStudentScores(ctx context.Context, courseID uuid.UUID) ([]domain.StudentScore, error) {
	rows, err := r.q.GetStudentScores(ctx, courseID)
	if err != nil {
		return nil, err
	}
	scores := make([]domain.StudentScore, len(rows))
	for i, row := range rows {
		scores[i] = domain.StudentScore{
			UserID:   row.UserID,
			Name:     row.Name,
			Rank:     row.Rank,
			AvgScore: row.AvgScore,
		}
	}
	return scores, nil
}

func (r *analyticsRepo) GetOverview(ctx context.Context, instructorID uuid.UUID) (*domain.Overview, error) {
	row, err := r.q.GetOverview(ctx, instructorID)
	if err != nil {
		return nil, err
	}
	avgScore, err := toFloat64(row.AvgScore)
	if err != nil {
		return nil, fmt.Errorf("avg_score: %w", err)
	}
	return &domain.Overview{
		TotalStudents: int(row.TotalStudents),
		TotalCourses:  int(row.TotalCourses),
		AvgScore:      avgScore,
	}, nil
}

func toFloat64(v interface{}) (float64, error) {
	switch val := v.(type) {
	case float64:
		return val, nil
	case int64:
		return float64(val), nil
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unexpected type %T", v)
	}
}

func (r *analyticsRepo) GetStudentAnalytics(ctx context.Context, userID uuid.UUID) (*domain.StudentAnalytics, error) {
	res := &domain.StudentAnalytics{}
	res.CourseProgress = []domain.StudentCourseProgressItem{}
	res.Leaderboard = []domain.LeaderboardEntry{}

	// 1. User profile
	err := r.pool.QueryRow(ctx, `SELECT name, rank, enrollment_id FROM users WHERE id = $1 AND role = 'student'`, userID).
		Scan(&res.UserProfile.Name, &res.UserProfile.Rank, &res.UserProfile.EnrollmentID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrNotFound
		}
		return nil, fmt.Errorf("scan user profile: %w", err)
	}

	// 2. Course Progress Items
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.title, c.rank,
			(SELECT COUNT(*) FROM lessons WHERE course_id = c.id) AS total_lessons,
			(SELECT COUNT(*) FROM user_lesson_progress ulp JOIN lessons l ON l.id = ulp.lesson_id WHERE l.course_id = c.id AND ulp.user_id = $1 AND ulp.is_completed = true) AS completed_lessons,
			COALESCE((SELECT is_completed FROM user_course_progress WHERE user_id = $1 AND course_id = c.id), false) AS is_completed,
			COALESCE((SELECT AVG(a.score) FROM attempts a JOIN quizzes q ON q.id = a.quiz_id WHERE q.course_id = c.id AND a.user_id = $1 AND a.ended_at IS NOT NULL), 0)::float AS quiz_avg
		FROM courses c
		WHERE c.published = true
		ORDER BY c.created_at ASC
	`, userID)
	if err == nil {
		defer rows.Close()
		var totalLessonsCompleted int
		var sumProgressPct int

		for rows.Next() {
			var item domain.StudentCourseProgressItem
			if err := rows.Scan(&item.CourseID, &item.Title, &item.Rank, &item.TotalLessons, &item.CompletedLessons, &item.IsCompleted, &item.QuizAvg); err == nil {
				if item.TotalLessons > 0 {
					item.ProgressPct = int(float64(item.CompletedLessons) / float64(item.TotalLessons) * 100)
					if item.ProgressPct > 100 {
						item.ProgressPct = 100
					}
				} else {
					item.ProgressPct = 0
				}
				if item.IsCompleted {
					res.Stats.CoursesCompleted++
				}
				totalLessonsCompleted += item.CompletedLessons
				sumProgressPct += item.ProgressPct
				res.CourseProgress = append(res.CourseProgress, item)
			}
		}
		res.Stats.CoursesEnrolled = len(res.CourseProgress)
		res.Stats.LessonsCompleted = totalLessonsCompleted
		if res.Stats.CoursesEnrolled > 0 {
			res.Stats.AvgLessonsPerCourse = float64(totalLessonsCompleted) / float64(res.Stats.CoursesEnrolled)
			res.Stats.OverallCompletionPct = sumProgressPct / res.Stats.CoursesEnrolled
		}
	}

	// 3. User's attempts & quiz averages (separated into lesson vs course quizzes)
	attemptRows, err := r.pool.Query(ctx, `
		SELECT a.score, q.lesson_id, q.course_id, c.title AS course_title, COALESCE(l.title, '') AS lesson_title
		FROM attempts a
		JOIN quizzes q ON q.id = a.quiz_id
		JOIN courses c ON c.id = q.course_id
		LEFT JOIN lessons l ON l.id = q.lesson_id
		WHERE a.user_id = $1 AND a.ended_at IS NOT NULL
		ORDER BY a.started_at DESC
	`, userID)
	if err == nil {
		defer attemptRows.Close()
		type scoreAggr struct {
			title string
			sum   float64
			count int
		}
		courseScores := make(map[uuid.UUID]*scoreAggr)
		lessonScores := make(map[string]*scoreAggr)

		var totalScore float64
		var lessonScoreSum float64
		var lessonScoreCount int
		var courseScoreSum float64
		var courseScoreCount int

		for attemptRows.Next() {
			var score float64
			var lessonID pgtype.UUID
			var courseID uuid.UUID
			var courseTitle, lessonTitle string

			if err := attemptRows.Scan(&score, &lessonID, &courseID, &courseTitle, &lessonTitle); err == nil {
				res.Stats.QuizzesAttempted++
				totalScore += score

				if lessonID.Valid {
					lessonScoreSum += score
					lessonScoreCount++
					key := lessonTitle + "|" + courseTitle
					if _, ok := lessonScores[key]; !ok {
						lessonScores[key] = &scoreAggr{title: lessonTitle, sum: score, count: 1}
					} else {
						lessonScores[key].sum += score
						lessonScores[key].count++
					}
				} else {
					courseScoreSum += score
					courseScoreCount++
				}

				if _, ok := courseScores[courseID]; !ok {
					courseScores[courseID] = &scoreAggr{title: courseTitle, sum: score, count: 1}
				} else {
					courseScores[courseID].sum += score
					courseScores[courseID].count++
				}
			}
		}

		if res.Stats.QuizzesAttempted > 0 {
			res.Stats.OverallAvgScore = totalScore / float64(res.Stats.QuizzesAttempted)
		}
		if lessonScoreCount > 0 {
			res.Stats.LessonQuizAvg = lessonScoreSum / float64(lessonScoreCount)
		}
		if courseScoreCount > 0 {
			res.Stats.CourseQuizAvg = courseScoreSum / float64(courseScoreCount)
		}

		// Calculate Strongest & Weakest course
		var maxCourseScore, minCourseScore float64 = -1, 101
		for _, aggr := range courseScores {
			avg := aggr.sum / float64(aggr.count)
			if avg > maxCourseScore {
				maxCourseScore = avg
				res.Spotlight.StrongestCourse = &domain.StrengthWeaknessCourse{Title: aggr.title, AvgScore: avg}
			}
			if avg < minCourseScore {
				minCourseScore = avg
				res.Spotlight.WeakestCourse = &domain.StrengthWeaknessCourse{Title: aggr.title, AvgScore: avg}
			}
		}

		// Calculate Strongest & Weakest lesson
		var maxLessonScore, minLessonScore float64 = -1, 101
		for key, aggr := range lessonScores {
			avg := aggr.sum / float64(aggr.count)
			parts := strings.Split(key, "|")
			courseTitle := ""
			if len(parts) > 1 {
				courseTitle = parts[1]
			}
			if avg > maxLessonScore {
				maxLessonScore = avg
				res.Spotlight.StrongestLesson = &domain.StrengthWeaknessLesson{Title: aggr.title, CourseTitle: courseTitle, Score: avg}
			}
			if avg < minLessonScore {
				minLessonScore = avg
				res.Spotlight.WeakestLesson = &domain.StrengthWeaknessLesson{Title: aggr.title, CourseTitle: courseTitle, Score: avg}
			}
		}
	}

	// Calculate Readiness Score (60% weight on Quiz Avg, 40% on Completion %)
	res.UserProfile.ReadinessScore = (res.Stats.OverallAvgScore * 0.6) + (float64(res.Stats.OverallCompletionPct) * 0.4)

	// 4. Category Leaderboard
	lbRows, err := r.pool.Query(ctx, `
		SELECT
			u.id,
			u.name,
			u.enrollment_id,
			u.rank,
			COALESCE(ucp.completed_count, 0) AS courses_completed,
			COALESCE(ulp.completed_count, 0) AS lessons_completed,
			COALESCE(att.avg_score, 0)::float AS avg_score,
			(SELECT COUNT(*) FROM courses WHERE published = true AND (rank = u.rank OR rank = 'officer')) AS courses_enrolled
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS completed_count
			FROM user_course_progress
			WHERE is_completed = true
			GROUP BY user_id
		) ucp ON ucp.user_id = u.id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS completed_count
			FROM user_lesson_progress
			WHERE is_completed = true
			GROUP BY user_id
		) ulp ON ulp.user_id = u.id
		LEFT JOIN (
			SELECT user_id, AVG(score) AS avg_score
			FROM attempts
			WHERE ended_at IS NOT NULL
			GROUP BY user_id
		) att ON att.user_id = u.id
		WHERE u.rank = $1 AND u.role = 'student'
		ORDER BY avg_score DESC, courses_completed DESC, u.name ASC
	`, res.UserProfile.Rank)
	if err == nil {
		defer lbRows.Close()
		rankPos := 1
		for lbRows.Next() {
			var entry domain.LeaderboardEntry
			if err := lbRows.Scan(
				&entry.UserID,
				&entry.Name,
				&entry.EnrollmentID,
				&entry.Rank,
				&entry.CoursesCompleted,
				&entry.LessonsCompleted,
				&entry.AvgScore,
				&entry.CoursesEnrolled,
			); err == nil {
				compPct := 0.0
				if entry.CoursesEnrolled > 0 {
					compPct = (float64(entry.CoursesCompleted) / float64(entry.CoursesEnrolled)) * 100.0
					if compPct > 100.0 {
						compPct = 100.0
					}
				}
				entry.ReadinessScore = (entry.AvgScore * 0.6) + (compPct * 0.4)
				entry.RankPosition = rankPos
				entry.IsCurrentUser = (entry.UserID == userID)
				res.Leaderboard = append(res.Leaderboard, entry)
				rankPos++
			}
		}
	}

	return res, nil
}

func (r *analyticsRepo) GetInstructorAnalytics(ctx context.Context, instructorID uuid.UUID) (*domain.InstructorAnalytics, error) {
	res := &domain.InstructorAnalytics{}
	res.Courses = []domain.InstructorCourseItem{}
	res.RecentActivity = []domain.RecentActivityItem{}

	// 1. Instructor Courses
	rows, err := r.pool.Query(ctx, `
		SELECT c.id, c.title, c.published,
			(SELECT COUNT(*) FROM lessons WHERE course_id = c.id) AS total_lessons,
			(SELECT COUNT(DISTINCT a.user_id) FROM attempts a JOIN quizzes q ON q.id = a.quiz_id JOIN users u ON u.id = a.user_id WHERE q.course_id = c.id AND u.role = 'student') AS enrolled_students,
			(SELECT COUNT(DISTINCT ucp.user_id) FROM user_course_progress ucp JOIN users u ON u.id = ucp.user_id WHERE ucp.course_id = c.id AND ucp.is_completed = true AND u.role = 'student') AS completed_students,
			COALESCE((SELECT AVG(a.score) FROM attempts a JOIN quizzes q ON q.id = a.quiz_id JOIN users u ON u.id = a.user_id WHERE q.course_id = c.id AND a.ended_at IS NOT NULL AND q.lesson_id IS NOT NULL AND u.role = 'student'), 0)::float AS lesson_quiz_avg,
			COALESCE((SELECT AVG(a.score) FROM attempts a JOIN quizzes q ON q.id = a.quiz_id JOIN users u ON u.id = a.user_id WHERE q.course_id = c.id AND a.ended_at IS NOT NULL AND q.lesson_id IS NULL AND u.role = 'student'), 0)::float AS course_quiz_avg
		FROM courses c
		WHERE c.instructor_id = $1
		ORDER BY c.created_at DESC
	`, instructorID)
	if err == nil {
		defer rows.Close()
		var totalCompletionRate float64
		for rows.Next() {
			var item domain.InstructorCourseItem
			if err := rows.Scan(&item.CourseID, &item.Title, &item.Published, &item.TotalLessons, &item.EnrolledStudents, &item.CompletedStudents, &item.LessonQuizAvg, &item.CourseQuizAvg); err == nil {
				if item.EnrolledStudents > 0 {
					item.CompletionRate = float64(item.CompletedStudents) / float64(item.EnrolledStudents) * 100
				}
				totalCompletionRate += item.CompletionRate
				res.Stats.TotalGraduates += item.CompletedStudents
				res.Courses = append(res.Courses, item)
			}
		}
		res.Stats.TotalCourses = len(res.Courses)
		if res.Stats.TotalCourses > 0 {
			res.Stats.AvgCompletionRate = totalCompletionRate / float64(res.Stats.TotalCourses)
		}
	}

	// 2. Distinct active students across all instructor's courses
	_ = r.pool.QueryRow(ctx, `
		SELECT COUNT(DISTINCT a.user_id)
		FROM attempts a
		JOIN quizzes q ON q.id = a.quiz_id
		JOIN courses c ON c.id = q.course_id
		JOIN users u ON u.id = a.user_id
		WHERE c.instructor_id = $1 AND u.role = 'student'
	`, instructorID).Scan(&res.Stats.TotalStudentsActive)

	// 3. Cohort averages
	_ = r.pool.QueryRow(ctx, `
		SELECT
			COALESCE(AVG(a.score), 0)::float AS overall_avg,
			COALESCE(AVG(CASE WHEN q.lesson_id IS NOT NULL THEN a.score END), 0)::float AS lesson_avg,
			COALESCE(AVG(CASE WHEN q.lesson_id IS NULL THEN a.score END), 0)::float AS course_avg
		FROM attempts a
		JOIN quizzes q ON q.id = a.quiz_id
		JOIN courses c ON c.id = q.course_id
		JOIN users u ON u.id = a.user_id
		WHERE c.instructor_id = $1 AND a.ended_at IS NOT NULL AND u.role = 'student'
	`, instructorID).Scan(&res.Stats.OverallAvgQuizScore, &res.Stats.CohortLessonQuizAvg, &res.Stats.CohortCourseQuizAvg)

	// 4. Recent activity
	actRows, err := r.pool.Query(ctx, `
		SELECT u.name, u.rank, u.enrollment_id, c.title,
			CASE WHEN q.lesson_id IS NOT NULL THEN 'Lesson Quiz' ELSE 'Course Assessment' END AS quiz_type,
			a.score, a.ended_at
		FROM attempts a
		JOIN quizzes q ON q.id = a.quiz_id
		JOIN courses c ON c.id = q.course_id
		JOIN users u ON u.id = a.user_id
		WHERE c.instructor_id = $1 AND a.ended_at IS NOT NULL AND u.role = 'student'
		ORDER BY a.ended_at DESC
		LIMIT 5
	`, instructorID)
	if err == nil {
		defer actRows.Close()
		for actRows.Next() {
			var act domain.RecentActivityItem
			var endedAt pgtype.Timestamptz
			if err := actRows.Scan(&act.StudentName, &act.Rank, &act.EnrollmentID, &act.CourseTitle, &act.QuizType, &act.Score, &endedAt); err == nil {
				if endedAt.Valid {
					act.Date = endedAt.Time
				}
				res.RecentActivity = append(res.RecentActivity, act)
			}
		}
	}

	// 5. All students directory & live rankings
	res.Students = []domain.StudentDirectoryItem{}
	stdRows, err := r.pool.Query(ctx, `
		SELECT
			u.id,
			u.name,
			u.rank,
			u.enrollment_id,
			COALESCE(ucp.completed_count, 0) AS courses_completed,
			COALESCE(ulp.completed_count, 0) AS lessons_completed,
			COALESCE(att.avg_score, 0)::float AS avg_score,
			(SELECT COUNT(*) FROM courses WHERE published = true AND (rank = u.rank OR rank = 'officer')) AS courses_enrolled
		FROM users u
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS completed_count
			FROM user_course_progress
			WHERE is_completed = true
			GROUP BY user_id
		) ucp ON ucp.user_id = u.id
		LEFT JOIN (
			SELECT user_id, COUNT(*) AS completed_count
			FROM user_lesson_progress
			WHERE is_completed = true
			GROUP BY user_id
		) ulp ON ulp.user_id = u.id
		LEFT JOIN (
			SELECT user_id, AVG(score) AS avg_score
			FROM attempts
			WHERE ended_at IS NOT NULL
			GROUP BY user_id
		) att ON att.user_id = u.id
		WHERE u.role = 'student'
		ORDER BY avg_score DESC, courses_completed DESC, u.name ASC
	`)
	if err == nil {
		defer stdRows.Close()
		for stdRows.Next() {
			var item domain.StudentDirectoryItem
			if err := stdRows.Scan(
				&item.ID,
				&item.Name,
				&item.Rank,
				&item.EnrollmentID,
				&item.CoursesCompleted,
				&item.LessonsCompleted,
				&item.AvgScore,
				&item.CoursesEnrolled,
			); err == nil {
				compPct := 0.0
				if item.CoursesEnrolled > 0 {
					compPct = (float64(item.CoursesCompleted) / float64(item.CoursesEnrolled)) * 100.0
					if compPct > 100.0 {
						compPct = 100.0
					}
				}
				item.ReadinessScore = (item.AvgScore * 0.6) + (compPct * 0.4)
				res.Students = append(res.Students, item)
			}
		}
	}

	return res, nil
}
