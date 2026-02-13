package grpcutil

import (
	"context"
	"fmt"

	"github.com/authenticvision/util-go/logutil"
	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func UnaryServerPanicInterceptor() grpc.UnaryServerInterceptor {
	return recovery.UnaryServerInterceptor(recovery.WithRecoveryHandlerContext(recoveryHandler))
}

func StreamServerPanicInterceptor() grpc.StreamServerInterceptor {
	return recovery.StreamServerInterceptor(recovery.WithRecoveryHandlerContext(recoveryHandler))
}

func recoveryHandler(_ context.Context, p any) error {
	return logutil.NewError(PanicError{Value: p}, "panic", logutil.Stack(5))
}

// PanicError wraps a panic value and is returned when a handler panics.
type PanicError struct {
	Value any
}

func (e PanicError) Error() string {
	return fmt.Sprintf("%v", e.Value)
}

func (e PanicError) Unwrap() error {
	if err, ok := e.Value.(error); ok {
		return err
	}
	return nil
}

func (e PanicError) GRPCStatus() *status.Status {
	return status.New(codes.Internal, e.Error())
}

func (e PanicError) GRPCPublicMessage() PublicMessage {
	return "RPC aborted due to panic"
}
