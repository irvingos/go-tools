package kafka

import "context"

type ConsumerRunner struct {
	name     string
	consumer Consumer
	handler  MessageHandler
}

func NewConsumerRunner(name string, consumer Consumer, handler MessageHandler) *ConsumerRunner {
	return &ConsumerRunner{name: name, consumer: consumer, handler: handler}
}

func (r *ConsumerRunner) Run(ctx context.Context) error {
	return r.consumer.Run(ctx, r.handler)
}

func (r *ConsumerRunner) Name() string {
	return r.name
}
