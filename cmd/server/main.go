package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/Amanyd/backend/internal/config"
	"github.com/Amanyd/backend/internal/handler"
	minioinfra "github.com/Amanyd/backend/internal/infra/minio"
	natsinfra "github.com/Amanyd/backend/internal/infra/nats"
	"github.com/Amanyd/backend/internal/infra/postgres"
	"github.com/Amanyd/backend/internal/infra/postgres/migrations"
	raginfra "github.com/Amanyd/backend/internal/infra/rag"
	redisinfra "github.com/Amanyd/backend/internal/infra/redis"
	tusinfra "github.com/Amanyd/backend/internal/infra/tus"
	"github.com/Amanyd/backend/internal/service"
	"github.com/Amanyd/backend/internal/worker"
	"github.com/Amanyd/backend/pkg/logger"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		panic("load config: " + err.Error())
	}

	log := logger.NewLogger(cfg.Log.Level)
	defer log.Sync()

	// Postgres
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, cfg.DB.DatabaseURL)
	if err != nil {
		log.Fatal("postgres connect", zap.Error(err))
	}
	defer pool.Close()
	log.Info("postgres connected")

	runMigrations(cfg.DB.DatabaseURL, log)

	// Redis
	rdb := redisinfra.NewRedisClient(cfg.Redis)
	defer rdb.Close()
	log.Info("redis connected")

	// NATS
	js, nc, err := natsinfra.NewJetStream(cfg.NATS)
	if err != nil {
		log.Fatal("nats connect", zap.Error(err))
	}
	defer nc.Drain()
	log.Info("nats connected")

	// MinIO
	minioClient, err := minioinfra.NewMinIOClient(cfg.MinIO)
	if err != nil {
		log.Fatal("minio connect", zap.Error(err))
	}
	log.Info("minio connected")

	// Repositories
	userRepo := postgres.NewUserRepo(pool)
	courseRepo := postgres.NewCourseRepo(pool)
	lessonRepo := postgres.NewLessonRepo(pool)
	fileRepo := postgres.NewFileRepo(pool)
	quizRepo := postgres.NewQuizRepo(pool)
	chatRepo := postgres.NewChatRepo(pool)
	analyticsRepo := postgres.NewAnalyticsRepo(pool)

	// Infra adapters
	storage := minioinfra.NewStorage(minioClient, cfg.MinIO.MinIOBucket)
	queue := natsinfra.NewPublisher(js)
	ragClient := raginfra.NewRAGClient(cfg.RAG.RAGBaseUrl, cfg.RAG.RAGInternalToken)
	cache := redisinfra.NewCache(rdb)
	rateLimiter := redisinfra.NewRateLimiter(rdb)

	// TUS handler
	tusH, err := tusinfra.NewTUSHandler(cfg.MinIO, tusinfra.TUSDeps{
		Files:   fileRepo,
		Lessons: lessonRepo,
		Courses: courseRepo,
		Queue:   queue,
		Bucket:  cfg.MinIO.MinIOBucket,
	}, log)
	if err != nil {
		log.Fatal("tus handler", zap.Error(err))
	}

	// Services
	userSvc := service.NewUserService(userRepo, cfg.JWT)
	courseSvc := service.NewCourseService(courseRepo, lessonRepo, cache)
	fileSvc := service.NewFileService(fileRepo, storage, cache)
	chatSvc := service.NewChatService(chatRepo, courseRepo, userRepo, ragClient)
	quizSvc := service.NewQuizService(quizRepo, courseRepo, queue, cache)
	analyticsSvc := service.NewAnalyticsService(analyticsRepo)

	// Handlers
	userH := handler.NewUserHandler(userSvc)
	courseH := handler.NewCourseHandler(courseSvc)
	lessonH := handler.NewLessonHandler(courseSvc)
	fileH := handler.NewFileHandler(fileSvc)
	quizH := handler.NewQuizHandler(quizSvc)
	chatH := handler.NewChatHandler(chatSvc)
	analytH := handler.NewAnalyticsHandler(analyticsSvc)
	healthH := handler.NewHealthHandler()

	router := handler.NewRouter(userH, courseH, lessonH, fileH, quizH, chatH, analytH, healthH, tusH, rateLimiter, cfg, log)

	// Workers
	workerCtx, workerCancel := context.WithCancel(ctx)
	defer workerCancel()

	go func() {
		if err := worker.StartIngestDoneWorker(workerCtx, js, worker.IngestDoneWorkerDeps{
			Files:   fileRepo,
			Lessons: lessonRepo,
			Quizzes: quizRepo,
			Queue:   queue,
		}, log); err != nil {
			log.Error("ingest_done_worker stopped", zap.Error(err))
		}
	}()

	go func() {
		if err := worker.StartQuizDoneWorker(workerCtx, js, worker.QuizDoneWorkerDeps{
			Quizzes: quizRepo,
		}, log); err != nil {
			log.Error("quiz_done_worker stopped", zap.Error(err))
		}
	}()

	// HTTP Server
	srv := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
	}

	// Graceful shutdown
	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigCh
		log.Info("shutting down", zap.String("signal", sig.String()))

		workerCancel()
		srv.Shutdown(context.Background())
	}()

	log.Info("server starting", zap.String("addr", srv.Addr))
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal("server error", zap.Error(err))
	}
	log.Info("server stopped")
}

func runMigrations(dsn string, log *zap.Logger) {
	source, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatal("migration source", zap.Error(err))
	}

	m, err := migrate.NewWithSourceInstance("iofs", source, "pgx5://"+strings.TrimPrefix(dsn, "postgres://"))
	if err != nil {
		log.Fatal("migration init", zap.Error(err))
	}

	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal("migration run", zap.Error(err))
	}
	log.Info("migrations applied")
}
