package configutil

import (
	"reflect"
	"time"
)

var customTypes = map[reflect.Type]func(value string) (any, error){}

func RegisterType[T any](parser func(value string) (T, error)) {
	customTypes[reflect.TypeFor[T]()] = func(value string) (any, error) {
		return parser(value)
	}
}

func init() {
	RegisterType[time.Duration](time.ParseDuration)
	RegisterType[time.Time](func(value string) (time.Time, error) {
		return time.ParseInLocation(time.RFC3339Nano, value, time.UTC)
	})
}
