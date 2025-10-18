package mainutil

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/authenticvision/util-go/buildinfo"
	"github.com/authenticvision/util-go/logutil"
	"github.com/mologie/nicecmd"
	"github.com/spf13/cobra"
)

var InKubernetes = func() bool {
	_, ok := os.LookupEnv("KUBERNETES_SERVICE_HOST")
	return ok
}()

// Main is equivalent to calling RootCommand, followed by Run. It does not return.
func Main[T LogConfigEmbedder](
	setup nicecmd.Hook[T],
	run nicecmd.Hook[T],
	cmdTmpl cobra.Command,
	cfg T,
	opts ...nicecmd.Option,
) {
	Run(RootCommand(setup, run, cmdTmpl, cfg, opts...))
}

// RootCommand creates a cobra.Command through nicecmd.RootCommand.
// The setup function is always run before any sub-command, but after global context setup (logger, etc.).
// The run function is run only when the main command is run, and is wrapped to print version info.
// Both are optional, i.e. it's fine for the root command to have no setup or no run line.
func RootCommand[T LogConfigEmbedder](
	setup nicecmd.Hook[T],
	run nicecmd.Hook[T],
	cmdTmpl cobra.Command,
	cfg T,
	opts ...nicecmd.Option,
) *cobra.Command {
	if setup != nil {
		setupNext := setup
		setup = func(cfg *T, cmd *cobra.Command, args []string) error {
			if err := setupContext(cfg, cmd, args); err != nil {
				return err
			}
			if err := setupNext(cfg, cmd, args); err != nil {
				return err
			}
			return nil
		}
	} else {
		setup = setupContext
	}

	if run != nil {
		runNext := run
		run = func(cfg *T, cmd *cobra.Command, args []string) error {
			LogVersion(cmd)
			return runNext(cfg, cmd, args)
		}
	}

	cmd := nicecmd.RootCommand(nicecmd.SetupAndRun(setup, run), cmdTmpl, cfg, opts...)
	cmd.SilenceErrors = true // for logging them ourselves via slog
	if !cmd.SilenceUsage {
		cmd.SilenceUsage = InKubernetes // to avoid noise, though locally this is quite helpful
	}

	if cmd.Version == "" {
		if buildinfo.Version != "" {
			cmd.Version = buildinfo.Version
		} else {
			cmd.Version = buildinfo.GitCommit
		}
	}

	return cmd
}

func setupContext[T LogConfigEmbedder](cfg *T, cmd *cobra.Command, args []string) error {
	// logutil replaces slog.Default() and the older log package's output
	if err := (*cfg).LogConfigEmbed().Log.InstallForProcess(); err != nil {
		slog.Error("error installing log handler", logutil.Err(err))
		os.Exit(1)
	}
	log := slog.Default()

	// replaced log handler must be applied to current and possibly separate root command
	ctx := cmd.Context()
	ctx = logutil.WithLogContext(ctx, log)
	cmd.SetContext(ctx)
	if rootCmd := cmd.Root(); rootCmd != cmd {
		rootCmd.SetContext(logutil.WithLogContext(rootCmd.Context(), log))
		// func Run will now log command failure in the requested format
	}

	// cgroup2 compatibility, native to Go 1.25, needed for everything before that
	if err := autoMaxProcs(); err != nil {
		log.ErrorContext(ctx, "failed to update GOMAXPROCS", logutil.Err(err))
		// continue regardless
	}

	return nil
}

type signalStopTag struct{}

// RestoreSignals restores default signal behavior for programs set up through RootCommand.
// This is useful for utility sub-commands which run in a terminal.
func RestoreSignals(ctx context.Context) {
	stop := ctx.Value(signalStopTag{}).(context.CancelFunc)
	stop()
}

// Run executes the given command and exits with an appropriate status code.
// The command's context is canceled upon SIGINT or SIGTERM.
func Run(cmd *cobra.Command) {
	err := func() error {
		ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
		ctx = context.WithValue(ctx, signalStopTag{}, stop)
		ctx = logutil.WithLogContext(ctx, slog.Default()) // usually unused, hit with e.g. --help
		defer stop()
		return cmd.ExecuteContext(ctx)
	}()
	if err != nil && !errors.Is(err, context.Canceled) {
		ctx := cmd.Context() // might be different from ctx set up in closure above
		log := logutil.FromContext(ctx)
		log.ErrorContext(ctx, "command failed", logutil.Err(err))
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

// LogVersion is already called via Main or RootCommand, but exported for use in custom sub-commands.
// It logs the application's version embedded via buildinfo.
func LogVersion(cmd *cobra.Command) {
	attrs := []any{
		slog.String("git_commit", buildinfo.GitCommit),
		slog.Any("git_commit_date", buildinfo.GitCommitDate),
	}
	if buildinfo.Version != "" {
		attrs = append(attrs, slog.String("version", buildinfo.Version))
	}

	ctx := cmd.Context()
	log := logutil.FromContext(ctx)
	log.InfoContext(ctx, "starting "+cmd.Root().DisplayName(), attrs...)
}

type LogConfigEmbedder interface {
	LogConfigEmbed() LogConfig
}

type LogConfig struct {
	Log logutil.Config `flag:"persistent"`
}

func (c LogConfig) LogConfigEmbed() LogConfig {
	return c
}

var LogDefault = LogConfig{Log: logutil.DefaultConfig}
