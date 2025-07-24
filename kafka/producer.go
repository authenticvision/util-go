package kafka

import (
	"github.com/IBM/sarama"
	"github.com/authenticvision/util-go/kafka/murmur2"
)

type ProducerConfig struct {
	Brokers []string `required:"1"`
	Topic   string
}

type Producer struct {
	Producer sarama.SyncProducer
	Topic    string
}

func (k Kafka) NewProducer(config ProducerConfig) (*Producer, error) {
	c := sarama.NewConfig()
	c.ClientID = k.ClientID
	c.Version = sarama.V3_0_0_0

	c.Producer.MaxMessageBytes = 16 * 1024 * 1024
	c.Producer.Compression = sarama.CompressionZSTD // using default mode "pretty fast"
	c.Producer.Partitioner = murmur2.Partitioner
	c.Producer.Return.Successes = true // must be set to true for sync. producer
	c.Producer.Return.Errors = true    // must be set to true for sync. producer
	c.Producer.Retry.Max = k.RetryMax
	c.Producer.Retry.Backoff = k.RetryBackoff

	producer, err := sarama.NewSyncProducer(config.Brokers, c)
	if err != nil {
		return nil, err
	}

	topic := k.Topic
	if config.Topic != "" {
		topic = config.Topic
	}

	return &Producer{
		Producer: producer,
		Topic:    topic,
	}, nil
}
