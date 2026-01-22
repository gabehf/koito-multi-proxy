package main

import (
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
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
	cfgDir := os.Getenv("KMP_CONFIG_DIR")
	if cfgDir == "" {
		cfgDir = "/etc/kmp"
	}
	if err := loadConfig(path.Join(cfgDir, "config.yml")); err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// 2. Configure Routes
	http.HandleFunc("/", proxyHandler)

	// 3. Start Server
	port := ":4111"
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

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	authHeader := r.Header.Get("Authorization")

	if authHeader == "" || !strings.HasPrefix(authHeader, "Token ") {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	token := strings.TrimPrefix(authHeader, "Token ")

	targetStr, exists := urlMap[token]
	if !exists {
		http.Error(w, "Unauthorized: Invalid token", http.StatusUnauthorized)
		return
	}

	// Parse the target URL
	targetURL, err := url.Parse(targetStr)
	if err != nil {
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}

	// Create a Reverse Proxy
	proxy := httputil.NewSingleHostReverseProxy(targetURL)

	// Update the request to match the target
	r.URL.Host = targetURL.Host
	r.URL.Scheme = targetURL.Scheme
	r.Header.Set("X-Forwarded-Host", r.Header.Get("Host"))
	r.Host = targetURL.Host

	// Note: The original 'Authorization' header is automatically
	// passed along by the proxy unless explicitly removed.

	// Serve the request via the proxy
	proxy.ServeHTTP(w, r)
}
