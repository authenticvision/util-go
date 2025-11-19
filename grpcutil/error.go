package grpcutil

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"strings"

	"github.com/authenticvision/util-go/logutil"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/protoadapt"
)

// PublicMessage is a string sent to the client as-is. This prevents accidentally forwarding a
// non-constant string, e.g. an error message that contains confidential data, to clients.
type PublicMessage string

// PublicError can appear in an error chain to provide a gRPC status to be sent as-is.
// Only the first PublicError in an error chain is used.
type PublicError interface {
	GRPCPublicStatus(ctx context.Context) *status.Status
}

func Err(err error, code codes.Code, msg PublicMessage, details ...protoadapt.MessageV1) error {
	return &grpcError{err: err, code: code, msg: msg, details: details}
}

func ErrInvalidArgument(err error, msg PublicMessage, details ...protoadapt.MessageV1) error {
	return Err(err, codes.InvalidArgument, msg, details...)
}

func ErrNotFound(err error, msg PublicMessage, details ...protoadapt.MessageV1) error {
	return Err(err, codes.NotFound, msg, details...)
}

func ErrPermissionDenied(err error, msg PublicMessage, details ...protoadapt.MessageV1) error {
	return Err(err, codes.PermissionDenied, msg, details...)
}

func ErrInternal(err error, msg PublicMessage, details ...protoadapt.MessageV1) error {
	return Err(err, codes.Internal, msg, details...)
}

func ErrUnavailable(err error, msg PublicMessage, details ...protoadapt.MessageV1) error {
	return Err(err, codes.Unavailable, msg, details...)
}

type grpcError struct {
	err     error
	code    codes.Code
	msg     PublicMessage
	details []protoadapt.MessageV1
}

func (e grpcError) Error() string {
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "grpc status %v", e.code)
	if e.msg != "" {
		_, _ = fmt.Fprintf(&sb, ", %v", e.msg)
	}
	if e.err != nil {
		_, _ = fmt.Fprintf(&sb, ": %v", e.err)
	}
	return sb.String()
}

func (e grpcError) Unwrap() error {
	return e.err
}

// GRPCStatus returns the status for logging prior to obfuscation.
func (e grpcError) GRPCStatus() *status.Status {
	var msg string
	if e.msg != "" && e.err != nil {
		msg = fmt.Sprintf("%v: %v", e.msg, e.err)
	} else if e.msg != "" {
		msg = string(e.msg)
	} else if e.err != nil {
		msg = e.err.Error()
	}
	return status.New(e.code, msg)
}

// GRPCPublicStatus returns the status that's passed to lower levels by obfuscateError.
func (e grpcError) GRPCPublicStatus(ctx context.Context) *status.Status {
	st := status.New(e.code, string(e.msg))
	if len(e.details) != 0 {
		if stDetails, err := st.WithDetails(e.details...); err != nil {
			log := logutil.FromContext(ctx)
			log.Error("failed to attach details to gRPC public status",
				logutil.Err(err),
				slog.String("error_public_message", string(e.msg)))
		} else {
			st = stDetails
		}
	}
	return st
}

func UnaryServerErrorObfuscationInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		resp, err := handler(ctx, req)
		return resp, obfuscateError(ctx, err)
	}
}

func StreamServerErrorObfuscationInterceptor() grpc.StreamServerInterceptor {
	return func(srv any, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		return obfuscateError(stream.Context(), handler(srv, stream))
	}
}

// obfuscateError replaces error messages by generic messages, unless the error is specifically
// tagged as safe for public consumption, or a custom message is provided. The error's gRPC code,
// if any, is always forwarded as-is.
// Errors obfuscated here are not gone as long as a logging and request ID middleware are in place,
// because the original error is usually logged prior to obfuscation.
func obfuscateError(ctx context.Context, err error) (result error) {
	if err == nil {
		return nil
	}

	// we're probably below the panic interceptor and call into user code, so handle panics too
	defer func() {
		if r := recover(); r != nil {
			log := logutil.FromContext(ctx)
			log.Error("panic in gRPC error interceptor", logutil.Err(err), logutil.Stack(1))
			result = status.Error(codes.Internal, "RPC aborted due to panic")
		}
	}()

	// escape hatch to let errors return any status to clients
	var publicErr PublicError
	if errors.As(err, &publicErr) {
		publicStatus := publicErr.GRPCPublicStatus(ctx)
		if publicStatus != nil {
			return publicStatus.Err()
		}
	}

	// generic handler for errors that were not created through this package
	var code codes.Code
	if st, ok := status.FromError(err); ok {
		code = st.Code()
	} else if errors.Is(err, context.Canceled) {
		code = codes.Canceled
	} else if errors.Is(err, context.DeadlineExceeded) {
		code = codes.DeadlineExceeded
	} else {
		code = codes.Unknown
	}

	return status.Error(code, "no error message available, please refer to logs with request ID")
}
