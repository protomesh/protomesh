package pubsub

import (
	"sync"
)

// Message is the struct that wraps the topic and payload.
type Message[T any] struct {
	Topic   string
	Payload T
}

type subscriber[T any] struct {
	id            uint64
	ch            chan *Message[T]
	subscriptions map[string]struct{}
}

func (s *subscriber[T]) ID() uint64 {
	return s.id
}

func (s *subscriber[T]) Stream() <-chan *Message[T] {
	return s.ch
}

func (s *subscriber[T]) Put(message *Message[T], blocking bool) {
	if blocking {
		s.ch <- message
		return
	}

	select {
	case s.ch <- message:
	default:
	}
}

func (s *subscriber[T]) Close() {
	close(s.ch)
}

func (s *subscriber[T]) SubscriptionCount() int {
	return len(s.subscriptions)
}

func (s *subscriber[T]) Subscriptions() []string {
	topics := make([]string, 0, len(s.subscriptions))
	for topic := range s.subscriptions {
		topics = append(topics, topic)
	}
	return topics
}

func (s *subscriber[T]) IsSubscribedTo(topic string) bool {
	_, found := s.subscriptions[topic]
	return found
}

func (s *subscriber[T]) unsubscribe(topic string) {
	delete(s.subscriptions, topic)
}

func (s *subscriber[T]) subscribe(topic string) {
	s.subscriptions[topic] = struct{}{}
}

func newSubscriber[T any](id uint64, listenerBufferSize int) *subscriber[T] {
	return &subscriber[T]{
		id:            id,
		ch:            make(chan *Message[T], listenerBufferSize),
		subscriptions: make(map[string]struct{}),
	}
}

// Subscriber is the interface that wraps the basic subscriber methods.
type Subscriber[T any] interface {
	// ID returns the subscriber ID.
	ID() uint64
	// Stream returns the channel that receives messages.
	Stream() <-chan *Message[T]
	// Put puts a message in the channel.
	// If blocking is true, the function will block until the message is
	// received.
	Put(message *Message[T], blocking bool)
	// Close closes the channel.
	Close()
	// SubscriptionCount returns the number of subscriptions.
	SubscriptionCount() int
	// Subscriptions returns the list of subscriptions.
	Subscriptions() []string
}

type pubSub[T any] struct {
	mu                 sync.RWMutex
	subscriptions      map[string]map[uint64]struct{}
	subscribers        map[uint64]Subscriber[T]
	stream             chan *Message[T]
	listenerBufferSize int
}

// NewPubSub creates a new pubsub instance.
// streamBufferSize is the size of the message buffer from de publisher side.
// listenerBufferSize is the size of the message buffer from the subscriber side.
// blockingSubscription determines if the subscription should block until
// all subscribers receive the message.
func NewPubSub[T any](blockingSubscription bool, streamBufferSize int, listenerBufferSize int) PubSub[T] {
	ps := &pubSub[T]{
		subscriptions:      make(map[string]map[uint64]struct{}),
		subscribers:        make(map[uint64]Subscriber[T]),
		stream:             make(chan *Message[T], streamBufferSize),
		listenerBufferSize: listenerBufferSize,
	}

	go ps.startStreaming(blockingSubscription)

	return ps
}

// PubSubPublisher is the interface that wraps the basic pubsub publisher methods.
type PubSubPublisher[T any] interface {
	Publish(blockingPublish bool, messages ...*Message[T]) int
	Close()
}

// PubSubSubscriber is the interface that wraps the basic pubsub subscriber methods.
type PubSubSubscriber[T any] interface {
	Subscribe(subscriberID uint64, topics ...string) Subscriber[T]
	Unsubscribe(subscriberID uint64, topics ...string)
}

// PubSub is the interface that wraps the basic pubsub methods.
type PubSub[T any] interface {
	Subscribe(subscriberID uint64, topics ...string) Subscriber[T]
	Unsubscribe(subscriberID uint64, topics ...string)
	Publish(blockingPublish bool, messages ...*Message[T]) int
	Close()
}

// startStreaming starts the streaming of messages to subscribers.
// If blockingSubscription is true, the function will block until all
// subscribers receive the message.
func (ps *pubSub[T]) startStreaming(blockingSubscription bool) {

	for message := range ps.stream {

		ps.mu.RLock()

		subscription, found := ps.subscriptions[message.Topic]

		if !found {
			ps.mu.RUnlock()
			continue
		}

		for subscriberID := range subscription {
			subs, found := ps.subscribers[subscriberID]
			if !found {
				continue
			}

			subs.Put(message, blockingSubscription)
		}

		ps.mu.RUnlock()

	}

}

// Subscribe subscribes a subscriber to topics.
func (ps *pubSub[T]) Subscribe(subscriberID uint64, topics ...string) Subscriber[T] {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs, found := ps.subscribers[subscriberID].(*subscriber[T])

	if !found {
		subs = newSubscriber[T](subscriberID, ps.listenerBufferSize)
		ps.subscribers[subscriberID] = subs
	}

	for _, topic := range topics {

		if _, found := ps.subscriptions[topic]; !found {
			ps.subscriptions[topic] = make(map[uint64]struct{})
		}

		ps.subscriptions[topic][subscriberID] = struct{}{}

		subs.subscribe(topic)

	}

	return subs
}

// Unsubscribe unsubscribes a subscriber from a topic.
// If no topics are provided, the subscriber will be unsubscribed from all
// topics.
func (ps *pubSub[T]) Unsubscribe(subscriberID uint64, topics ...string) {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	subs, ok := ps.subscribers[subscriberID].(*subscriber[T])
	if !ok || subs == nil {
		return
	}

	if len(topics) == 0 {
		topics = subs.Subscriptions()
	}

	for _, topic := range topics {

		if subs.IsSubscribedTo(topic) {
			continue
		}

		subs.unsubscribe(topic)

		delete(ps.subscriptions[topic], subscriberID)

		if len(ps.subscriptions[topic]) == 0 {
			delete(ps.subscriptions, topic)
		}

	}

	if subs.SubscriptionCount() == 0 {
		delete(ps.subscribers, subscriberID)
	}

}

// Publish publishes a message to a topic and returns the number of subscribers
// that received the message.
// If blockingPublish is false, the function will return immediately after
// publishing the message.
// If blockingPublish is true, the function will block until all subscribers
// receive the message.
func (ps *pubSub[T]) Publish(blockingPublish bool, messages ...*Message[T]) int {
	ps.mu.RLock()
	defer ps.mu.RUnlock()

	subscribersCount := 0

	for _, message := range messages {

		subscription, found := ps.subscriptions[message.Topic]

		if !found {
			continue
		}

		subscribersCount += len(subscription)

		if blockingPublish {
			ps.stream <- message
			return subscribersCount
		}

		select {
		case ps.stream <- message:
		default:
		}

	}

	return subscribersCount

}

// Close closes the pubsub instance.
func (ps *pubSub[T]) Close() {
	ps.mu.Lock()
	defer ps.mu.Unlock()

	for _, subscriber := range ps.subscribers {
		subscriber.Close()
	}

	close(ps.stream)
}
