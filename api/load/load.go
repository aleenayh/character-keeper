package handler

import (
	"encoding/json"
	"net/http"
	"os"
	"sync"
	"time"

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

type LoadResponse struct {
	Key      string      `json:"key"`
	Data     interface{} `json:"data"`
	LoadedAt string      `json:"loadedAt"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	if EnableCORS(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Key parameter is required", http.StatusBadRequest)
		return
	}

	client := GetRedisClient()
	ctx := r.Context()
	redisKey := "user_save:" + key

	jsonData, err := client.Get(ctx, redisKey).Result()
	if err == redis.Nil {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	} else if err != nil {
		http.Error(w, "Failed to load data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var data interface{}
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		http.Error(w, "Failed to parse data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := LoadResponse{
		Key:      key,
		Data:     data,
		LoadedAt: time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}
