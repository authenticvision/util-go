package grpcutil

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"reflect"

	"github.com/authenticvision/util-go/logutil"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/proto"
)

func UnaryServerLogContextInterceptor(log *slog.Logger) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		ctx = logutil.WithLogContext(ctx, log)
		ctx = context.WithValue(ctx, accessLogTag{}, &accessLog{})
		return handler(ctx, req)
	}
}

func StreamServerLogContextInterceptor(log *slog.Logger) grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		ctx := stream.Context()
		ctx = logutil.WithLogContext(ctx, log)
		ctx = context.WithValue(ctx, accessLogTag{}, &accessLog{})
		wrapped := middleware.WrapServerStream(stream)
		wrapped.WrappedContext = ctx
		return handler(srv, wrapped)
	}
}

func UnaryServerLogWriterInterceptor() grpc.UnaryServerInterceptor {
	return logging.UnaryServerInterceptor(&logHandler{},
		logging.WithCodes(errToCode),
		logging.WithErrorFields(errToFields),
	)
}

func StreamServerLogWriterInterceptor() grpc.StreamServerInterceptor {
	return logging.StreamServerInterceptor(&logHandler{},
		logging.WithCodes(errToCode),
		logging.WithErrorFields(errToFields),
	)
}

func errToCode(err error) codes.Code {
	if errors.Is(err, context.Canceled) {
		return codes.Canceled
	} else {
		return logging.DefaultErrorToCode(err)
	}
}

// errToFields simply exposes the error to ErrKey as-is, so that logutil's sinks can properly
// destructure it. grpc-middleware unfortunately just dumps a stringified error into grpc.error and
// thereby loses any type information and attachments.
func errToFields(err error) logging.Fields {
	return []any{logutil.ErrKey, logutil.Err(err).Value}
}

type accessLogTag struct{}

type accessLog struct {
	SuppressInfoLog bool
	User            *logutil.UserValue
}

// DisableAccessLog suppresses informational access log lines for the request.
// This only affects the application's internal access log.
func DisableAccessLog(ctx context.Context) {
	if p, ok := ctx.Value(accessLogTag{}).(*accessLog); ok {
		p.SuppressInfoLog = true
	}
}

// WithRequestUser attaches the given user identity to the context's log, and
// additionally adds it to the call's top-level log scope for access logs.
func WithRequestUser(ctx context.Context, user logutil.UserValue) context.Context {
	if p, ok := ctx.Value(accessLogTag{}).(*accessLog); ok {
		p.User = &user
	}
	log := logutil.FromContext(ctx)
	log = log.With(slog.Any(logutil.UserKey, user))
	return logutil.WithLogContext(ctx, log)
}

// logHandler implements logging.Logger of go-grpc-middleware, which will measure request duration
// and write a log message at the start and at the end of a request.
type logHandler struct{}

// Log is a sink for logging.Logger. It additionally consults the context's accessLog tag to honor
// requests made via DisableAccessLog and WithRequestUser.
func (*logHandler) Log(ctx context.Context, levelInt logging.Level, msg string, fields ...any) {
	level := slog.Level(levelInt)
	log := logutil.FromContext(ctx)
	if p, ok := ctx.Value(accessLogTag{}).(*accessLog); ok {
		if p.SuppressInfoLog && level == slog.LevelInfo {
			return
		}
		if user := p.User; user != nil {
			log = log.With(slog.Any(logutil.UserKey, *user))
		}
	}
	f := (*logging.Fields)(&fields)
	f.Delete("grpc.error") // stringified duplicate of errToFields's proper error field
	log.Log(ctx, level, msg, fields...)
}

func LogAttr(key string, m interface {
	proto.Message
	fmt.Stringer
}) slog.Attr {
	if reflect.ValueOf(m).IsNil() {
		return slog.String(key, "<nil>")
	}
	j, err := protojson.Marshal(m)
	if err != nil {
		return slog.String(key, m.String())
	} else {
		return logutil.JSON(key, j)
	}
}
