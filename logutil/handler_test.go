package logutil

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHandler_ScopedError(t *testing.T) {
	r := require.New(t)
	a := assert.New(t)

	testErr := NewError(nil, "test error", slog.String("foo", "bar"))
	a.EqualValues(`test error with foo="bar"`, testErr.Error())

	buf := bytes.NewBuffer(nil)
	h, err := NewHandlerTo(buf, FormatJSON, LevelTrace)
	r.NoError(err)
	log := slog.New(h)
	log.Info("message", Err(testErr))

	var slogMsg map[string]any
	r.NoError(json.Unmarshal(buf.Bytes(), &slogMsg))

	a.EqualValues("message", slogMsg[slog.MessageKey])
	a.EqualValues("test error", slogMsg[ErrKey])
	a.EqualValues("bar", slogMsg["foo"])
}
