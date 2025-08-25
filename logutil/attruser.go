package logutil

import "log/slog"

const UserKey = "usr"

func User(user UserValue) slog.Attr {
	return slog.Any(UserKey, user)
}

// UserValue is a generic user identifier modeled after Datadog's standard log attributes.
type UserValue struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}
