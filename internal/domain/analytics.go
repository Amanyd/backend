package domain

import (
	"time"

	"github.com/google/uuid"
)

type EventType string

type Event struct {
	ID        uuid.UUID      `json:"id"`
	UserID    uuid.UUID      `json:"user_id"`
	CourseID  *uuid.UUID     `json:"course_id,omitempty"`
	Type      EventType      `json:"type"`
	Metadata  map[string]any `json:"metadata,omitempty"`
	CreatedAt time.Time      `json:"created_at"`
}

type Metric struct {
	CourseID      uuid.UUID `json:"course_id"`
	TotalStudents int       `json:"total_students"`
	AvgQuizScore  float64   `json:"avg_quiz_score"`
	TotalMessages int       `json:"total_messages"`
	TotalFiles    int       `json:"total_files"`
}

type StudentScore struct {
	UserID   uuid.UUID `json:"user_id"`
	Name     string    `json:"name"`
	Rank     string    `json:"rank"`
	AvgScore float64   `json:"avg_score"`
}

type Overview struct {
	TotalStudents int     `json:"total_students"`
	TotalCourses  int     `json:"total_courses"`
	AvgScore      float64 `json:"avg_score"`
}

type StudentCourseProgressItem struct {
	CourseID         uuid.UUID `json:"course_id"`
	Title            string    `json:"title"`
	Rank             string    `json:"rank"`
	TotalLessons     int       `json:"total_lessons"`
	CompletedLessons int       `json:"completed_lessons"`
	ProgressPct      int       `json:"progress_pct"`
	IsCompleted      bool      `json:"is_completed"`
	QuizAvg          float64   `json:"quiz_avg"`
}

type StrengthWeaknessCourse struct {
	Title    string  `json:"title"`
	AvgScore float64 `json:"avg_score"`
}

type StrengthWeaknessLesson struct {
	Title       string  `json:"title"`
	CourseTitle string  `json:"course_title"`
	Score       float64 `json:"score"`
}

type LeaderboardEntry struct {
	RankPosition     int       `json:"rank_position"`
	UserID           uuid.UUID `json:"user_id"`
	Name             string    `json:"name"`
	EnrollmentID     string    `json:"enrollment_id"`
	Rank             string    `json:"rank"`
	CoursesCompleted int       `json:"courses_completed"`
	AvgScore         float64   `json:"avg_score"`
	IsCurrentUser    bool      `json:"is_current_user"`
}

type StudentUserProfile struct {
	Name           string  `json:"name"`
	Rank           string  `json:"rank"`
	EnrollmentID   string  `json:"enrollment_id"`
	ReadinessScore float64 `json:"readiness_score"`
}

type StudentStats struct {
	CoursesEnrolled      int     `json:"courses_enrolled"`
	CoursesCompleted     int     `json:"courses_completed"`
	OverallCompletionPct int     `json:"overall_completion_pct"`
	LessonsCompleted     int     `json:"lessons_completed"`
	AvgLessonsPerCourse  float64 `json:"avg_lessons_per_course"`
	QuizzesAttempted     int     `json:"quizzes_attempted"`
	OverallAvgScore      float64 `json:"overall_avg_score"`
	LessonQuizAvg        float64 `json:"lesson_quiz_avg"`
	CourseQuizAvg        float64 `json:"course_quiz_avg"`
}

type StudentSpotlight struct {
	StrongestCourse *StrengthWeaknessCourse `json:"strongest_course"`
	WeakestCourse   *StrengthWeaknessCourse `json:"weakest_course"`
	StrongestLesson *StrengthWeaknessLesson `json:"strongest_lesson"`
	WeakestLesson   *StrengthWeaknessLesson `json:"weakest_lesson"`
}

type StudentAnalytics struct {
	UserProfile    StudentUserProfile          `json:"user_profile"`
	Stats          StudentStats                `json:"stats"`
	CourseProgress []StudentCourseProgressItem `json:"course_progress"`
	Spotlight      StudentSpotlight            `json:"spotlight"`
	Leaderboard    []LeaderboardEntry          `json:"leaderboard"`
}

type InstructorCourseItem struct {
	CourseID          uuid.UUID `json:"course_id"`
	Title             string    `json:"title"`
	Published         bool      `json:"published"`
	TotalLessons      int       `json:"total_lessons"`
	EnrolledStudents  int       `json:"enrolled_students"`
	CompletedStudents int       `json:"completed_students"`
	CompletionRate    float64   `json:"completion_rate"`
	LessonQuizAvg     float64   `json:"lesson_quiz_avg"`
	CourseQuizAvg     float64   `json:"course_quiz_avg"`
}

type RecentActivityItem struct {
	StudentName  string    `json:"student_name"`
	Rank         string    `json:"rank"`
	EnrollmentID string    `json:"enrollment_id"`
	CourseTitle  string    `json:"course_title"`
	QuizType     string    `json:"quiz_type"`
	Score        float64   `json:"score"`
	Date         time.Time `json:"date"`
}

type InstructorStats struct {
	TotalCourses        int     `json:"total_courses"`
	TotalStudentsActive int     `json:"total_students_active"`
	TotalGraduates      int     `json:"total_graduates"`
	AvgCompletionRate   float64 `json:"avg_completion_rate"`
	OverallAvgQuizScore float64 `json:"overall_avg_quiz_score"`
	CohortLessonQuizAvg float64 `json:"cohort_lesson_quiz_avg"`
	CohortCourseQuizAvg float64 `json:"cohort_course_quiz_avg"`
}

type StudentDirectoryItem struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Rank             string    `json:"rank"`
	EnrollmentID     string    `json:"enrollment_id"`
	CoursesCompleted int       `json:"courses_completed"`
	CoursesEnrolled  int       `json:"courses_enrolled"`
	LessonsCompleted int       `json:"lessons_completed"`
	AvgScore         float64   `json:"avg_score"`
	ReadinessScore   float64   `json:"readiness_score"`
}

type InstructorAnalytics struct {
	Stats          InstructorStats        `json:"stats"`
	Courses        []InstructorCourseItem `json:"courses"`
	RecentActivity []RecentActivityItem   `json:"recent_activity"`
	Students       []StudentDirectoryItem `json:"students"`
}
