package tmppg

import (
	"errors"
	"fmt"
	"git.avdev.at/dev/util/logutil"
	"log/slog"
	"os"
	"os/exec"
	"syscall"
	"time"
)

func makeCmd(args ...string) *exec.Cmd {
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd
}

// pg_isready exit codes
const (
	pqPingReject     = 1
	pqPingNoResponse = 2
	pqPingNoAttempt  = 3
)

func WithPostgresql(fn func(socketDir string) error) error {
	dir, err := os.MkdirTemp("", "tmppg")
	if err != nil {
		return fmt.Errorf("setup temporary directory: %w", err)
	}
	defer func() {
		if err := os.RemoveAll(dir); err != nil {
			slog.Error("failed to remove temporary directory", logutil.Err(err))
		}
	}()
	cmd := makeCmd("initdb", "-D", dir, "--no-sync", "--no-instructions")
	if err := cmd.Run(); err != nil {
		slog.Debug("initdb failed with arguments", slog.Any("args", cmd.Args), logutil.Err(err))
		return fmt.Errorf("initdb: %w", err)
	}
	pgCmd := makeCmd("postgres", "-D", dir, "--listen_addresses=", "--unix_socket_directories="+dir, "--fsync=off", "--synchronous_commit=off", "--full_page_writes=off")
	if err := pgCmd.Start(); err != nil {
		slog.Debug("postgres failed with arguments", slog.Any("args", pgCmd.Args), logutil.Err(err))
		return fmt.Errorf("start postgres: %w", err)
	}
	exitErrCh := make(chan error, 1)
	go func() {
		exitErrCh <- pgCmd.Wait()
		close(exitErrCh)
	}()

	// run database removal deferred, so the database also gets removed on
	// runtime.Goexit() and t.FailNow()
	defer func() {
		err := pgCmd.Process.Signal(syscall.SIGTERM)
		if err != nil {
			slog.Error("failed to send SIGTERM to postgres", logutil.Err(err))
		}
		err = <-exitErrCh
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			slog.Error("postgres exited with error", logutil.Err(err))
		} else if err != nil {
			slog.Error("failed to wait for postgres to exit", logutil.Err(err))
		}
	}()

	for {
		select {
		case err := <-exitErrCh:
			return fmt.Errorf("postgres exited unexpectedly: %w", err)
		case <-time.After(100 * time.Millisecond):
		}
		cmd := makeCmd("pg_isready", "-q", "-h", dir, "-d", "postgres")
		err := cmd.Run()
		var exitErr *exec.ExitError
		if err == nil {
			break
		} else if errors.As(err, &exitErr) {
			if exitErr.ExitCode() == pqPingReject || exitErr.ExitCode() == pqPingNoResponse {
				slog.Info("waiting for PostgreSQL to be ready")
			} else {
				// this will trigger the deferred postgres shutdown, too
				return fmt.Errorf("pg_isready: %w", err)
			}
		}
	}

	return fn(dir)
}

// RunWithPostgresql runs the given command with a PostgreSQL instance available.
// Connection information is available via the standard PG* environment variables.
// See https://www.postgresql.org/docs/current/libpq-envars.html
func RunWithPostgresql(args []string) error {
	return WithPostgresql(func(socketDir string) error {
		wrapped := exec.Command(args[0], args[1:]...)
		wrapped.Stdout = os.Stdout
		wrapped.Stderr = os.Stderr
		wrapped.Env = append(os.Environ(), "PGHOST="+socketDir)
		if err := wrapped.Run(); err != nil {
			return fmt.Errorf("%v: %v", wrapped.Args, err)
		}
		return nil
	})
}
