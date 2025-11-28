package handler

import (
	"net/http"
	"os"
	"sync"

	"github.com/go-redis/redis/v8"
)

var (
	redisClient *redis.Client
	once        sync.Once
)

// GetRedisClient initializes and returns a Redis client (singleton)
func GetRedisClient() *redis.Client {
	once.Do(func() {
		redisURL := os.Getenv("UPSTASH_REDIS_URL")
		if redisURL == "" {
			panic("UPSTASH_REDIS_URL environment variable is not set")
		}

		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			panic("Failed to parse Redis URL: " + err.Error())
		}

		redisClient = redis.NewClient(opt)
	})
	return redisClient
}

// EnableCORS adds CORS headers to the response
func EnableCORS(w http.ResponseWriter, r *http.Request) bool {
	origin := r.Header.Get("Origin")

	allowedOrigins := map[string]bool{
		"http://localhost:4200":               true,
		"https://character-keeper.vercel.app": true,
	}

	if allowedOrigins[origin] {
		w.Header().Set("Access-Control-Allow-Origin", origin)
	}

	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Credentials", "true")

	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return true
	}

	return false
}
