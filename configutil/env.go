package configutil

import (
	"bufio"
	"fmt"
	"git.avdev.at/dev/util"
	"os"
	"slices"
	"strings"
)

type EnvGetter interface {
	Get(key string) string
	List() []string
}

type MapEnv map[string]string

func (e MapEnv) Get(key string) string {
	return e[key]
}

func (e MapEnv) List() []string {
	return util.Keys(e)
}

type OSEnv struct{}

func (e OSEnv) Get(key string) string {
	return os.Getenv(key)
}

func (e OSEnv) List() []string {
	// TODO: windows support, see os.Getenv for windows
	return util.Map(os.Environ(), func(s string) string {
		name, _, found := strings.Cut(s, "=")
		if !found || name == "" {
			panic(fmt.Sprintf("invalid environment variable string %q", s))
		}
		return name
	})
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
	return util.Unique(slices.Concat(e.Primary.List(), e.Fallback.List()))
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
