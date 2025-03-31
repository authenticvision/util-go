package configutil

import (
	"git.avdev.at/dev/util/fmtutil"
)

type Bytes uint64

func (b *Bytes) UnmarshalText(s []byte) error {
	n, err := fmtutil.ParseBytes(string(s))
	if err != nil {
		return err
	}
	*b = Bytes(n)
	return nil
}

func (b Bytes) Bytes() uint64 {
	return uint64(b)
}

func (b Bytes) String() string {
	return fmtutil.FormatBytes(uint64(b))
}
