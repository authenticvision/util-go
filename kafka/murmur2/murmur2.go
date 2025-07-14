package murmur2

import (
	"encoding/binary"
)

type KafkaMurmur struct {
	v uint32
}

func (m *KafkaMurmur) Write(p []byte) (int, error) {
	if m.v != 0 {
		// Kafka 1.0's Murmur2 partitioner wisely xors the initial seed against the input's length,
		// which makes appending data impossible without recomputing the entire checksum.
		// This panic should be impossible to hit since Sarama always passes the entire key buffer
		// into Write in one go, and has no reason not to do that in future versions.
		panic("expected a single KafkaMurmur.Write call from Sarama")
	}
	m.v = murmur2(p)
	return len(p), nil
}

func (m *KafkaMurmur) Reset() {
	m.v = 0
}

func (m *KafkaMurmur) Size() int {
	return 4
}

func (m *KafkaMurmur) BlockSize() int {
	return 4
}

// Sum is not called by Samara. It exists for completing the hash.Hash32 interface only.
func (m *KafkaMurmur) Sum(in []byte) []byte {
	var out [4]byte
	binary.BigEndian.PutUint32(out[:], m.Sum32())
	return append(in, out[:]...)
}

func (m *KafkaMurmur) Sum32() uint32 {
	return m.v
}

// murmur2 implements the hashing algorithm used by JVM clients for Kafka.
// Derived from https://github.com/burdiyan/kafkautil/blob/eaf83ed/partitioner.go (2019-01-31)
// Original implementation:
// https://github.com/apache/kafka/blob/1.0.0/clients/src/main/java/org/apache/kafka/common/utils/Utils.java#L353
func murmur2(data []byte) uint32 {
	const seed = uint32(0x9747b28c)
	const m = uint32(0x5bd1e995)
	const r = 24

	n := len(data)
	h := seed ^ uint32(n)
	length4 := n / 4

	for i := 0; i < length4; i++ {
		i4 := i * 4
		k := uint32(data[i4+0]&0xff) + (uint32(data[i4+1]&0xff) << 8) +
			(uint32(data[i4+2]&0xff) << 16) + (uint32(data[i4+3]&0xff) << 24)
		k *= m
		k ^= k >> r
		k *= m
		h *= m
		h ^= k
	}

	switch n % 4 {
	case 3:
		h ^= uint32(data[(n & ^3)+2]&0xff) << 16
		fallthrough
	case 2:
		h ^= uint32(data[(n & ^3)+1]&0xff) << 8
		fallthrough
	case 1:
		h ^= uint32(data[n & ^3] & 0xff)
		h *= m
	}

	h ^= h >> 13
	h *= m
	h ^= h >> 15

	return h
}
