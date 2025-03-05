package testutil

import (
	"os"
	"strconv"
	"testing"
)

func SkipInternet(t *testing.T) {
	s, ok := os.LookupEnv("TEST_WITH_INTERNET")
	if ok {
		if v, err := strconv.ParseBool(s); err != nil {
			t.Fatalf("failed to parse TEST_WITH_INTERNET: %v", err)
		} else if v {
			return
		}
	}
	t.Skip("skipping test that requires internet access")
}
