/*
Package cancel provides gRPC middleware interceptors for handling context cancellation.
The package includes both unary and stream interceptors that check if the context or stream context
is already canceled before processing the request. If the context is canceled, these interceptors
stop the processing immediately and return the cancellation error.
*/
package cancel

import (
	"context"

	"github.com/kmmania/er_commonlib/pkg/logger"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// CancelUnaryInterceptor returns a gRPC UnaryServerInterceptor that cancels processing
// if the context is already canceled.
//
// This interceptor checks the context's Done() channel before invoking the handler.
// If the context is canceled, the interceptor logs a warning message including the
// method name and cancellation error, and immediately returns nil and the context's
// error. Otherwise, it proceeds with the handler invocation.
//
// Parameters:
// - logger (logger.Logger): The logger instance used to log cancellation events.
//
// Returns:
//   - grpc.UnaryServerInterceptor: A grpc.UnaryServerInterceptor that can be used with `grpc.Server`'s
//     `UnaryInterceptor` option.
func CancelUnaryInterceptor(logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Check if the context has been canceled before starting processing.
		select {
		case <-ctx.Done():
			// Log the cancellation error with relevant request information.
			logger.Warn("Unary call canceled",
				zap.String("method", info.FullMethod),
				zap.Error(ctx.Err()),
			)
			// If the context is canceled, immediately stop processing.
			return nil, ctx.Err()
		default:
			// If the context is valid, proceed with processing.
			return handler(ctx, req)
		}
	}
}

// CancelStreamInterceptor returns a gRPC StreamServerInterceptor that cancels
// processing if the context associated with the stream is already canceled.
//
// This interceptor retrieves the context from the `grpc.ServerStream` and checks
// its Done() channel before invoking the handler. If the context is canceled,
// the interceptor logs a warning message including the method name and cancellation
// error, and immediately returns the context's error. Otherwise, it proceeds with
// the handler invocation.
//
// Parameters:
// - logger (logger.Logger): The logger instance used to log cancellation events.
//
// Returns:
//   - grpc.StreamServerInterceptor: A grpc.StreamServerInterceptor that can be used with `grpc.Server`'s
//     `StreamInterceptor` option.
func CancelStreamInterceptor(logger logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Get stream context
		ctx := ss.Context()

		// Check if context is already canceled.
		select {
		case <-ctx.Done():
			// Log the cancellation error with relevant stream information.
			logger.Warn("Stream call canceled",
				zap.String("method", info.FullMethod),
				zap.Error(ctx.Err()),
			)
			// If the context is canceled, return the error immediately.
			return ctx.Err()
		default:
			// If the context is valid, continue processing the stream.
			return handler(srv, ss)
		}
	}
}
