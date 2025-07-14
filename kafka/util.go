package kafka

import "github.com/IBM/sarama"

type ConsumerFunc func(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error

func (c ConsumerFunc) Setup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c ConsumerFunc) Cleanup(session sarama.ConsumerGroupSession) error {
	return nil
}

func (c ConsumerFunc) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	return c(session, claim)
}
