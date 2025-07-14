package murmur2

import (
	"testing"
)

func TestMurmur2(t *testing.T) {
	// Test cases are generated offline using JVM Kafka client for version 1.0.0.
	// Imported from https://github.com/burdiyan/kafkautil/blob/eaf83ed/partitioner_test.go (2019-01-31)

	cases := []struct {
		Input    []byte
		Expected int32
	}{
		{Input: []byte("21"), Expected: -973932308},
		{Input: []byte("foobar"), Expected: -790332482},
		{Input: []byte{12, 42, 56, 24, 109, 111}, Expected: 274204207},
		{Input: []byte("a-little-bit-long-string"), Expected: -985981536},
		{Input: []byte("a-little-bit-longer-string"), Expected: -1486304829},
		{Input: []byte("lkjh234lh9fiuh90y23oiuhsafujhadof229phr9h19h89h8"), Expected: -58897971},
		{Input: []byte{'a', 'b', 'c'}, Expected: 479470107},
	}

	for _, c := range cases {
		if res := int32(murmur2(c.Input)); res != c.Expected {
			t.Errorf("murmur2(%v) = %d, expected %d", c.Input, res, c.Expected)
		}
	}
}
