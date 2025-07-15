package configutil

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

type CustomType string

func (c *CustomType) UnmarshalText(b []byte) error {
	*c = CustomType(b)
	return nil
}

type TestConfig struct {
	A            string     `default:"a"`
	B            string     `required:"true"`
	C            CustomType `default:"c"`
	NestedConfig NestedConfig
}

type NestedConfig struct {
	X string
}

func TestParse(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	cfg, err := Parse[TestConfig](MapEnv{"B": "b", "NESTED_CONFIG_X": "x"}, "")
	r.NoError(err)
	a.Equal("a", cfg.A)
	a.Equal("b", cfg.B)
	a.Equal(CustomType("c"), cfg.C)
	a.Equal("x", cfg.NestedConfig.X)
}

func testTagsCase[T any](t *testing.T, name string, wantErrContains string) {
	t.Run(name, func(t *testing.T) {
		r := require.New(t)
		env := MapEnv{"BOOL": "1", "NESTED_BOOL": "true"}
		_, err := Parse[T](env, "")
		if wantErrContains != "" {
			r.ErrorContains(err, wantErrContains)
			fmt.Printf("expected error: %s\n", err.Error())
		} else {
			r.NoError(err)
		}
	})
}

func TestTags(t *testing.T) {
	const NoErr = ""

	testTagsCase[struct {
		Bool bool `required:"true" blubb:""`
	}](t, "required=true", NoErr)
	testTagsCase[struct {
		Bool bool `required:"" blubb:""`
	}](t, "required=empty", "invalid syntax")
	testTagsCase[struct {
		Bool bool `required:"0" blubb:""`
	}](t, "required=zero", NoErr)
	testTagsCase[struct {
		OtherBool bool `required:"0" blubb:""`
	}](t, "required=zero and not present in env", NoErr)
	testTagsCase[struct {
		Bool bool `required:"true" default:"hi"`
	}](t, "required and default", "field cannot be required and have a default")
	testTagsCase[struct {
		Nested struct{ Bool bool } `required:"true"`
	}](t, "nested struct can't be required", "nested structs can't be required")
	testTagsCase[struct {
		Nested struct {
			Bool bool `required:"true"`
		}
	}](t, "nested struct element required", NoErr)

	testTagsCase[struct {
		OtherBool bool `required:"false" default:"true"`
	}](t, "default", NoErr)
	testTagsCase[struct {
		OtherBool bool `required:"false" default:"true"`
	}](t, "explicit not required and default", NoErr)
	testTagsCase[struct {
		OtherBool bool `default:"hi"`
	}](t, "default with invalid syntax", "invalid syntax")
	//testTagsCase[struct {
	//	Bool bool `default:"hi"`
	//}](t, "unused default with invalid syntax", "invalid syntax")
}

func TestParsePrefix(t *testing.T) {
	a := assert.New(t)
	r := require.New(t)
	type config struct {
		A string `default:"a"`
	}
	cfg, err := Parse[config](MapEnv{"PFX_A": "a", "UNRELATED": "hi"}, "PFX_")
	r.NoError(err)
	a.Equal("a", cfg.A)
}

func TestParsePrefixFail(t *testing.T) {
	r := require.New(t)
	type config struct {
		A string `default:"a"`
	}
	cfg, err := Parse[config](MapEnv{"PFX_B": "b"}, "PFX_")
	r.Nil(cfg)
	r.ErrorContains(err, "not all defined environment variables used in config: PFX_B")
}
