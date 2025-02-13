package event

import "context"

type IPublisher interface {
	SendMessage(ctx context.Context, value interface{}) error
}

type ISubscriber interface {
	SubscribeToTopic(ctx context.Context) error
	ConsumeMessages(ctx context.Context, msgTypeConf func() ConsumerMessage) (chMsg <-chan ConsumerMessage, chErr <-chan error, chCommitRequest chan<- bool)
}
type ConsumerMessage interface {
	EventName() string
}
