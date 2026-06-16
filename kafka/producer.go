package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/IBM/sarama"
)

type Producer interface {
	Send(ctx context.Context, message Message) error
	Close() error
}

type Message struct {
	Topic string
	Key   string
	Value []byte
}

type ProducerConfig struct {
	Brokers  string
	User     string
	Password string
}

type producer struct {
	syncProducer sarama.SyncProducer
}

func NewProducer(cfg ProducerConfig) (Producer, error) {
	config, err := newProducerSaramaConfig(cfg)
	if err != nil {
		return nil, err
	}

	brokers := strings.Split(cfg.Brokers, ",")
	syncProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, err
	}

	return &producer{syncProducer: syncProducer}, nil
}

func newProducerSaramaConfig(cfg ProducerConfig) (*sarama.Config, error) {
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	configureSASLSHA512(config, cfg.User, cfg.Password)
	if err := config.Validate(); err != nil {
		return nil, err
	}
	return config, nil
}

func (p *producer) Send(ctx context.Context, message Message) error {
	if strings.TrimSpace(message.Topic) == "" {
		return errors.New("kafka message topic is required")
	}
	if err := ctx.Err(); err != nil {
		return err
	}

	producerMessage := &sarama.ProducerMessage{
		Topic: message.Topic,
		Value: sarama.ByteEncoder(message.Value),
	}
	if message.Key != "" {
		producerMessage.Key = sarama.StringEncoder(message.Key)
	}

	_, _, err := p.syncProducer.SendMessage(producerMessage)
	if err != nil {
		return fmt.Errorf("send kafka message topic=%s: %w", message.Topic, err)
	}
	return nil
}

func (p *producer) Close() error {
	return p.syncProducer.Close()
}
