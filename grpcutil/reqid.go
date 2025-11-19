package grpcutil

import (
	"context"
	"log/slog"

	"github.com/authenticvision/util-go/logutil"
	"github.com/google/uuid"
	middleware "github.com/grpc-ecosystem/go-grpc-middleware/v2"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func UnaryServerRequestIdInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		return handler(attachRequestID(ctx), req)
	}
}

func StreamServerRequestIdInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrapped := middleware.WrapServerStream(stream)
		wrapped.WrappedContext = attachRequestID(stream.Context())
		return handler(srv, wrapped)
	}
}

func attachRequestID(ctx context.Context) context.Context {
	id := uuid.NewString()
	log := logutil.FromContext(ctx).With(slog.String("request_id", id))
	ctx = logutil.WithLogContext(ctx, log)
	md, ok := metadata.FromIncomingContext(ctx)
	if ok {
		md.Set("request-id", id)
		ctx = metadata.NewIncomingContext(ctx, md)
	}
	return ctx
}
