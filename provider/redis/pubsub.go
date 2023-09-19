package redis

import (
	"context"
	"time"

	"github.com/protomesh/go-app"
	"github.com/protomesh/protomesh/pkg/pubsub"
	"github.com/redis/go-redis/v9"
)

type RedisListener interface {
	Close() error
}

type RedisPubSubDriver interface {
	Listen(ctx context.Context, log app.Logger, blockingPublish bool, topics ...string) RedisListener
}

type RedisPubSubDeserializer[T any] func(*redis.Message) *pubsub.Message[T]

type redisPubSubDriver[T any] struct {
	rdb          redis.UniversalClient
	pub          pubsub.PubSubPublisher[T]
	deserializer RedisPubSubDeserializer[T]
}

func DefaultRedisPubSubDeserializer() RedisPubSubDeserializer[string] {
	return func(msg *redis.Message) *pubsub.Message[string] {
		return &pubsub.Message[string]{Topic: msg.Channel, Payload: msg.Payload}
	}
}

func NewRedisPubSubDriver[T any](rdb redis.UniversalClient, pub pubsub.PubSubPublisher[T], deserializer RedisPubSubDeserializer[T]) RedisPubSubDriver {
	return &redisPubSubDriver[T]{
		rdb:          rdb,
		pub:          pub,
		deserializer: deserializer,
	}
}

func (r *redisPubSubDriver[T]) Listen(ctx context.Context, log app.Logger, blockingPublish bool, topics ...string) RedisListener {

	redisPubsub := r.rdb.PSubscribe(ctx, topics...)

	v, err := redisPubsub.Receive(ctx)
	if err != nil {
		log.Panic("Redis pubsub listener failed", "err", err)
	}

	switch v := v.(type) {
	case redis.Subscription:
		log.Info("Redis pubsub listener started", "channel", v.Channel, "kind", v.Kind, "count", v.Count)

	}

	ch := redisPubsub.Channel(redis.WithChannelSendTimeout(60 * time.Second))

	go func(ctx context.Context, ch <-chan *redis.Message) {
		for {
			select {
			case msg := <-ch:
				log.Debug("Redis pubsub message received", "topic", msg.Channel, "payload", msg.Payload)
				r.pub.Publish(blockingPublish, r.deserializer(msg))
				log.Debug("Redis pubsub message published", "topic", msg.Channel, "payload", msg.Payload)
			case <-ctx.Done():
				log.Debug("Redis pubsub listener stopped", "err", ctx.Err())
				return
			}
		}
	}(ctx, ch)

	return redisPubsub

}
