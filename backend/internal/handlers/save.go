package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-redis/redis/v8"
)

var RedisClient *redis.Client

type SaveRequest struct {
	Key  string      `json:"key"`
	Data interface{} `json:"data"`
}

type SaveResponse struct {
	Status  string `json:"status"`
	Key     string `json:"key"`
	SavedAt string `json:"savedAt"`
}

type LoadResponse struct {
	Key      string      `json:"key"`
	Data     interface{} `json:"data"`
	LoadedAt string      `json:"loadedAt"`
}

func SaveCharacterData(w http.ResponseWriter, r *http.Request) {
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

	ctx := r.Context()
	redisKey := "user_save:" + req.Key
	expiration := 30 * 24 * time.Hour

	err = RedisClient.Set(ctx, redisKey, jsonData, expiration).Err()
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

func LoadCharacterData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	redisKey := "user_save:" + key

	jsonData, err := RedisClient.Get(ctx, redisKey).Bytes()
	if err != nil {
		if err == redis.Nil {
			http.Error(w, "Save key not found", http.StatusNotFound)
			return
		}

		http.Error(w, "Failed to load data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	var data interface{}
	if err := json.Unmarshal(jsonData, &data); err != nil {
		http.Error(w, "Corrupted data in storage", http.StatusInternalServerError)
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

func DeleteCharacterData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		http.Error(w, "Missing 'key' parameter", http.StatusBadRequest)
		return
	}

	ctx := r.Context()
	redisKey := "user_save:" + key

	err := RedisClient.Del(ctx, redisKey).Err()
	if err != nil {
		http.Error(w, "Failed to delete: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "deleted",
		"key":    key,
	})
}