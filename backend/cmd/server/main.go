package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/aleenayh/character-keeper/internal/handlers"
	"github.com/go-redis/redis/v8"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load("../../../.env")
	if err != nil {
		log.Println("No .env file found, using environment variables")
	}
	redisURL := os.Getenv("UPSTASH_REDIS_URL")
	if redisURL == "" {
		log.Println("WARNING: UPSTASH_REDIS_URL not set — save/load features will be unavailable")
	} else {
		opt, err := redis.ParseURL(redisURL)
		if err != nil {
			log.Printf("WARNING: Failed to parse Redis URL: %v — save/load features will be unavailable", err)
		} else {
			client := redis.NewClient(opt)
			ctx := context.Background()
			_, err = client.Ping(ctx).Result()
			if err != nil {
				log.Printf("WARNING: Failed to connect to Redis: %v — save/load features will be unavailable", err)
			} else {
				handlers.RedisClient = client
				fmt.Println("Successfully connected to Redis")
			}
		}
	}

	mux := http.NewServeMux()

	mux.HandleFunc("GET /api/character", handlers.GetCharacterData)
	mux.HandleFunc("POST /api/save", handlers.SaveCharacterData)
	mux.HandleFunc("GET /api/load", handlers.LoadCharacterData)
	mux.HandleFunc("DELETE /api/delete", handlers.DeleteCharacterData)
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "OK")
	})

	handler := enableCORS(mux)

	port := ":8080"
	fmt.Printf("Server starting on http://localhost%s\n", port)

	if err := http.ListenAndServe(port, handler); err != nil {
		log.Fatal(err)
	}
}

func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Allow localhost for development and your Vercel domain for production
		allowedOrigins := map[string]bool{
			"http://localhost:4200":                    true,
			"https://character-keeper.vercel.app":      true,
		}

		if allowedOrigins[origin] {
			w.Header().Set("Access-Control-Allow-Origin", origin)
		}

		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
		w.Header().Set("Access-Control-Allow-Credentials", "true")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}