package configutil

import (
	"encoding"
	"errors"
	"fmt"
	"git.avdev.at/dev/util"
	"github.com/iancoleman/strcase"
	"os"
	"reflect"
	"strconv"
	"strings"
)

// Parse parses config from the environment (via EnvGetter) into a struct T.
// Maps and Pointers in the struct are not supported. Custom types are supported
// via encoding.TextUnmarshaller.
// Whether a field is required is parsed from the `required` tag. The default
// value to use when no value is set in the environment is taken from the
// `default` tag of a field. If there is none, the field will not be explicitly
// initialized.
func Parse[T any](env EnvGetter) (*T, error) {
	if env == nil {
		env = OSEnv{}
	}
	var ret T
	v := reflect.ValueOf(&ret).Elem()
	if v.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %T must be a struct", ret)
	}
	p := parser{}
	err := p.parseRoot(v)
	if err != nil {
		return nil, fmt.Errorf("invalid config struct: %w", err)
	}
	errs := p.execute(env)
	if len(errs) > 0 {
		_, _ = fmt.Fprintf(os.Stderr, "The following errors occured while parsing configuration:\n")
		for _, cve := range errs {
			_, _ = fmt.Fprintf(os.Stderr, "%s: %s\n", cve.EnvVar, cve.err.Error())
		}
		return nil, errs
	}
	return &ret, nil
}

func (p *parser) parseRoot(v reflect.Value) error {
	for i := 0; i < v.NumField(); i++ {
		f := v.Field(i)
		t := v.Type().Field(i)
		err := p.parseValue("", t, f)
		if err != nil {
			return fmt.Errorf("field %q: %w", t.Name, err)
		}
	}
	return nil
}

type parser struct {
	configValues []configValue
}

var unmarshalerType = reflect.TypeFor[encoding.TextUnmarshaler]()

type configValueSetter func(value string) error

type configValue struct {
	EnvVar   string
	Default  string // empty string if no default
	Required bool
	Setter   configValueSetter
}

var ErrRequired = errors.New("required field is not set")

func (p *parser) parseValue(prefix string, field reflect.StructField, value reflect.Value) error {
	if !field.IsExported() {
		return nil
	}
	name := prefix + strcase.ToScreamingSnake(field.Name)
	prefix = name + "_"
	required, err := parseRequiredTag(field)
	if err != nil {
		return err
	}
	def, defOK := field.Tag.Lookup("default")
	if defOK && def == "" {
		return errors.New("default tag cannot be empty")
	}

	if required && defOK {
		return fmt.Errorf("field cannot be required and have a default")
	}

	cf := func(setter func(s string) error) {
		p.configValues = append(p.configValues, configValue{
			EnvVar:   name,
			Default:  def,
			Required: required,
			Setter:   setter,
		})
	}

	if p, ok := customTypes[value.Type()]; ok {
		cf(func(s string) error {
			parsed, err := p(s)
			if err != nil {
				return err
			}
			value.Set(reflect.ValueOf(parsed))
			return nil
		})
		return nil
	}

	if value.Addr().Type().Implements(unmarshalerType) {
		unmarshaler := value.Addr().Interface().(encoding.TextUnmarshaler)
		cf(func(s string) error {
			return unmarshaler.UnmarshalText([]byte(s))
		})
		return nil
	}

	if value.Kind() == reflect.Pointer {
		return errors.New("destination field cannot be a pointer")
	}

	switch value.Kind() {
	case
		reflect.Bool, reflect.String,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		cf(func(s string) error {
			return p.parsePrimitive(value, s)
		})
	case reflect.Slice:
		cf(func(s string) error {
			parts := strings.Split(s, ",")
			slice := reflect.MakeSlice(value.Type(), len(parts), len(parts))
			for i, s := range parts {
				err := p.parsePrimitive(slice.Index(i), s)
				if err != nil {
					return fmt.Errorf("element %d %q: %w", i, s, err)
				}
			}
			value.Set(slice)
			return nil
		})
	case reflect.Struct:
		if required {
			return fmt.Errorf("nested structs can't be required")
		}
		for i := 0; i < value.NumField(); i++ {
			f := value.Field(i)
			t := value.Type().Field(i)
			err := p.parseValue(prefix, t, f)
			if err != nil {
				return fmt.Errorf("field %q: %w", t.Name, err)
			}
		}
	default:
		return fmt.Errorf("unsupported destination type: %s", value.Type())
	}
	return nil
}

func (p *parser) parsePrimitive(dest reflect.Value, value string) error {
	if dest.Type().Implements(unmarshalerType) {
		unmarshaler := dest.Interface().(encoding.TextUnmarshaler)
		return unmarshaler.UnmarshalText([]byte(value))
	}

	parseInt := func(bits int) error {
		i, err := strconv.ParseInt(value, 0, bits)
		if err != nil {
			return err
		}
		dest.SetInt(i)
		return nil
	}
	parseUint := func(bits int) error {
		i, err := strconv.ParseUint(value, 0, bits)
		if err != nil {
			return err
		}
		dest.SetUint(i)
		return nil
	}
	parseFloat := func(bits int) error {
		i, err := strconv.ParseFloat(value, bits)
		if err != nil {
			return err
		}
		dest.SetFloat(i)
		return nil
	}

	if dest.Kind() == reflect.Pointer {
		return errors.New("destination field cannot be a pointer")
	}

	switch dest.Kind() {
	case reflect.Bool:
		b, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		dest.SetBool(b)
		return nil
	case reflect.Int8:
		return parseInt(8)
	case reflect.Int16:
		return parseInt(16)
	case reflect.Int32:
		return parseInt(32)
	case reflect.Int, reflect.Int64:
		return parseUint(64)
	case reflect.Uint8:
		return parseUint(8)
	case reflect.Uint16:
		return parseUint(16)
	case reflect.Uint32:
		return parseUint(32)
	case reflect.Uint, reflect.Uint64:
		return parseUint(64)
	case reflect.Float32:
		return parseFloat(32)
	case reflect.Float64:
		return parseFloat(64)
	case reflect.String:
		dest.SetString(value)
		return nil
	default:
		return fmt.Errorf("unsupported primitive destination type: %s", dest.Type())
	}
}

type ConfigValueError struct {
	EnvVar string
	err    error
}

func (e ConfigValueError) Error() string {
	return fmt.Sprintf("env var %s: %s", e.EnvVar, e.err.Error())
}

func (e ConfigValueError) Unwrap() error {
	return e.err
}

type configErrors []ConfigValueError

func (c configErrors) Error() string {
	return fmt.Sprintf(
		"one or more config errors have occured: %s",
		strings.Join(util.Map(c, func(cve ConfigValueError) string {
			return cve.Error()
		}), "; "))
}

func (c configErrors) Unwrap() []error {
	return util.Map(c, func(cve ConfigValueError) error {
		return cve
	})
}

func (p *parser) execute(env EnvGetter) configErrors {
	var errs configErrors = nil
	for _, cv := range p.configValues {

		envVal := env.Get(cv.EnvVar)
		if envVal != "" {
			err := cv.Setter(envVal)
			if err != nil {
				errs = append(errs, ConfigValueError{EnvVar: cv.EnvVar, err: err})
			}
			continue
		}
		if cv.Required {
			errs = append(errs, ConfigValueError{EnvVar: cv.EnvVar, err: ErrRequired})
			continue
		}
		if cv.Default == "" {
			// no default set, don't initialize
			continue
		}
		err := cv.Setter(cv.Default)
		if err != nil {
			err = fmt.Errorf("setting default value %q: %w", cv.Default, err)
			errs = append(errs, ConfigValueError{EnvVar: cv.EnvVar, err: err})
		}
	}
	return errs
}

func parseRequiredTag(field reflect.StructField) (bool, error) {
	required := false
	if reqStr, ok := field.Tag.Lookup("required"); ok {
		var err error
		required, err = strconv.ParseBool(reqStr)
		if err != nil {
			return false, err
		}
	}
	return required, nil
}
