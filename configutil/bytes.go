package configutil

import (
	"git.avdev.at/dev/util/fmtutil"
)

type Bytes uint64

func (b *Bytes) SetValue(s string) (err error) {
	var n uint64
	n, err = fmtutil.ParseBytes(s)
	if err == nil {
		*b = Bytes(n)
	}
	return
}

func (b *Bytes) Value() uint64 {
	return uint64(*b)
}
