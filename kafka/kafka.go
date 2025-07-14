package kafka

import "time"

type Kafka struct {
	Topic         string
	InitialOffset int64
	RetryMax      int
	RetryBackoff  time.Duration
	ClientID      string
}
