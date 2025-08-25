package murmur2

import (
	"hash"

	"github.com/IBM/sarama"
)

var Partitioner = sarama.NewCustomPartitioner(
	sarama.WithAbsFirst(),
	sarama.WithCustomHashFunction(func() hash.Hash32 { return &KafkaMurmur{} }),
)
