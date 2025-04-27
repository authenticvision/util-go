package logutil

import (
	"fmt"
	"strings"
)

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

func (f *Format) String() string {
	return string(*f)
}

func (f *Format) CmdTypeDesc() string {
	return "format"
}
