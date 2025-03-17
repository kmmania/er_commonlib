/*
Package db provides utilities for managing database connections.

It contains structures and functions to configure and establish connections
to a PostgreSQL database using the pgxpool package.

The package includes:
  - Config: A structure for holding database connection information.
  - NewDBPool: A function to create a new database connection pool.
  - buildDSN: A helper function to construct the Data Source Name (DSN) for connecting to the database.
*/
package db

import (
	"context"
	"fmt"
	"time"

	"github.com/kmmania/er_backend/er_lib/pkg/logger"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// Config contains the database connection information.
type Config struct {
	User     string // Database username
	Password string // Database password
	Host     string // Host address for the database
	Port     int    // Port number for the database
	DBName   string // Name of the database
	SSLMode  string // SSL mode for the database connection
}

// DBTX defines the methods for executing SQL queries.  It's designed to be
// compatible with the database operations provided by pgxpool.Pool, allowing
// for easier testing and abstraction of database access.  Implementations of
// this interface can wrap different database connection pools or even mock
// database connections for testing.
type DBTX interface {

	// Exec executes a SQL command.
	Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error)

	// Query executes a SQL query that returns rows.
	Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error)

	// QueryRow executes a SQL query that returns a single row.
	QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row

	// Ping verifies a connection to the database is still alive.
	Ping(ctx context.Context) error

	// Close releases all resources associated with the DBTX.  For pooled
	// connections, this typically returns the connections to the pool.  For
	// non-pooled connections, this typically closes the underlying connection.
	Close()
}

// PoolWrapper encapsulates a pgxpool.Pool and implements the DBTX interface.
// It provides a way to use pgxpool.Pool with code that expects a DBTX.
type PoolWrapper struct {
	pool *pgxpool.Pool
}

// NewPoolWrapper creates a new PoolWrapper.
func NewPoolWrapper(pool *pgxpool.Pool) *PoolWrapper {
	return &PoolWrapper{pool: pool}
}

// Exec implements the Exec method of the DBTX interface. It executes a SQL
// command using the underlying pgxpool.Pool.
func (pw *PoolWrapper) Exec(ctx context.Context, sql string, arguments ...interface{}) (pgconn.CommandTag, error) {
	return pw.pool.Exec(ctx, sql, arguments...)
}

// Query implements the Query method of the DBTX interface. It executes a SQL
// query that returns rows using the underlying pgxpool.Pool.
func (pw *PoolWrapper) Query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	return pw.pool.Query(ctx, sql, args...)
}

// QueryRow implements the QueryRow method of the DBTX interface. It executes a
// SQL query that returns a single row using the underlying pgxpool.Pool.
func (pw *PoolWrapper) QueryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	return pw.pool.QueryRow(ctx, sql, args...)
}

// Ping implements the Ping method of the DBTX interface. It verifies a
// connection to the database is still alive using the underlying pgxpool.Pool.
func (pw *PoolWrapper) Ping(ctx context.Context) error {
	return pw.pool.Ping(ctx)
}

// Close implements the Close method of the DBTX interface. It closes the
// underlying pgxpool.Pool, releasing all connections back to the pool.
func (pw *PoolWrapper) Close() {
	pw.pool.Close()
}

// NewDBPool creates a new connection pool to the database using pgxpool.
// It establishes a connection, pings the database to ensure it's reachable,
// and returns a DBTX interface wrapping the pool. It uses a timeout for the
// initial connection attempt. If any error occurs during connection or
// pinging, the function logs a fatal error and returns an error.
//
// Parameters:
// - config (Config): The database configuration struct.
// - logger (logger.Logger): The logger for recording information and errors.
//
// Returns:
// - DBTX: A DBTX interface wrapping the connection pool.
// - error: An error if the connection or ping fails.
func NewDBPool(config Config, logger logger.Logger) (DBTX, error) {
	// Create a context with a timeout to manage long connections
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Initialize the connection with pgxpool using the DSN built from the config
	pool, err := pgxpool.New(ctx, BuildDSN(config, logger))
	if err != nil {
		logger.Fatal("Error connecting to database", zap.Error(err))
		return nil, err
	}

	// Test the connection with a simple query
	if err := pool.Ping(ctx); err != nil {
		logger.Fatal("Error pinging database", zap.Error(err))
		return nil, err
	}

	logger.Info("Successfully connected to PostgreSQL")
	return NewPoolWrapper(pool), nil
}

// BuildDSN constructs the Data Source Name (DSN) from the database
// configuration information. It logs the DSN components (excluding the
// password) for informational purposes. The DSN is returned as a string.
//
// Parameters:
// - config (Config): The database configuration struct.
// - logger (logger.Logger): The logger for recording information.
//
// Returns:
// - string: The constructed DSN.
func BuildDSN(config Config, logger logger.Logger) string {
	logger.Info("DSN successfully build", zap.String("host", config.Host), zap.Int("port", config.Port))

	return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s",
		config.User,
		config.Password,
		config.Host,
		config.Port,
		config.DBName,
		config.SSLMode,
	)
}
