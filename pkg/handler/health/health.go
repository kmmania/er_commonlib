// Package health provides a health check endpoint for verifying the availability
// of critical service dependencies such as the database and Redis.
package health

import (
	"context"
	"net/http"
	"time"

	"github.com/kmmania/er_commonlib/pkg/db"
	"github.com/kmmania/er_commonlib/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"
)

// HealthCheckDependencies contains the external dependencies required for the health check.
//
// These typically include a database pool and a Redis client. Each of these can be nil,
// in which case the related health check will be skipped and assumed healthy.
type HealthCheckDependencies struct {
	DBPool      db.DBTX       // Interface-based DB connection for flexibility.
	RedisClient *redis.Client // Redis client instance.
	Logger      logger.Logger // Logger for logging health check outcomes.
}

// MakeHealthzHandler returns a Gin handler for the /healthz endpoint.
//
// This handler performs health checks on configured dependencies such as the database
// and Redis. It responds with a 200 status code if all checks pass, or 503 if one or more fail.
//
// Parameters:
//   - deps: HealthCheckDependencies containing the services to check.
//
// Returns:
//   - gin.HandlerFunc: an HTTP handler function for the /healthz route.
func MakeHealthzHandler(deps HealthCheckDependencies) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Use a short timeout for health checks to avoid hanging on dependency failures
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
		defer cancel()

		dbOk := true
		redisOk := true
		var dbErrMsg, redisErrMsg string

		// --- Check Database Connectivity ---
		if deps.DBPool != nil {
			if err := deps.DBPool.Ping(ctx); err != nil {
				deps.Logger.Warn("Health check: Database ping failed", zap.Error(err))
				dbOk = false
				dbErrMsg = err.Error()
			}
		} else {
			dbOk = true
			deps.Logger.Debug("Health check: Database pool not provided, skipping check.")
		}

		// --- Check Redis Connectivity ---
		if deps.RedisClient != nil {
			if _, err := deps.RedisClient.Ping(ctx).Result(); err != nil {
				deps.Logger.Warn("Health check: Redis ping failed", zap.Error(err))
				redisOk = false
				redisErrMsg = err.Error()
			}
		} else {
			redisOk = true
			deps.Logger.Debug("Health check: Redis client not provided, skipping check.")
		}

		// --- Build the Response ---
		response := gin.H{}
		httpStatus := http.StatusOK

		if dbOk && redisOk {
			response["status"] = "ok"
			response["database"] = "connected"
			response["cache"] = "connected"
		} else {
			httpStatus = http.StatusServiceUnavailable
			response["status"] = "unhealthy"

			if !dbOk {
				response["database"] = "disconnected: " + dbErrMsg
			} else {
				response["database"] = "connected"
			}

			if !redisOk {
				response["cache"] = "disconnected: " + redisErrMsg
			} else {
				response["cache"] = "connected"
			}

			deps.Logger.Warn("Health check failed", zap.Bool("db_ok", dbOk), zap.Bool("redis_ok", redisOk))
		}

		c.JSON(httpStatus, response)
	}
}
