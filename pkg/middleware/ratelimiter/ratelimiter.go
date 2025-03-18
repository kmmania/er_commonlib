/*
Package ratelimiter provides HTTP and gRPC middleware interceptors for handling rate limiting.
The package includes both unary and stream interceptors that enforce rate limits on incoming requests.
When the rate limit is exceeded, the interceptor responds with a rate-limit error.
*/
package ratelimiter

import (
	"context"

	"github.com/kmmania/er_commonlib/pkg/logger"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// RateLimiterHTTP returns a rate limiting middleware for HTTP requests.
//
// This middleware uses the provided `rate.Limiter` to control the rate of incoming
// HTTP requests. If a request exceeds the configured rate limit, the middleware
// logs a warning message (including the HTTP method and path) and responds with
// a 429 Too Many Requests status code, along with a JSON error message.  Otherwise,
// it allows the request to proceed to the next handler in the chain.
//
// Parameters:
// - rl (*rate.Limiter): The `rate.Limiter` instance to use for rate limiting.
// - logger (logger.Logger): The logger instance used to log rate limiting events.
//
// Returns:
// - gin.HandlerFunc: A `gin.HandlerFunc` that can be used as middleware in a Gin router.
func RateLimiterHTTP(rl *rate.Limiter, logger logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !rl.Allow() {
			// Log the rate limit exceed
			logger.Warn("HTTP rate limit exceeded",
				zap.String("method", c.Request.Method),
				zap.String("path", c.Request.URL.Path))
			// Respond with a 429 status code
			c.AbortWithStatusJSON(429, gin.H{"error": "Too many requests"})
			return
		}
		// Continue processing the request
		c.Next()
	}
}

// RateLimiterUnaryInterceptor returns a gRPC UnaryServerInterceptor that limits
// the rate of incoming unary gRPC requests.
//
// This interceptor uses the provided `rate.Limiter` to control the rate of incoming
// unary gRPC requests. If a request exceeds the configured rate limit, the
// interceptor logs a warning message (including the full method name) and returns
// a `codes.ResourceExhausted` error with the message "too many requests". Otherwise,
// it allows the request to proceed to the handler.
//
// Parameters:
// - rl (*rate.Limiter): The `rate.Limiter` instance to use for rate limiting.
// - logger (logger.Logger): The logger instance used to log rate limiting events.
//
// Returns:
//   - grpc.UnaryServerInterceptor: A `grpc.UnaryServerInterceptor` that can be used with `grpc.Server`'s
//     `UnaryInterceptor` option.
func RateLimiterUnaryInterceptor(rl *rate.Limiter, logger logger.Logger) grpc.UnaryServerInterceptor {
	return func(
		ctx context.Context,
		req interface{},
		info *grpc.UnaryServerInfo,
		handler grpc.UnaryHandler,
	) (interface{}, error) {
		if !rl.Allow() {
			// Log the rate limit exceed
			logger.Warn("gRPC unary rate limit exceeded", zap.String("method", info.FullMethod))
			// Return a ResourceExhausted error
			return nil, status.Error(codes.ResourceExhausted, "too many requests")
		}

		// Proceed with the handler
		return handler(ctx, req)
	}
}

// RateLimiterStreamInterceptor returns a gRPC StreamServerInterceptor that limits
// the rate of incoming streaming gRPC requests.
//
// This interceptor uses the provided `rate.Limiter` to control the rate of incoming
// streaming gRPC requests. If a request exceeds the configured rate limit, the
// interceptor logs a warning message (including the full method name) and returns
// a `codes.ResourceExhausted` error with the message "too many requests: rate limiting on stream".
// Otherwise, it allows the request to proceed to the handler.
//
// Parameters:
// - rl (*rate.Limiter): The `rate.Limiter` instance to use for rate limiting.
// - logger (logger.Logger): The logger instance used to log rate limiting events.
//
// Returns:
//   - grpc.StreamServerInterceptor: A `grpc.StreamServerInterceptor` that can be used with `grpc.Server`'s
//     `StreamInterceptor` option.
func RateLimiterStreamInterceptor(rl *rate.Limiter, logger logger.Logger) grpc.StreamServerInterceptor {
	return func(
		srv interface{},
		ss grpc.ServerStream,
		info *grpc.StreamServerInfo,
		handler grpc.StreamHandler,
	) error {
		if !rl.Allow() {
			// Log the rate limit exceed
			logger.Warn("gRPC stream rate limit exceeded", zap.String("method", info.FullMethod))
			// Return a ResourceExhausted error
			return status.Errorf(codes.ResourceExhausted, "too many requests: rate limiting on stream")
		}

		// Proceed with the handler
		return handler(srv, ss)
	}
}
