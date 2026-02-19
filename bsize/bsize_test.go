package bsize

import (
	"math"
	"testing"
)

func TestParse(t *testing.T) {
	tests := []struct {
		name    string
		arg     string
		want    Bytes
		wantErr bool
	}{
		{name: "invalid test", arg: "asdf", wantErr: true},
		{name: "empty string", arg: "", wantErr: true},
		{name: "zero", arg: "0", want: 0},
		{name: "one", arg: "1", want: 1},
		{name: "1B", arg: "1B", want: B},
		{name: "1KiB", arg: "1KiB", want: KiB},
		{name: "1TiB", arg: "1TiB", want: TiB},
		{name: "1PiB", arg: "1PiB", want: PiB},
		{name: "MAXINT+1", arg: "18446744073709551616B", wantErr: true},
		{name: "MAXINT/1024+1KiB (overflow)", arg: "18014398509481985KiB", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.arg)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse(%q) error = %v, wantErr %v", tt.arg, err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Parse(%q) = %v, want %v", tt.arg, got, tt.want)
			}
		})
	}
}

func TestString(t *testing.T) {
	tests := []struct {
		name string
		arg  Bytes
		want string
	}{
		{name: "0B", arg: 0, want: "0B"},
		{name: "1B", arg: 1, want: "1B"},
		{name: "1023B", arg: KiB - 1, want: "1023B"},
		{name: "1KiB", arg: KiB, want: "1KiB"},
		{name: "1337", arg: 1337, want: "1.3KiB"},
		{name: "5MiB", arg: 5 * MiB, want: "5MiB"},
		{name: "5.1MiB", arg: 5*MiB + 123*KiB, want: "5.1MiB"}, // fractional values use one decimal place
		{name: "1.5GiB", arg: GiB + 512*MiB, want: "1.5GiB"},
		{name: "2TiB", arg: 2 * TiB, want: "2TiB"},
		{name: "1PiB", arg: PiB, want: "1PiB"},
		{name: "MAXINT", arg: Bytes(math.MaxUint64), want: "16383.9PiB"}, // 16384PiB - 1B
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.arg.String()
			if got != tt.want {
				t.Errorf("Bytes(%d).String() = %q, want %q", tt.arg, got, tt.want)
			}
		})
	}
}
