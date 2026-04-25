package nats

import (
	"context"
	"fmt"
	"time"

	"github.com/eulerbutcooler/wingman-backend/internal/config"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

func NewJetStream(cfg config.NATSConfig) (jetstream.JetStream, *nats.Conn, error) {
	opts := []nats.Option{
		nats.Name("aeromentor-backend"),
		nats.RetryOnFailedConnect(true),
		nats.MaxReconnects(-1),
		nats.ReconnectWait(2 * time.Second),
	}
	if cfg.NATSUser != "" {
		opts = append(opts, nats.UserInfo(cfg.NATSUser, cfg.NATSPassword))
	}

	nc, err := nats.Connect(cfg.NATSUrl, opts...)
	if err != nil {
		return nil, nil, fmt.Errorf("nats connect: %w", err)
	}

	js, err := jetstream.New(nc)
	if err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("jetstream new: %w", err)
	}

	if err := ensureStreams(context.Background(), js); err != nil {
		nc.Close()
		return nil, nil, fmt.Errorf("ensure streams: %w", err)
	}

	return js, nc, nil
}

func ensureStreams(ctx context.Context, js jetstream.JetStream) error {
	streams := []jetstream.StreamConfig{
		{
			Name:      StreamIngest,
			Subjects:  []string{SubjectIngestRequest},
			Retention: jetstream.WorkQueuePolicy,
		},
		{
			Name:      StreamIngestDone,
			Subjects:  []string{SubjectIngestDone},
			Retention: jetstream.LimitsPolicy,
			MaxAge:    1 * time.Hour,
		},
		{
			Name:      StreamQuiz,
			Subjects:  []string{SubjectQuizRequest},
			Retention: jetstream.WorkQueuePolicy,
		},
		{
			Name:      StreamQuizDone,
			Subjects:  []string{SubjectQuizDone},
			Retention: jetstream.LimitsPolicy,
			MaxAge:    1 * time.Hour,
		},
	}

	for _, cfg := range streams {
		if _, err := js.CreateOrUpdateStream(ctx, cfg); err != nil {
			return fmt.Errorf("stream %s: %w", cfg.Name, err)
		}
	}
	return nil
}
