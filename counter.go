package main

import (
	"context"
	"fmt"
	"os"

	"github.com/redis/go-redis/v9"
)

func NewDB() *redis.Client {
	addr := ":6379"
	if v := os.Getenv("REDIS_ADDR"); v != "" {
		addr = v
	}

	return redis.NewClient(&redis.Options{Addr: addr})
}

type IncreaseCounter = func(ctx context.Context) (counter int, err error)

func Counter(db *redis.Client) IncreaseCounter {
	return func(ctx context.Context) (int, error) {
		next, err := db.Incr(ctx, "counter").Result()
		if err != nil {
			return 0, fmt.Errorf("db: increase counter: %w", err)
		}

		return int(next), nil
	}
}
