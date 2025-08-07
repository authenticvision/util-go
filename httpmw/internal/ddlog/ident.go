package ddlog

const UserKey = "usr"

// User is a generic user identifier modeled after Datadog's standard log attributes.
type User struct {
	ID    string `json:"id,omitempty"`
	Name  string `json:"name,omitempty"`
	Email string `json:"email,omitempty"`
}
