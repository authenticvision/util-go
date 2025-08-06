package mainutil

import (
	"context"
	"errors"
	"fmt"
	"github.com/authenticvision/util-go/httpmw"
	"github.com/authenticvision/util-go/httpp"
	"github.com/authenticvision/util-go/logutil"
	"github.com/mologie/nicecmd"
	"github.com/spf13/cobra"
	"log/slog"
	"net"
	"net/http"
	"time"
)

// ShutdownTimeout is the grace period that requests have to finish after shutdown is requested.
var ShutdownTimeout = 20 * time.Second

type ServerMain[T any] func(cfg *T, cmd *cobra.Command, args []string) (httpp.Handler, error)

// Server is a convenience function for single-purpose HTTP servers that require no additional
// setup or teardown. It should wrap a function that returns an httpp.Handler for mainutil.Main.
func Server[T ServerConfigEmbedder](serverMain ServerMain[T]) nicecmd.Hook[T] {
	return func(cfg *T, cmd *cobra.Command, args []string) error {
		addr := (*cfg).ServerConfigEmbed().BindAddr
		if handler, err := serverMain(cfg, cmd, args); err != nil {
			return fmt.Errorf("server main: %w", err)
		} else if err := ListenAndServe(cmd.Context(), addr, handler); err != nil {
			return fmt.Errorf("listen and serve %q: %w", addr, err)
		} else {
			return nil
		}
	}
}

func ListenAndServe(ctx context.Context, addr string, handler httpp.Handler) error {
	log := logutil.FromContext(ctx)

	l, err := net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", addr, err)
	}
	//goland:noinspection HttpUrlsUsage
	log.Info("listening",
		slog.String("bind_addr", addr),
		slog.String("link", fmt.Sprintf("http://%s", addr)))

	reqCtx, reqCancel := context.WithCancel(context.Background())
	defer reqCancel()
	server := &http.Server{
		Addr: addr,
		Handler: httpp.NeverErrors(httpmw.Chain(handler,
			httpmw.NewCompressionMiddleware(),
			httpmw.NewPanicMiddleware(),
			httpmw.NewLogMiddleware(log),
		)),
		BaseContext: func(net.Listener) context.Context {
			// Requests are not launched from cmd's context (which is canceled on SIGTERM), but
			// instead from context.Background. They get a grace period of 20 seconds to complete
			// after termination is requested.
			return reqCtx
		},
	}
	serveErr := make(chan error)
	go func() {
		// This goroutine runs until server.Shutdown() is called.
		defer close(serveErr)
		serveErr <- server.Serve(l)
	}()

	select {
	case <-ctx.Done():
		// Notify active requests to terminate after grace period.
		time.AfterFunc(ShutdownTimeout, reqCancel)

		// Give K8s's load balancer time to settle before stopping to accept connections.
		// This is a cheap way to avoid 502 Bad Gateway errors for in-flight requests.
		if InKubernetes {
			time.Sleep(3 * time.Second)
		}

		// Shutdown makes ListenAndServer return immediately.
		// Shutdown returns when all handlers have completed, or after at most 20-ish seconds.
		if err := server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}

		// This is purely cosmetic to catch stray errors, I don't expect anything in practice.
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("server serve: %w", err)
		}

		return nil

	case <-serveErr:
		return fmt.Errorf("server startup: %w", err)
	}
}

type ServerConfigEmbedder interface {
	ServerConfigEmbed() ServerConfig
}

type ServerConfig struct {
	BindAddr string `usage:"address for HTTP connections"`
}

func (c ServerConfig) ServerConfigEmbed() ServerConfig {
	return c
}

var ServerDefault = ServerConfig{
	BindAddr: "localhost:8080",
}
