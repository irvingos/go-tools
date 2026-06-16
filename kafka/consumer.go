package kafka

import (
	"context"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

type MessageHandler func(ctx context.Context, message []byte) error

type Consumer interface {
	Run(ctx context.Context, handler MessageHandler) error
}

type OffsetMode string

const (
	OffsetModeNewest OffsetMode = "newest" // default
	OffsetModeOldest OffsetMode = "oldest"
)

func (m OffsetMode) ToSaramaOffset() int64 {
	switch m {
	case OffsetModeNewest:
		return sarama.OffsetNewest
	case OffsetModeOldest:
		return sarama.OffsetOldest
	}
	return sarama.OffsetNewest
}

type ConsumerConfig struct {
	Brokers    string
	Group      string
	Topic      string
	User       string
	Password   string
	OffsetMode OffsetMode
}

func NewConsumer(cfg ConsumerConfig) (Consumer, error) {
	config, err := newConsumerSaramaConfig(cfg)
	if err != nil {
		return nil, err
	}

	brokers := strings.Split(cfg.Brokers, ",")
	group, err := sarama.NewConsumerGroup(brokers, cfg.Group, config)
	if err != nil {
		return nil, err
	}

	return &consumer{
		topics: []string{cfg.Topic},
		group:  group,
	}, nil
}

func newConsumerSaramaConfig(cfg ConsumerConfig) (*sarama.Config, error) {
	config := sarama.NewConfig()
	config.Consumer.Return.Errors = true
	config.Consumer.Offsets.Initial = cfg.OffsetMode.ToSaramaOffset()
	configureSASLSHA512(config, cfg.User, cfg.Password)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

type consumer struct {
	topics []string
	group  sarama.ConsumerGroup
}

func (c *consumer) Run(ctx context.Context, handler MessageHandler) error {
	defer c.group.Close()

	h := &groupHandler{handler: handler}
	for {
		if err := c.group.Consume(ctx, c.topics, h); err != nil {
			if ctx.Err() != nil {
				return nil
			}
			return err
		}
		if ctx.Err() != nil {
			return nil
		}
	}
}

type groupHandler struct {
	handler MessageHandler
}

func (h *groupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (h *groupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }

func (h *groupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				return nil
			}
			handlerCtx := context.WithoutCancel(session.Context())
			if err := h.handler(handlerCtx, message.Value); err != nil {
				return fmt.Errorf("handle kafka message topic=%s partition=%d offset=%d: %w",
					message.Topic, message.Partition, message.Offset, err)
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}
