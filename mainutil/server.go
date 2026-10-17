package mainutil

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/authenticvision/util-go/httpmw"
	"github.com/authenticvision/util-go/httpp"
	"github.com/authenticvision/util-go/logutil"
	"github.com/mologie/nicecmd"
	"github.com/spf13/cobra"
)

// ShutdownTimeout is the grace period that requests have to finish after shutdown is requested.
var ShutdownTimeout = 20 * time.Second

type ServerMain[T any] func(cfg *T, cmd *cobra.Command, args []string) (httpp.Handler, error)

// Server is a convenience function for single-purpose HTTP servers that require no additional
// setup or teardown. It should wrap a function that returns an httpp.Handler for mainutil.Main.
func Server[T ServerConfigEmbedder](serverMain ServerMain[T], opts ...ServerOption) nicecmd.Hook[T] {
	return func(cfg *T, cmd *cobra.Command, args []string) error {
		sc := (*cfg).ServerConfigEmbed()
		if sc.TLSCertFile != "" || sc.TLSKeyFile != "" {
			if sc.TLSCertFile == "" {
				return fmt.Errorf("TLS cert is given but TLS key is empty")
			} else if sc.TLSKeyFile == "" {
				return fmt.Errorf("TLS key is given but TLS cert is empty")
			}
			opts = append([]ServerOption{WithTLS(sc.TLSCertFile, sc.TLSKeyFile)}, opts...)
		}
		if handler, err := serverMain(cfg, cmd, args); err != nil {
			return fmt.Errorf("server main: %w", err)
		} else if err := ListenAndServe(cmd.Context(), sc.BindAddr, handler, opts...); err != nil {
			return fmt.Errorf("listen and serve %q: %w", sc.BindAddr, err)
		} else {
			return nil
		}
	}
}

type ServerOption func(*http.Server) error

// WithPlainHTTP2 enables plain-text HTTP 2 in addition to HTTP 1, e.g. for a gRPC server.
func WithPlainHTTP2() ServerOption {
	return func(server *http.Server) error {
		server.Protocols.SetUnencryptedHTTP2(true)
		return nil
	}
}

// WithOnShutdown launches f in a separate goroutine when the HTTP server is shut down.
// Shutdown usually happens a few seconds after termination is signaled.
func WithOnShutdown(f func()) ServerOption {
	return func(server *http.Server) error {
		server.RegisterOnShutdown(f)
		return nil
	}
}

// WithTLSCert enables HTTPS with the given certificate.
func WithTLSCert(cert tls.Certificate) ServerOption {
	return func(server *http.Server) error {
		if server.TLSConfig == nil {
			server.TLSConfig = &tls.Config{}
		}
		server.TLSConfig.Certificates = append(server.TLSConfig.Certificates, cert)
		return nil
	}
}

// WithTLS loads the given cert and key and enables HTTPS.
func WithTLS(certFile, keyFile string) ServerOption {
	return func(server *http.Server) error {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			return fmt.Errorf("load TLS keypair: %w", err)
		}
		return WithTLSCert(cert)(server)
	}
}

func ListenAndServe(ctx context.Context, addr string, handler httpp.Handler, opts ...ServerOption) error {
	log := logutil.FromContext(ctx)

	network := "tcp"
	var ok bool
	if addr, ok = strings.CutPrefix(addr, "unix:"); ok {
		network = "unix"
		defer func() {
			if err := os.Remove(addr); err != nil {
				log.Error("failed to remove unix socket file", logutil.Err(err))
			}
		}()
	}
	l, err := net.Listen(network, addr)
	if err != nil {
		return fmt.Errorf("listen %q: %w", addr, err)
	}

	reqCtx, reqCancel := context.WithCancel(context.WithoutCancel(ctx))
	defer reqCancel()
	server := &http.Server{
		Addr: addr,
		Handler: httpp.NeverErrors(httpmw.Chain(handler,
			httpmw.NewCompressionMiddleware(),
			httpmw.NewPanicMiddleware(),
			httpmw.NewLogMiddleware(log),
		)),
		BaseContext: func(net.Listener) context.Context {
			// Requests are not launched with a separate cancellation scope, so that they get a
			// grace period of 20 seconds to complete after termination is requested.
			return reqCtx
		},
		Protocols: func() *http.Protocols {
			// in contrast to the http.Server defaults, we don't offer
			// disabling HTTP2 via GODEBUG. It can still be disabled via
			// a custom ServerOption
			var p http.Protocols
			p.SetHTTP1(true)
			p.SetHTTP2(true)
			return &p
		}(),
	}
	for _, opt := range opts {
		err = opt(server)
		if err != nil {
			return err
		}
	}

	scheme := "http"
	if server.TLSConfig != nil {
		scheme = "https"
	}
	log.Info("listening",
		slog.String("bind_addr", addr),
		slog.String("link", fmt.Sprintf("%s://%s", scheme, addr)))

	serveErr := make(chan error)
	go func() {
		// This goroutine runs until server.Shutdown() is called.
		defer close(serveErr)
		if server.TLSConfig != nil {
			serveErr <- server.ServeTLS(l, "", "")
		} else {
			serveErr <- server.Serve(l)
		}
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

		// Shutdown makes ListenAndServe return immediately.
		// Shutdown returns when all handlers have completed, or after at most 20-ish seconds.
		if err := server.Shutdown(context.Background()); err != nil {
			return fmt.Errorf("server shutdown: %w", err)
		}

		// This is purely cosmetic to catch stray errors, I don't expect anything in practice.
		if err := <-serveErr; err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("serve goroutine after shutdown: %w", err)
		}

		return ctx.Err()

	case <-serveErr:
		return fmt.Errorf("serve goroutine: %w", err)
	}
}

type ServerConfigEmbedder interface {
	ServerConfigEmbed() ServerConfig
}

type ServerConfig struct {
	BindAddr    string `usage:"bind address for HTTP connections. to use a unix socket, prefix with 'unix:'"`
	TLSCertFile string `usage:"PEM cert file. enables HTTPS when set together with --tls-key"`
	TLSKeyFile  string `usage:"PEM key file. enables HTTPS when set together with --tls-cert"`
}

func (c ServerConfig) ServerConfigEmbed() ServerConfig {
	return c
}

var ServerDefault = ServerConfig{
	BindAddr: "localhost:8080",
}
