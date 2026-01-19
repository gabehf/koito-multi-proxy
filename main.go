package main

import (
	"log"
	"net/http"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config structure to match the yaml file
type Config struct {
	Mappings []struct {
		Token string `yaml:"token"`
		URL   string `yaml:"url"`
	} `yaml:"mappings"`
}

// Global map to store token -> url associations for fast lookup
var urlMap map[string]string

func main() {
	// 1. Load and Parse Config
	if err := loadConfig("config.yml"); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Configure Routes
	http.HandleFunc("/", redirectHandler)

	// 3. Start Server
	port := ":8080"
	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}

func loadConfig(filename string) error {
	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	// Unmarshal YAML
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return err
	}

	// Build the map
	urlMap = make(map[string]string)
	for _, m := range cfg.Mappings {
		urlMap[m.Token] = m.URL
	}

	log.Printf("Loaded %d token mappings", len(urlMap))
	return nil
}

func redirectHandler(w http.ResponseWriter, r *http.Request) {
	// Get the Authorization header
	authHeader := r.Header.Get("Authorization")

	// Check if header is missing or doesn't start with "Token "
	if authHeader == "" || !strings.HasPrefix(authHeader, "Token ") {
		log.Printf("Warning: Request received with missing or malformed Authorization header from %s", r.RemoteAddr)
		http.Error(w, "Unauthorized: Missing or malformed token", http.StatusUnauthorized)
		return
	}

	// Extract the actual token string (remove "Token " prefix)
	token := strings.TrimPrefix(authHeader, "Token ")

	// Lookup the token in our map
	targetURL, exists := urlMap[token]
	if !exists {
		log.Printf("Warning: Request received with unknown token: %s", token)
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	// Perform the redirect
	// StatusFound (302) is generally used for temporary redirects
	http.Redirect(w, r, targetURL, http.StatusFound)
}
