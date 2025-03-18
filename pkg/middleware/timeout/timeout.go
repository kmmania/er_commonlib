/*
Package timeout provides HTTP and gRPC middleware interceptors for handling request timeouts.
The package includes both unary and stream interceptors that set a timeout for the context of
the request. If the request processing exceeds this timeout, the context is canceled automatically
and returns a timeout error.
*/
package timeout

import (
	"context"
	"errors"

	"github.com/kmmania/er_commonlib/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"time"
)

// TimeoutMiddleware applies a timeout to HTTP requests.
//
// This middleware wraps the provided HTTP handler function with a timeout.  If a request
// exceeds the specified `timeout` duration, the middleware logs a warning message (including
// the HTTP method, path, and timeout duration) and responds with a 504 Gateway Timeout
// status code and a JSON error message. Otherwise, the request is allowed to proceed.
//
// Parameters:
// - timeout (time.Duration): The maximum duration allowed for the HTTP request.
// - logger (logger.Logger):  The logger instance used to log timeout events.
//
// Returns:
// - gin.HandlerFunc: A `gin.HandlerFunc` suitable for use as middleware in a Gin router.
func TimeoutMiddleware(timeout time.Duration, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Create a context with timeout.
		ctx, cancel := context.WithTimeout(c.Request.Context(), timeout)
		defer cancel()

		// Replace the request context with the new context that has a timeout.
		c.Request = c.Request.WithContext(ctx)

		// Continue processing the request.
		c.Next()

		// Check if the context deadline was exceeded.
		if errors.Is(ctx.Err(), context.DeadlineExceeded) {
			logger.Warn("HTTP request timeout exceeded",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path),
				zap.Duration("timeout", timeout))
			c.AbortWithStatusJSON(504, gin.H{"error": "request timeout"})
		}
	}
}

// TimeoutUnaryServerInterceptor applies a timeout to unary gRPC requests.
//
// This interceptor wraps the provided gRPC unary handler function with a timeout. If a
// request exceeds the specified `timeout` duration, the interceptor logs a warning message
// (including the full method name and timeout duration) and returns a
// `codes.DeadlineExceeded` error.  Otherwise, the request is allowed to proceed.
//
// Parameters:
//   - timeout (time.Duration): The maximum duration allowed for the gRPC request.
//   - logger (logger.Logger):  The logger instance used to log timeout events. Should be a logger
//     that supports structured logging (e.g., zap, logrus).
//
// Returns:
//   - grpc.UnaryServerInterceptor: A `grpc.UnaryServerInterceptor` that can be used with `grpc.Server`'s
//     `UnaryInterceptor` option.
func TimeoutUnaryServerInterceptor(timeout time.Duration, logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		// Create a context with timeout.
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()

		// Call the handler and check for deadline exceeded.
		resp, err := handler(ctx, req)
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("gRPC unary call timeout exceeded",
				zap.String("method", info.FullMethod),
				zap.Duration("timeout", timeout))
		}
		return resp, err
	}
}

// TimeoutStreamServerInterceptor applies a timeout to streaming gRPC requests.
//
// This interceptor wraps the provided gRPC stream handler function with a timeout.  If a
// request exceeds the specified `timeout` duration, the interceptor logs a warning message
// (including the full method name and timeout duration) and returns a
// `codes.DeadlineExceeded` error. Otherwise, the request is allowed to proceed.  It uses
// a wrapped `ServerStream` to ensure the timeout context is correctly propagated.
//
// Parameters:
// - timeout (time.Duration): The maximum duration allowed for the gRPC stream.
// - logger (logger.Logger):  The logger instance used to log timeout events.
//
// Returns:
//   - grpc.StreamServerInterceptor: A `grpc.StreamServerInterceptor` that can be used with `grpc.Server`'s
//     `StreamInterceptor` option.
func TimeoutStreamServerInterceptor(timeout time.Duration, logger logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		// Create a context with timeout.
		ctx, cancel := context.WithTimeout(ss.Context(), timeout)
		defer cancel()

		// Wrap the server stream to inject the context with timeout.
		wrapped := &wrappedServerStream{
			ServerStream: ss,
			ctx:          ctx,
		}

		// Call the handler with the new context.
		err := handler(srv, wrapped)
		if errors.Is(err, context.DeadlineExceeded) {
			logger.Warn("gRPC stream call timeout exceeded",
				zap.String("method", info.FullMethod),
				zap.Duration("timeout", timeout))
		}
		return err
	}
}

// wrappedServerStream wraps grpc.ServerStream to inject an updated context with a timeout.
//
// wrappedServerStream embeds the original grpc.ServerStream and overrides the Context method
// to return the updated context with a timeout.
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context returns the updated context with timeout.
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}
