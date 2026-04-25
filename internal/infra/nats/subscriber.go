package nats

import (
	"context"
	"fmt"

	"github.com/nats-io/nats.go/jetstream"
)

func CreateOrUpdateConsumer(
	ctx context.Context,
	js jetstream.JetStream,
	stream, durable, filterSubject string,
) (jetstream.Consumer, error) {
	cons, err := js.CreateOrUpdateConsumer(ctx, stream, jetstream.ConsumerConfig{
		Durable:       durable,
		FilterSubject: filterSubject,
		AckPolicy:     jetstream.AckExplicitPolicy,
	})
	if err != nil {
		return nil, fmt.Errorf("consumer %s on %s: %w", durable, stream, err)
	}
	return cons, nil
}

// ConsumeLoop blocks until ctx is cancelled, invoking handler for each message.
// Caller is responsible for msg.Ack() / msg.Nak() inside the handler.
func ConsumeLoop(ctx context.Context, cons jetstream.Consumer, handler func(msg jetstream.Msg)) error {
	cc, err := cons.Consume(func(msg jetstream.Msg) {
		handler(msg)
	})
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	<-ctx.Done()
	cc.Stop()
	return nil
}
