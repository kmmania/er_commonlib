/*
Package backoff provides utilities for handling operations with retry mechanisms using exponential backoff.

This package includes functions to retry operations with configurable exponential backoff intervals.
It is designed to handle transient errors by retrying operations with increasing delays between attempts,
up to a maximum elapsed time. The package also provides a utility function to encapsulate common retry logic,
including error handling and logging.

The package is particularly useful in microservices architectures where transient errors (e.g., network issues,
temporary database unavailability) are common, and retrying operations can improve reliability.

Constants:
  - BackoffInitialInterval: The initial delay before the first retry.
  - BackoffMaxInterval: The maximum delay between retries.
  - BackoffMaxElapsedTime: The maximum total time allowed for all retries combined.

Functions:
  - RetryWithExponentialBackOff: Retries an operation with exponential backoff.
  - RetryOperationWithBackoff: Encapsulates retry logic with exponential backoff, including error handling and logging.
*/
package backoff

import (
	"context"
	"errors"
	"time"

	"github.com/kmmania/er_commonlib/pkg/controller"
	"github.com/kmmania/er_commonlib/pkg/logger"

	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"
)

const (
	// BackoffInitialInterval represents the initial interval before the first retry attempt.
	// Set to 100 milliseconds.
	BackoffInitialInterval = 100 * time.Millisecond

	// BackoffMaxInterval represents the maximum interval between retry attempts.
	// Set to 10 seconds.
	BackoffMaxInterval = 10 * time.Second

	// BackoffMaxElapsedTime represents the maximum total elapsed time for all retry attempts.
	// Set to 1 minute.
	BackoffMaxElapsedTime = 1 * time.Minute
)

// RetryWithExponentialBackOff retries the given operation with exponential backoff.
//
// This function uses an exponential backoff strategy to retry the provided operation.
// The retry intervals increase exponentially up to the maximum interval, and the total
// retry time is capped by the maximum elapsed time.
//
// Parameters:
// - ctx (context.Context): The context for managing the lifecycle of the retry operation.
// - operation (func() error): The operation to retry. It should return an error if the operation fails.
//
// Returns:
// - error: An error if the operation fails after all retries, or nil if the operation succeeds.
func RetryWithExponentialBackOff(ctx context.Context, operation func() error) error {
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = BackoffInitialInterval
	expBackoff.MaxInterval = BackoffMaxInterval
	expBackoff.MaxElapsedTime = BackoffMaxElapsedTime

	backoffCtx := backoff.WithContext(expBackoff, ctx)
	return backoff.Retry(operation, backoffCtx)
}

// RetryOperationWithBackoff is a utility function that encapsulates the retry logic with exponential backoff.
//
// This function retries the provided operation using exponential backoff and handles common error scenarios,
// such as permanent errors (e.g., ErrNotFound). It also logs errors and operation completion for debugging
// and monitoring purposes.
//
// Parameters:
// - ctx (context.Context): The context for managing the lifecycle of the retry operation.
// - logger (logger.Logger): The logger instance used for logging errors and operation status.
// - operation (func() error): The operation to retry. It should return an error if the operation fails.
//
// Returns:
// - error: An error if the operation fails after all retries, or nil if the operation succeeds.
func RetryOperationWithBackoff(ctx context.Context, logger logger.Logger, operation func() error) error {
	return RetryWithExponentialBackOff(ctx, func() error {
		err := operation()
		if err != nil {
			// If the error is not found, we stop the retry.
			if errors.Is(err, controller.ErrNotFound) {
				logger.Info("Error is not found, we stop the retry.")
				return backoff.Permanent(err) // Stop the retry
			}
			logger.Error("Operation failed", zap.Error(err))
		}
		logger.Info("Operation complete")
		return err
	})
}
