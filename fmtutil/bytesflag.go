package fmtutil

import (
	"github.com/BooleanCat/go-functional/v2/it/op"
	"github.com/spf13/pflag"
)

var _ pflag.Value = op.Ref(Bytes(0))

type Bytes uint64

func (b *Bytes) UnmarshalText(s []byte) error {
	n, err := ParseBytes(string(s))
	if err != nil {
		return err
	}
	*b = Bytes(n)
	return nil
}

func (b *Bytes) Set(s string) error {
	return b.UnmarshalText([]byte(s))
}

func (b *Bytes) Type() string {
	return "bytes"
}

func (b *Bytes) Bytes() uint64 {
	return uint64(*b)
}

func (b *Bytes) String() string {
	return FormatBytes(uint64(*b))
}
