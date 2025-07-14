package murmur2

import (
	"github.com/IBM/sarama"
	"hash"
)

var Partitioner = sarama.NewCustomPartitioner(
	sarama.WithAbsFirst(),
	sarama.WithCustomHashFunction(func() hash.Hash32 { return &KafkaMurmur{} }),
)
