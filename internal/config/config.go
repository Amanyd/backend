package config

import (
	"os"
	"strconv"
	"time"

	"github.com/spf13/viper"
	"github.com/subosito/gotenv"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Redis  RedisConfig
	MinIO  MinIOConfig
	RAG    RAGConfig
	JWT    JWTConfig
	NATS   NATSConfig
	Log    LogConfig
}

type ServerConfig struct {
	Host         string        `mapstructure:"SERVER_HOST"`
	Port         int           `mapstructure:"SERVER_PORT"`
	ReadTimeout  time.Duration `mapstructure:"SERVER_READ_TIMEOUT"`
	WriteTimeout time.Duration `mapstructure:"SERVER_WRITE_TIMEOUT"`
}

type DBConfig struct {
	DatabaseURL       string        `mapstructure:"DATABASE_URL"`
	DbMaxOpen         int           `mapstructure:"DB_MAX_OPEN_CONNS"`
	DbMaxIdle         int           `mapstructure:"DB_MAX_IDLE_CONNS"`
	DbConnMaxLifetime time.Duration `mapstructure:"DB_CONN_MAX_LIFETIME"`
}

type RedisConfig struct {
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
}

type NATSConfig struct {
	NATSUrl      string `mapstructure:"NATS_URL"`
	NATSUser     string `mapstructure:"NATS_USER"`
	NATSPassword string `mapstructure:"NATS_PASSWORD"`
}

type MinIOConfig struct {
	MinIOEndpoint  string `mapstructure:"MINIO_ENDPOINT"`
	MinIOAccessKey string `mapstructure:"MINIO_ACCESS_KEY"`
	MinIOSecretKey string `mapstructure:"MINIO_SECRET_KEY"`
	MinIOUseSSL    bool   `mapstructure:"MINIO_USE_SSL"`
	MinIOBucket    string `mapstructure:"MINIO_BUCKET"`
}

type RAGConfig struct {
	RAGBaseUrl       string `mapstructure:"RAG_BASE_URL"`
	RAGInternalToken string `mapstructure:"RAG_INTERNAL_TOKEN"`
}

type JWTConfig struct {
	JWTAccessSecret  string        `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret string        `mapstructure:"JWT_REFRESH_SECRET"`
	JWTAccessTTL     time.Duration `mapstructure:"JWT_ACCESS_TTL"`
	JWTRefreshTTL    time.Duration `mapstructure:"JWT_REFRESH_TTL"`
}

type LogConfig struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

func Load() (*Config, error) {
	// Load .env file into OS environment (gotenv handles quoted URLs correctly).
	// Ignore error — file may not exist in prod where real env vars are set.
	_ = gotenv.Load()

	v := viper.New()

	v.SetDefault("SERVER_HOST", "0.0.0.0")
	v.SetDefault("SERVER_PORT", 8080)
	v.SetDefault("SERVER_READ_TIMEOUT", "15s")
	v.SetDefault("SERVER_WRITE_TIMEOUT", "30s")
	v.SetDefault("DB_MAX_OPEN_CONNS", 25)
	v.SetDefault("DB_MAX_IDLE_CONNS", 10)
	v.SetDefault("DB_CONN_MAX_LIFETIME", "30m")
	v.SetDefault("REDIS_ADDR", "localhost:6379")
	v.SetDefault("REDIS_PASSWORD", "")
	v.SetDefault("REDIS_DB", 0)
	v.SetDefault("NATS_URL", "nats://localhost:4222")
	v.SetDefault("MINIO_ENDPOINT", "localhost:9000")
	v.SetDefault("MINIO_USE_SSL", false)
	v.SetDefault("MINIO_BUCKET", "course-files")
	v.SetDefault("JWT_ACCESS_TTL", 15*time.Minute)
	v.SetDefault("JWT_REFRESH_TTL", 168*time.Hour)
	v.SetDefault("LOG_LEVEL", "debug")

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	// Viper's Unmarshal cannot auto-bind flat env vars (e.g. DATABASE_URL) into
	// nested struct fields (e.g. Config.DB.DatabaseURL). Patch empty fields from
	// OS environment as a reliable fallback.
	envOrDefault(&cfg.Server.Host, "SERVER_HOST", "0.0.0.0")
	envOrDefaultInt(&cfg.Server.Port, "SERVER_PORT", 8080)
	envOrDefaultDuration(&cfg.Server.ReadTimeout, "SERVER_READ_TIMEOUT", 15*time.Second)
	envOrDefaultDuration(&cfg.Server.WriteTimeout, "SERVER_WRITE_TIMEOUT", 30*time.Second)
	envOrDefault(&cfg.DB.DatabaseURL, "DATABASE_URL", "")
	envOrDefault(&cfg.Redis.RedisAddr, "REDIS_ADDR", "localhost:6379")
	envOrDefault(&cfg.Redis.RedisPassword, "REDIS_PASSWORD", "")
	envOrDefault(&cfg.NATS.NATSUrl, "NATS_URL", "nats://localhost:4222")
	envOrDefault(&cfg.MinIO.MinIOEndpoint, "MINIO_ENDPOINT", "localhost:9000")
	envOrDefault(&cfg.MinIO.MinIOAccessKey, "MINIO_ACCESS_KEY", "")
	envOrDefault(&cfg.MinIO.MinIOSecretKey, "MINIO_SECRET_KEY", "")
	envOrDefault(&cfg.MinIO.MinIOBucket, "MINIO_BUCKET", "course-files")
	envOrDefault(&cfg.RAG.RAGBaseUrl, "RAG_BASE_URL", "http://localhost:8000")
	envOrDefault(&cfg.RAG.RAGInternalToken, "RAG_INTERNAL_TOKEN", "")
	envOrDefault(&cfg.JWT.JWTAccessSecret, "JWT_ACCESS_SECRET", "")
	envOrDefault(&cfg.JWT.JWTRefreshSecret, "JWT_REFRESH_SECRET", "")
	envOrDefaultDuration(&cfg.JWT.JWTAccessTTL, "JWT_ACCESS_TTL", 15*time.Minute)
	envOrDefaultDuration(&cfg.JWT.JWTRefreshTTL, "JWT_REFRESH_TTL", 168*time.Hour)
	envOrDefault(&cfg.Log.Level, "LOG_LEVEL", "debug")

	return &cfg, nil
}

// envOrDefault sets *dst from os env if *dst is still zero-valued.
func envOrDefault(dst *string, key, fallback string) {
	if *dst != "" {
		return
	}
	if v := os.Getenv(key); v != "" {
		*dst = v
	} else {
		*dst = fallback
	}
}

func envOrDefaultInt(dst *int, key string, fallback int) {
	if *dst != 0 {
		return
	}
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			*dst = n
			return
		}
	}
	*dst = fallback
}

func envOrDefaultDuration(dst *time.Duration, key string, fallback time.Duration) {
	if *dst != 0 {
		return
	}
	if v := os.Getenv(key); v != "" {
		if d, err := time.ParseDuration(v); err == nil {
			*dst = d
			return
		}
	}
	*dst = fallback
}
