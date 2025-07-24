package fmtutil

import (
	"testing"
)

func TestParseBytes(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    uint64
		wantErr bool
	}{
		{name: "invalid test", arg: "asdf", wantErr: true},
		{name: "empty string", arg: "", wantErr: true},
		{name: "zero", arg: "0", want: 0},
		{name: "one", arg: "1", want: 1},
		{name: "1B", arg: "1B", want: 1},
		{name: "1KiB", arg: "1KiB", want: 1024},
		{name: "1TiB", arg: "1TiB", want: 1024 * 1024 * 1024 * 1024},
		{name: "MAXINT+1", arg: "18446744073709551616B", wantErr: true},
		{name: "MAXINT/1024+1KiB (overflow)", arg: "18014398509481985KiB", want: 1024},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseBytes(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseBytes() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("ParseBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		name string
		arg  uint64
		want string
	}{
		{name: "0B", want: "0B", arg: 0},
		{name: "1B", want: "1B", arg: 1},
		{name: "1KiB", want: "1.0KiB", arg: 1024},
		{name: "1337", want: "1.3KiB", arg: 1337},
		{name: "1TiB", want: "1.0TiB", arg: 1024 * 1024 * 1024 * 1024},
		{name: "MAXINT", want: "16777216.0TiB", arg: 18446744073709551615},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatBytes(tt.arg)
			if got != tt.want {
				t.Errorf("FormatBytes() got = %v, want %v", got, tt.want)
			}
		})
	}
}
