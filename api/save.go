package handler

import (
	"encoding/json"
	"net/http"
	"time"
)

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
	if EnableCORS(w, r) {
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

	client := GetRedisClient()
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
