package handler

import (
	"encoding/json"
	"net/http"
)

type DeleteResponse struct {
	Status string `json:"status"`
	Key    string `json:"key"`
}

func Handler(w http.ResponseWriter, r *http.Request) {
	// Handle CORS
	if EnableCORS(w, r) {
		return
	}

	if r.Method != http.MethodDelete {
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

	result, err := client.Del(ctx, redisKey).Result()
	if err != nil {
		http.Error(w, "Failed to delete data: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if result == 0 {
		http.Error(w, "Key not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	response := DeleteResponse{
		Status: "deleted",
		Key:    key,
	}

	json.NewEncoder(w).Encode(response)
}
