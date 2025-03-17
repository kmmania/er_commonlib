package logger

import "go.uber.org/zap"

// Logger defines an interface for structured logging.
// This interface aligns with the method signatures of zap.Logger.
type Logger interface {

	// Info logs informational messages that highlight the progress of the application.
	// Structured data should be passed as zap.Field.
	Info(msg string, fields ...zap.Field)

	// Error logs error messages, typically used for recovering from unexpected conditions.
	// Structured data should be passed as zap.Field.
	Error(msg string, fields ...zap.Field)

	// Debug logs messages intended for debugging and diagnostics.
	// Structured data should be passed as zap.Field.
	Debug(msg string, fields ...zap.Field)

	// Warn logs
	Warn(msg string, fields ...zap.Field)

	// Fatal logs critical errors that lead to application termination.
	// Structured data should be passed as zap.Field.
	// The implementation is expected to terminate the application.
	Fatal(msg string, fields ...zap.Field)
}
