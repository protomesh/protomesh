package redis

import (
	"context"

	"github.com/protomesh/protomesh/pkg/pubsub"
	"github.com/redis/go-redis/v9"
)

type RedisListener interface {
	Close() error
}

type RedisPubSubDriver interface {
	Listen(ctx context.Context, blockingPublish bool, topics ...string) RedisListener
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

func (r *redisPubSubDriver[T]) Listen(ctx context.Context, blockingPublish bool, topics ...string) RedisListener {

	redisPubsub := r.rdb.Subscribe(ctx, topics...)

	ch := redisPubsub.Channel()

	go func(ctx context.Context, ch <-chan *redis.Message) {
		for {
			select {
			case msg := <-ch:
				r.pub.Publish(blockingPublish, r.deserializer(msg))
			case <-ctx.Done():
				return
			}
		}
	}(ctx, ch)

	return redisPubsub

}
