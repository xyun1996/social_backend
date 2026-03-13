package db

import (
	"context"
	"database/sql"

	"github.com/redis/go-redis/v9"
)

// OpenMySQL opens and validates a shared MySQL connection.
func OpenMySQL(ctx context.Context, config MySQLConfig) (*sql.DB, error) {
	sqlDB, err := sql.Open("mysql", config.DSN())
	if err != nil {
		return nil, err
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		_ = sqlDB.Close()
		return nil, err
	}
	return sqlDB, nil
}

// OpenRedis opens and validates a shared Redis client.
func OpenRedis(ctx context.Context, config RedisConfig) (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Username: config.Username,
		Password: config.Password,
		DB:       config.DB,
	})
	if err := client.Ping(ctx).Err(); err != nil {
		_ = client.Close()
		return nil, err
	}
	return client, nil
}
