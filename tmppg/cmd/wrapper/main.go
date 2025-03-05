// inspired by: https://michael.stapelberg.ch/posts/2024-11-19-testing-with-go-and-postgresql-ephemeral-dbs/
package main

import (
	"git.avdev.at/dev/util/logutil"
	"git.avdev.at/dev/util/tmppg"
	"log/slog"
	"os"
)

func main() {
	if len(os.Args) < 3 || os.Args[1] != "--" {
		slog.Error("usage: tmppg -- command [args...]")
		os.Exit(1)
	}
	args := os.Args[2:]
	if err := tmppg.RunWithPostgresql(args); err != nil {
		slog.Error("uncaught error", logutil.Err(err))
		os.Exit(1)
	}
}
