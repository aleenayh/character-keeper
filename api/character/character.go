package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
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

type CharacterResponse struct {
	URL     string `json:"url"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	if EnableCORS(w, r) {
		return
	}

	if r.Method != http.MethodGet {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(CharacterResponse{
			Error: "method not allowed",
		})
		return
	}

	targetURL := r.URL.Query().Get("url")
	if targetURL == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CharacterResponse{
			Error: "url parameter is required",
		})
		return
	}

	// Validate URL
	if _, err := url.ParseRequestURI(targetURL); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(CharacterResponse{
			Error: "invalid url format",
		})
		return
	}

	// Fetch the URL
	resp, err := http.Get(targetURL)
	if err != nil {
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(CharacterResponse{
			Error: fmt.Sprintf("failed to fetch: %v", err),
		})
		return
	}
	defer resp.Body.Close()

	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(CharacterResponse{
			Error: "failed to read response",
		})
		return
	}

	// Return the fetched content
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(CharacterResponse{
		URL:     targetURL,
		Content: string(body),
	})
}
