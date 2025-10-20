package logutil

import (
	"encoding/json"
	"log/slog"
	"maps"
	"slices"
)

func JSON(key string, j json.RawMessage) slog.Attr {
	if json.Valid(j) {
		return slog.Any(key, jsonAttr(j))
	} else {
		// separate to avoid exposing MarshalJSON for invalid input
		return slog.String(key, string(j))
	}
}

type jsonAttr json.RawMessage

func (j jsonAttr) MarshalJSON() ([]byte, error) {
	return j, nil
}

func (j jsonAttr) LogValue() slog.Value {
	var v any
	if err := json.Unmarshal(j, &v); err != nil {
		// should be unreachable due to the json.Valid check in JSON()
		return slog.StringValue(string(j))
	}
	return toGroupAttr(v)
}

func toGroupAttr(v any) slog.Value {
	switch x := v.(type) {
	case map[string]any:
		attrs := make([]slog.Attr, 0, len(x))
		for _, k := range slices.Sorted(maps.Keys(x)) {
			attrs = append(attrs, slog.Any(k, toGroupAttr(x[k])))
		}
		return slog.GroupValue(attrs...)
	default:
		return slog.AnyValue(x)
	}
}
