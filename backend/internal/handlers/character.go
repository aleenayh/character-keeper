package handlers

import (
"encoding/json"
"fmt"
"io"
"net/http"
"net/url"
)

type CharacterResponse struct {
URL string `json:"url"`
Content string `json:"content"`
Error string `json:"error,omitempty"`
}

func GetCharacterData(w http.ResponseWriter, r *http.Request) {
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

      resp, err := http.Get(targetURL)
      if err != nil {
          w.WriteHeader(http.StatusBadGateway)
          json.NewEncoder(w).Encode(CharacterResponse{
              Error: fmt.Sprintf("failed to fetch: %v", err),
          })
          return
      }
      defer resp.Body.Close() 

      body, err := io.ReadAll(resp.Body)
      if err != nil {
          w.WriteHeader(http.StatusInternalServerError)
          json.NewEncoder(w).Encode(CharacterResponse{
              Error: "failed to read response",
          })
          return
      }

      w.Header().Set("Content-Type", "application/json")
      json.NewEncoder(w).Encode(CharacterResponse{
          URL:     targetURL,
          Content: string(body),
      })

}