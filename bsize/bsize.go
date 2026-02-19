// Package bsize provides typed bytes/size handling and parsing, modeled after time.Duration.
package bsize

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/spf13/pflag"
)

const (
	B   Bytes = 1
	KiB Bytes = 1 << 10
	MiB Bytes = 1 << 20
	GiB Bytes = 1 << 30
	TiB Bytes = 1 << 40
	PiB Bytes = 1 << 50
)

var _ pflag.Value = (*Bytes)(nil)

type Bytes uint64

func (b *Bytes) UnmarshalText(s []byte) error {
	n, err := Parse(string(s))
	if err != nil {
		return err
	}
	*b = n
	return nil
}

func (b *Bytes) Set(s string) error {
	return b.UnmarshalText([]byte(s))
}

func (b *Bytes) Type() string {
	return "bytes"
}

var parseBytesRe = regexp.MustCompile(`^(\d+)([KMGTP]iB|B)?$`)

// Parse parses a string of bytes size with an optional unit suffix (e.g. 1024 or 4KiB) into a number of bytes.
// Not safe against overflows.
func Parse(s string) (Bytes, error) {
	units := map[string]Bytes{"": B, "B": B, "KiB": KiB, "MiB": MiB, "GiB": GiB, "TiB": TiB, "PiB": PiB}
	m := parseBytesRe.FindStringSubmatch(s)
	if len(m) != 3 {
		return 0, fmt.Errorf("input does not match format: %s", parseBytesRe.String())
	}
	n, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return Bytes(n) * units[m[2]], nil // note regex captures only keys of units
}

// String returns a string representing the byte size in the form "1.5GiB". The fractional digit
// and decimal point are omitted when the value is an even multiple of the unit. Sizes below 1KiB
// format as plain bytes. The zero size formats as 0B.
func (b Bytes) String() string {
	if b < KiB {
		return strconv.FormatUint(uint64(b), 10) + "B"
	}
	for _, u := range [...]struct {
		size Bytes
		name string
	}{
		{KiB, "KiB"}, {MiB, "MiB"}, {GiB, "GiB"}, {TiB, "TiB"},
	} {
		if b < u.size*1024 {
			return fmtUnit(uint64(b), uint64(u.size)) + u.name
		}
	}
	return fmtUnit(uint64(b), uint64(PiB)) + "PiB"
}

// fmtUnit formats v/unit with at most one decimal digit, taken from go stdlib's time package
func fmtUnit(v, unit uint64) string {
	whole := v / unit
	tenths := v % unit * 10 / unit
	if tenths == 0 {
		return strconv.FormatUint(whole, 10)
	}
	return strconv.FormatUint(whole, 10) + "." + strconv.FormatUint(tenths, 10)
}
