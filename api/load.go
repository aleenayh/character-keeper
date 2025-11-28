package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

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
