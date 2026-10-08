package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/example/go-redis-rate-limiter/internal/httpapi"
	"github.com/example/go-redis-rate-limiter/internal/limiter"
	"github.com/redis/go-redis/v9"
)

func main() {
	addr := getenv("HTTP_ADDR", ":8080")
	redisAddr := getenv("REDIS_ADDR", "localhost:6379")
	limit := getenvInt("RATE_LIMIT", 5)
	window := getenvDuration("RATE_WINDOW", 10*time.Second)

	rdb := redis.NewClient(&redis.Options{
		Addr: redisAddr,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Fatalf("redis connection failed: %v", err)
	}

	rateLimiter := limiter.New(rdb, limit, window)

	mux := http.NewServeMux()
	mux.Handle("/api", httpapi.RateLimit(rateLimiter)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"message":"request allowed"}`))
	})))

	log.Printf("server listening on %s", addr)
	if err := http.ListenAndServe(addr, mux); err != nil {
		log.Fatal(err)
	}
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getenvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	n, err := strconv.Atoi(value)
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

func getenvDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	d, err := time.ParseDuration(value)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}
