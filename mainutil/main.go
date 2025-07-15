package mainutil

import (
	"context"
	"errors"
	"git.avdev.at/dev/util/buildinfo"
	"git.avdev.at/dev/util/configutil"
	"git.avdev.at/dev/util/logutil"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
)

func Main[ConfigType interface{ LogConfig() logutil.Config }](name string, configPrefix string, main func(context.Context, ConfigType) error) {
	var err error
	var env configutil.EnvGetter = configutil.OSEnv{}
	if len(os.Args) > 1 {
		fileEnv, err := configutil.EnvFromFile(os.Args[1])
		if err != nil {
			slog.Error("error loading env from file", slog.String("file", os.Args[1]), logutil.Err(err))
			os.Exit(1)
		}
		env = configutil.FallbackEnv{
			Primary:  env,
			Fallback: fileEnv,
		}
	}
	cfg, err := configutil.Parse[ConfigType](env, configPrefix)
	if err != nil {
		slog.Error("error parsing config", logutil.Err(err))
		os.Exit(1)
	}

	if err = (*cfg).LogConfig().InstallForProcess(); err != nil {
		slog.Error("error installing log handler", logutil.Err(err))
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	ctx = logutil.WithLogContext(ctx, slog.Default())

	buildinfo.Log(name)

	if err = main(ctx, *cfg); err != nil {
		if errors.Is(ctx.Err(), context.Canceled) {
			// SIGTERM'd
			return
		}
		slog.Error("error running server", logutil.Err(err))
		os.Exit(1)
	}
}
