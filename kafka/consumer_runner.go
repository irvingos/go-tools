package kafka

import "context"

type ConsumerRunner struct {
	consumer Consumer
	handler  MessageHandler
}

func NewConsumerRunner(consumer Consumer, handler MessageHandler) *ConsumerRunner {
	return &ConsumerRunner{consumer: consumer, handler: handler}
}

func (r *ConsumerRunner) Run(ctx context.Context) error {
	return r.consumer.Run(ctx, r.handler)
}
