package tmppg

import (
	"github.com/stretchr/testify/require"
	"log/slog"
	"os"
	"testing"
)

func TestRunWithPostgresql(t *testing.T) {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})))
	r := require.New(t)
	err := RunWithPostgresql([]string{"psql", "-d", "postgres", "-c", "SELECT 1"})
	r.NoError(err)
}
