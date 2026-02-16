package testutil

import (
	"context"
	"fmt"
	"os"
	"testing"

	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupTestDB connects to pointofsale_test database, runs migrations, and
// starts a transaction for test isolation. The transaction is automatically
// rolled back via t.Cleanup, so no data leaks between tests — even when
// multiple test packages run in parallel with go test ./...
//
// NOTE: The test database 'pointofsale_test' must exist before running tests.
// Create it with: CREATE DATABASE pointofsale_test;
func SetupTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	// Load .env from backend directory (ignore error if not found)
	_ = godotenv.Load("../.env")

	// Get database configuration from environment with test defaults
	dbHost := getEnvOrDefault("DB_HOST", "localhost")
	dbPort := getEnvOrDefault("DB_PORT", "5432")
	dbUser := getEnvOrDefault("DB_USER", "pointofsale")
	dbPassword := getEnvOrDefault("DB_PASSWORD", "secret")
	dbSSLMode := getEnvOrDefault("DB_SSLMODE", "disable")

	// Always use test database name
	dbName := "pointofsale_test"

	// Build DSN
	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		dbHost, dbPort, dbUser, dbPassword, dbName, dbSSLMode,
	)

	// Connect via GORM
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	require.NoError(t, err, "failed to connect to test database")

	// Get underlying sql.DB for migrations
	sqlDB, err := db.DB()
	require.NoError(t, err, "failed to get sql.DB")

	// Run goose migrations (idempotent, uses advisory lock for concurrency safety)
	goose.SetDialect("postgres")
	err = goose.Up(sqlDB, "../migrations")
	require.NoError(t, err, "failed to run migrations")

	// Start a transaction for test isolation.
	// Each test gets its own transaction that is rolled back on cleanup,
	// preventing cross-package interference and avoiding deadlocks from TRUNCATE.
	tx := db.Begin()
	require.NoError(t, tx.Error, "failed to begin test transaction")

	t.Cleanup(func() {
		tx.Rollback()
	})

	return tx
}

// CleanupTestDB is a no-op. Cleanup is handled by transaction rollback
// registered in SetupTestDB. Kept for backward compatibility with existing tests.
func CleanupTestDB(t *testing.T, db *gorm.DB) {
	// No-op: transaction rollback in t.Cleanup handles all data cleanup.
}

// SetupTestRedis connects to Redis for testing using DB 1 (separate from dev DB 0).
func SetupTestRedis(t *testing.T) *redis.Client {
	t.Helper()

	// Load .env from backend directory (ignore error if not found)
	_ = godotenv.Load("../.env")

	// Get Redis configuration from environment with test defaults
	redisHost := getEnvOrDefault("REDIS_HOST", "localhost")
	redisPort := getEnvOrDefault("REDIS_PORT", "6379")
	redisPassword := getEnvOrDefault("REDIS_PASSWORD", "")

	// Use DB 1 for tests (dev uses DB 0)
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
		Password: redisPassword,
		DB:       1, // Test database
	})

	// Verify connection
	err := rdb.Ping(context.Background()).Err()
	require.NoError(t, err, "failed to connect to test Redis")

	// Flush the test database to start clean
	err = rdb.FlushDB(context.Background()).Err()
	require.NoError(t, err, "failed to flush test Redis DB")

	return rdb
}

// CleanupTestRedis flushes the test Redis DB.
func CleanupTestRedis(t *testing.T, rdb *redis.Client) {
	t.Helper()

	err := rdb.FlushDB(context.Background()).Err()
	require.NoError(t, err, "failed to flush test Redis DB")
}

// getEnvOrDefault gets an environment variable or returns the default value.
func getEnvOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
