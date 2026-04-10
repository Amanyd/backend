package config

import (
	"time"

	"github.com/spf13/viper"
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
	DbMaxOpen         int           `mapstructure:"DB_MAX_OPEN_CONSS"`
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
	JWTAccessSecret  string `mapstructure:"JWT_ACCESS_SECRET"`
	JWTRefreshSecret string `mapstructure:"JWT_REFRESH_SECRET"`
	JWTAccessTTL     string `mapstructure:"JWT_ACCESS_TTL"`
	JWTRefreshTTL    string `mapstructure:"JWT_REFRESH_TTL"`
}

type LogConfig struct {
	Level string `mapstructure:"LOG_LEVEL"`
}

func Load() (*Config, error) {
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
	v.SetDefault("JWT_ACCESS_TTL", "15m")
	v.SetDefault("JWT_REFRESH_TTL", "168h")
	v.SetDefault("LOG_LEVEL", "debug")

	// config file for local dev
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath("./internal/config")
	v.AddConfigPath(".")
	v.ReadInConfig()

	v.AutomaticEnv()

	var cfg Config
	if err := v.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
