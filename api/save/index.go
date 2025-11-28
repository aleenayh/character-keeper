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
func getRedisClient() *redis.Client {
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
func enableCORS(w http.ResponseWriter, r *http.Request) bool {
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

type SaveRequest struct {
	Key  string      `json:"key"`
	Data interface{} `json:"data"`
}

type SaveResponse struct {
	Status  string `json:"status"`
	Key     string `json:"key"`
	SavedAt string `json:"savedAt"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	if enableCORS(w, r) {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req SaveRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	if req.Key == "" {
		http.Error(w, "Key cannot be empty", http.StatusBadRequest)
		return
	}

	if len(req.Key) < 3 {
		http.Error(w, "Key must be at least 3 characters", http.StatusBadRequest)
		return
	}

	if len(req.Key) > 50 {
		http.Error(w, "Key must be 50 characters or less", http.StatusBadRequest)
		return
	}

	jsonData, err := json.Marshal(req.Data)
	if err != nil {
		http.Error(w, "Failed to serialize data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	client := getRedisClient()
	ctx := r.Context()
	redisKey := "user_save:" + req.Key
	expiration := 30 * 24 * time.Hour

	err = client.Set(ctx, redisKey, jsonData, expiration).Err()
	if err != nil {
		http.Error(w, "Failed to save data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := SaveResponse{
		Status:  "saved",
		Key:     req.Key,
		SavedAt: time.Now().Format(time.RFC3339),
	}

	json.NewEncoder(w).Encode(response)
}
