package nats

import (
	"context"

	"github.com/eulerbutcooler/wingman-backend/internal/port"
	"github.com/nats-io/nats.go/jetstream"
)

type publisher struct {
	js jetstream.JetStream
}

func NewPublisher(js jetstream.JetStream) port.MessageQueue {
	return &publisher{js: js}
}

func (p *publisher) Publish(ctx context.Context, subject string, data []byte) error {
	_, err := p.js.Publish(ctx, subject, data)
	return err
}
