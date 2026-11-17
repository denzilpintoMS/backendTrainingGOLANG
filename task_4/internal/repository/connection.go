package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openDatabase configures and checks the PostgreSQL connection pool.
func openDatabase(ctx context.Context, databaseURL string) (*gorm.DB, *sql.DB, error) {
	db, err := gorm.Open(postgres.Open(databaseURL), &gorm.Config{
		DisableAutomaticPing: true,
		// Repository errors are logged by the handler; avoid logging SQL values.
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		// Connection parsing errors may contain credentials.
		return nil, nil, errors.New("invalid PostgreSQL connection configuration")
	}
	pool, err := db.DB()
	if err != nil {
		return nil, nil, fmt.Errorf("get PostgreSQL connection pool: %w", err)
	}
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(5)
	pool.SetConnMaxLifetime(30 * time.Minute)
	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, nil, fmt.Errorf("connect to PostgreSQL: %w", err)
	}
	return db, pool, nil
}
