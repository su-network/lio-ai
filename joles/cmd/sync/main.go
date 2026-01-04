package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	
	_ "github.com/mattn/go-sqlite3"
	"lio-ai/internal/repositories"
)

func main() {
	// Connect to database
	db, err := sql.Open("sqlite3", "./data/lio.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create repository (it will get encryption key from env)
	repo := repositories.NewProviderKeyRepository(db)

	// Get all active keys for user 1
	keyResponses, err := repo.GetAllByUser("1")
	if err != nil {
		log.Fatal("Failed to get keys:", err)
	}

	// Build API keys map with decrypted keys
	apiKeys := make(map[string]string)
	for _, keyResp := range keyResponses {
		if keyResp.IsActive {
			fullKey, err := repo.GetByUserAndProvider("1", keyResp.Provider)
			if err != nil {
				log.Printf("Failed to fetch key for %s: %v", keyResp.Provider, err)
				continue
			}
			if fullKey != nil {
				apiKeys[fullKey.Provider] = fullKey.APIKey
				log.Printf("Found %s key (length: %d)", fullKey.Provider, len(fullKey.APIKey))
			}
		}
	}

	// Send to Python backend
	backendURL := os.Getenv("BACKEND_URL")
	if backendURL == "" {
		backendURL = "http://localhost:8000"
	}

	payload := map[string]interface{}{
		"user_id":  "1",
		"api_keys": apiKeys,
	}

	jsonData, _ := json.Marshal(payload)
	resp, err := http.Post(
		backendURL+"/api/v1/models/sync-keys",
		"application/json",
		bytes.NewBuffer(jsonData),
	)

	if err != nil {
		log.Fatal("Failed to sync:", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == 200 {
		log.Println("✓ API keys synced successfully")
	} else {
		log.Printf("Failed to sync: HTTP %d", resp.StatusCode)
	}
}
