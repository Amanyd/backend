package handler

import (
	"net/http"
	"time"

	"github.com/Amanyd/backend/internal/config"
	"github.com/Amanyd/backend/internal/domain"
	redisinfra "github.com/Amanyd/backend/internal/infra/redis"
	"github.com/Amanyd/backend/internal/service"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"
)

func NewRouter(
	userH *UserHandler,
	courseH *CourseHandler,
	lessonH *LessonHandler,
	fileH *FileHandler,
	quizH *QuizHandler,
	chatH *ChatHandler,
	analytH *AnalyticsHandler,
	healthH *HealthHandler,
	tusH http.Handler,
	rl *redisinfra.RateLimiter,
	cfg *config.Config,
	log *zap.Logger,
) *chi.Mux {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(RequestLogger(log))
	r.Use(middleware.Recoverer)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"http://localhost:3000"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "PATCH", "HEAD", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "Tus-Resumable", "Upload-Length", "Upload-Metadata", "Upload-Offset"},
		ExposedHeaders:   []string{"Link", "Location", "Tus-Resumable", "Upload-Offset", "Upload-Length"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	r.Get("/health", healthH.Health)
	r.Post("/api/v1/auth/register", userH.Register)
	r.Post("/api/v1/auth/login", userH.Login)
	r.Post("/api/v1/auth/refresh", userH.RefreshToken)

	r.Group(func(r chi.Router) {
		r.Use(JWTAuthMiddleware(cfg.JWT.JWTAccessSecret))
		r.Use(RateLimitMiddleware(rl, 60, 10))
		r.Use(middleware.Timeout(30 * time.Second))

		r.Get("/api/v1/users/me", userH.Me)

		r.Get("/api/v1/courses", courseH.List)
		r.Get("/api/v1/courses/{courseId}", courseH.Get)
		r.Get("/api/v1/courses/{courseId}/lessons", lessonH.List)
		r.Get("/api/v1/lessons/{lessonId}/files", fileH.ListByLesson)
		r.Get("/api/v1/files/{fileId}/status", fileH.IngestStatus)
		r.Get("/api/v1/files/{fileId}/view", fileH.ViewURL)

		r.Get("/api/v1/courses/{courseId}/quizzes", quizH.ListByCourse)
		r.Get("/api/v1/quizzes/{quizId}", quizH.Get)
		r.Post("/api/v1/quizzes/{quizId}/attempt", quizH.StartAttempt)
		r.Post("/api/v1/attempts/{attemptId}/answer", quizH.SubmitAnswer)
		r.Post("/api/v1/attempts/{attemptId}/finish", quizH.FinishAttempt)
		r.Get("/api/v1/attempts/{attemptId}/results", quizH.Results)

		r.Get("/api/v1/chat/sessions", chatH.ListSessions)
		r.Post("/api/v1/chat/sessions", chatH.CreateSession)
		r.Get("/api/v1/chat/sessions/{sessionId}/history", chatH.GetHistory)

		r.Group(func(r chi.Router) {
			r.Use(RBACMiddleware(domain.RoleInstructor))

			r.Post("/api/v1/courses", courseH.Create)
			r.Put("/api/v1/courses/{courseId}", courseH.Update)
			r.Post("/api/v1/courses/{courseId}/finalize", courseH.Finalize)
			r.Delete("/api/v1/courses/{courseId}", courseH.Delete)

			r.Post("/api/v1/courses/{courseId}/lessons", lessonH.Create)
			r.Put("/api/v1/lessons/{lessonId}", lessonH.Update)
			r.Delete("/api/v1/lessons/{lessonId}", lessonH.Delete)

			r.Post("/api/v1/quizzes/{quizId}/reset", quizH.Reset)

			r.Get("/api/v1/analytics", analytH.Overview)
			r.Get("/api/v1/analytics/{courseId}", analytH.CourseMetrics)
		})
	})

	// Chat streaming route — outside the Timeout middleware group because
	// SSE streams can run for the full duration of LLM generation (which
	// can exceed 30s with a local model). The context is kept alive by the
	// client connection; cancellation is handled naturally when the client
	// disconnects or the stream ends.
	r.Group(func(r chi.Router) {
		r.Use(JWTAuthMiddleware(cfg.JWT.JWTAccessSecret))
		r.Use(RateLimitMiddleware(rl, 60, 10))
		r.Post("/api/v1/chat/sessions/{sessionId}/message", chatH.SendMessage)
	})

	// TUS upload routes — mounted outside the Timeout middleware group
	// because file uploads can take much longer than 30 seconds.
	// Chi's r.Mount only sets an internal route-context path; it does NOT
	// rewrite r.URL.Path. tusd's internal router reads r.URL.Path directly
	// (strings.Trim(r.URL.Path, "/")) and needs it to be "/" for POST
	// creation. http.StripPrefix rewrites r.URL.Path so tusd sees the
	// correct stripped path. Location headers remain correct because tusd
	// builds them from its config BasePath, not from the request path.
	r.Group(func(r chi.Router) {
		r.Use(JWTAuthMiddleware(cfg.JWT.JWTAccessSecret))
		r.Use(RBACMiddleware(domain.RoleInstructor))
		r.Mount("/api/v1/files/tus", http.StripPrefix("/api/v1/files/tus", tusH))
	})

	return r
}

func NewUserHandler(svc *service.UserService) *UserHandler       { return &UserHandler{svc: svc} }
func NewCourseHandler(svc *service.CourseService) *CourseHandler  { return &CourseHandler{svc: svc} }
func NewLessonHandler(svc *service.CourseService) *LessonHandler  { return &LessonHandler{svc: svc} }
func NewFileHandler(svc *service.FileService) *FileHandler        { return &FileHandler{svc: svc} }
func NewQuizHandler(svc *service.QuizService) *QuizHandler        { return &QuizHandler{svc: svc} }
func NewChatHandler(svc *service.ChatService) *ChatHandler        { return &ChatHandler{svc: svc} }
func NewAnalyticsHandler(svc *service.AnalyticsService) *AnalyticsHandler {
	return &AnalyticsHandler{svc: svc}
}
func NewHealthHandler() *HealthHandler { return &HealthHandler{} }
