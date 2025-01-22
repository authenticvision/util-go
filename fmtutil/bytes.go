package fmtutil

import (
	"fmt"
	"regexp"
	"strconv"
)

// ParseBytes parses a string of bytes size with an optional unit suffix (e.g. 1024 or 4KiB) into a number of bytes.
// Not safe against overflows.
func ParseBytes(s string) (uint64, error) {
	units := map[string]uint64{"": 1, "B": 1, "KiB": 1024, "MiB": 1024 * 1024, "GiB": 1024 * 1024 * 1024, "TiB": 1024 * 1024 * 1024 * 1024}
	re := regexp.MustCompile(`^(\d+)([KMGT]iB|B)?$`)
	m := re.FindStringSubmatch(s)
	if len(m) != 3 {
		return 0, fmt.Errorf("input does not match format: %s", re.String())
	}
	n, err := strconv.ParseUint(m[1], 10, 64)
	if err != nil {
		return 0, err
	}
	return n * units[m[2]], nil
}

func FormatBytes(n uint64) string {
	if n < 1024 {
		return fmt.Sprintf("%dB", n)
	}
	f := float64(n) / 1024
	units := []string{"KiB", "MiB", "GiB", "TiB"}
	for i, unit := range units {
		if f < 1024 || i == len(units)-1 {
			return fmt.Sprintf("%.1f%s", f, unit)
		}
		f /= 1024
	}
	panic("unreachable")
}
