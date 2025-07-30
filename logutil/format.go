package logutil

import (
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it/op"
	"github.com/spf13/pflag"
	"strings"
)

var _ pflag.Value = op.Ref(FormatText)

type Format string

const (
	FormatText Format = "TEXT"
	FormatJSON Format = "JSON"
)

func (f *Format) UnmarshalText(text []byte) error {
	format := Format(strings.ToUpper(string(text)))
	switch format {
	case FormatText, FormatJSON:
		*f = format
		return nil
	}
	return fmt.Errorf("invalid log format text: %q", text)
}

func (f *Format) Set(s string) error {
	return f.UnmarshalText([]byte(s))
}

func (f *Format) String() string {
	return string(*f)
}

func (f *Format) Type() string {
	return "format"
}
