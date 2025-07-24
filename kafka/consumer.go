package kafka

import (
	"context"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/authenticvision/util-go/logutil"
)

type ConsumerConfig struct {
	Brokers       []string `required:"1"`
	ConsumerGroup string
	Topic         string
}

type Consumer struct {
	ConsumerGroup sarama.ConsumerGroup
	Topic         string
}

func (k Kafka) NewConsumer(config ConsumerConfig) (*Consumer, error) {
	c := sarama.NewConfig()
	c.ClientID = k.ClientID
	c.Version = sarama.V3_0_0_0

	c.Consumer.Offsets.Initial = k.InitialOffset
	// messages have to be marked with MarkMessage anyway, autocommit just commits them automatically
	// Consumer.Consume takes care of that
	c.Consumer.Offsets.AutoCommit.Enable = true
	c.Consumer.Offsets.Retry.Max = k.RetryMax
	c.Consumer.Retry.Backoff = k.RetryBackoff
	c.Consumer.Fetch.Max = 16 * 1024 * 1024
	// since processing of messages is pretty slow anyway, it makes no sense to buffer more than one message internally
	c.ChannelBufferSize = 1

	consumer, err := sarama.NewConsumerGroup(config.Brokers, config.ConsumerGroup, c)
	if err != nil {
		return nil, err
	}

	topic := k.Topic
	if config.Topic != "" {
		topic = config.Topic
	}

	return &Consumer{
		ConsumerGroup: consumer,
		Topic:         topic,
	}, nil
}

// Consume is a helper for consuming from Kafka and provides are more intuitive
// interface than sarama's consumer.Consume(). Mainly this means consumption
// will abort if the consumerFn returns an error.
// This expects autocommit to be enabled. Note that consumerFn will be called
// concurrently (but sequentially for each partition).
func (c *Consumer) Consume(ctx context.Context, consumerFn func(context.Context, *sarama.ConsumerMessage) error) error {
	log := logutil.FromContext(ctx)
	ctx, cancel := context.WithCancelCause(ctx)
	for {
		err := c.ConsumerGroup.Consume(ctx, []string{c.Topic}, ConsumerFunc(func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
			// note: returning errors from this function does effectively nothing unless we also cancel the context
			for {
				select {
				case message, ok := <-claim.Messages():
					if !ok {
						log.Debug("message channel was closed")
						return nil
					}
					log := log.With("kafka_key", string(message.Key))
					ctx = logutil.WithLogContext(ctx, log)
					log.Debug("received message")

					err := consumerFn(ctx, message)
					if err != nil {
						cancel(fmt.Errorf("process message (key %q): %w", string(message.Key), err))
						return nil
					}
					log.Debug("processed message")

					// This only marks a message (actually an offset) as ready
					// to be commited. This will be picked up by autocommit.
					// Alternatively this would require a session.Commit() here.
					session.MarkMessage(message, "")
				case <-session.Context().Done():
					return nil
				}
			}
		}))
		if cause := context.Cause(ctx); cause != nil {
			return fmt.Errorf("consume ctx: %w", cause)
		} else if err != nil {
			return fmt.Errorf("consume: %w", err)
		}
	}
}
