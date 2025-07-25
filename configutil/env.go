package configutil

import (
	"bufio"
	"fmt"
	"github.com/BooleanCat/go-functional/v2/it"
	"github.com/BooleanCat/go-functional/v2/it/itx"
	"maps"
	"os"
	"slices"
	"strings"
)

type EnvGetter interface {
	// Get returns the value associated with a given variable name or an empty string if absent or empty
	Get(key string) string
	// List returns a list of all defined variable names
	List() []string
}

type MapEnv map[string]string

func (e MapEnv) Get(key string) string {
	return e[key]
}

func (e MapEnv) List() []string {
	return slices.Collect(maps.Keys(e))
}

type OSEnv struct{}

func (e OSEnv) Get(key string) string {
	return os.Getenv(key)
}

func (e OSEnv) List() []string {
	// TODO: windows support, see os.Getenv for windows
	return itx.FromSlice(os.Environ()).Transform(func(s string) string {
		name, _, found := strings.Cut(s, "=")
		if !found || name == "" {
			panic(fmt.Sprintf("invalid environment variable string %q", s))
		}
		return name
	}).Collect()
}

type FallbackEnv struct {
	Primary  EnvGetter
	Fallback EnvGetter
}

func (e FallbackEnv) Get(key string) string {
	value := e.Primary.Get(key)
	if value == "" {
		value = e.Fallback.Get(key)
	}
	return value
}

func (e FallbackEnv) List() []string {
	a := slices.Values(e.Primary.List())
	b := slices.Values(e.Fallback.List())
	return slices.Collect(it.FilterUnique(it.Chain(a, b)))
}

func EnvFromFile(path string) (EnvGetter, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer func() { _ = file.Close() }()

	env := MapEnv{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		name, value, found := strings.Cut(line, "=")
		if !found {
			return nil, fmt.Errorf("invalid line in env file: %s", line)
		}
		env[name] = value
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan file: %w", err)
	}
	return env, nil
}

// TrackingEnv tracks which environment variables have been fetched.
type TrackingEnv struct {
	fetched  map[string]struct{}
	delegate EnvGetter
}

func NewTrackingEnv(delegate EnvGetter) *TrackingEnv {
	return &TrackingEnv{
		fetched:  make(map[string]struct{}),
		delegate: delegate,
	}
}

// Fetched returns the list of env keys that have been fetched (regardless of whether they existed or not)
func (e *TrackingEnv) Fetched() []string {
	return slices.Collect(maps.Keys(e.fetched))
}

// Unfetched returns the list of keys that have *not* been fetched.
func (e *TrackingEnv) Unfetched() []string {
	return itx.FromSlice(e.delegate.List()).Filter(func(name string) bool {
		_, found := e.fetched[name]
		return !found
	}).Collect()
}

func (e *TrackingEnv) Get(key string) string {
	e.fetched[key] = struct{}{}
	return e.delegate.Get(key)
}

func (e *TrackingEnv) List() []string {
	return e.delegate.List()
}
