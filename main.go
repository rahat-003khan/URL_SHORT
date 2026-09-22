package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"
)

type Link struct {
	Code        string    `json:"code"`
	OriginalURL string    `json:"original_url"`
	Clicks      int       `json:"clicks"`
	CreatedAt   time.Time `json:"created_at"`
}

type Database struct {
	Links map[string]*Link `json:"links"`
}

type App struct {
	db Database
	mu sync.RWMutex
}

type ShortenRequest struct {
	URL   string `json:"url"`
	Alias string `json:"alias"`
}

type ShortenResponse struct {
	ShortURL string `json:"short_url"`
	Code     string `json:"code"`
}

const dbFile = "data.json"

func loadDatabase() Database {
	data, err := os.ReadFile(dbFile)
	if err != nil {
		return Database{Links: make(map[string]*Link)}
	}

	var db Database
	if json.Unmarshal(data, &db) != nil || db.Links == nil {
		return Database{Links: make(map[string]*Link)}
	}

	return db
}

func (a *App) saveDatabase() error {
	data, err := json.MarshalIndent(a.db, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(dbFile, data, 0644)
}

func validURL(value string) bool {
	parsed, err := url.ParseRequestURI(value)
	if err != nil {
		return false
	}

	return (parsed.Scheme == "http" || parsed.Scheme == "https") &&
		parsed.Host != ""
}

func validAlias(alias string) bool {
	if len(alias) < 3 || len(alias) > 20 {
		return false
	}

	for _, c := range alias {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_') {
			return false
		}
	}

	return true
}

func (a *App) generateCode() string {
	const chars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for {
		code := ""
		for i := 0; i < 6; i++ {
			code += string(chars[time.Now().UnixNano()%int64(len(chars))])
			time.Sleep(time.Nanosecond)
		}

		a.mu.RLock()
		_, exists := a.db.Links[code]
		a.mu.RUnlock()

		if !exists {
			return code
		}
	}
}

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
}

func (a *App) shorten(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method == http.MethodOptions {
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req ShortenRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, `{"error":"Invalid JSON"}`, http.StatusBadRequest)
		return
	}

	req.URL = strings.TrimSpace(req.URL)
	req.Alias = strings.TrimSpace(req.Alias)

	if !validURL(req.URL) {
		http.Error(w, `{"error":"Please enter a valid http or https URL"}`, http.StatusBadRequest)
		return
	}

	code := req.Alias

	a.mu.Lock()
	defer a.mu.Unlock()

	if code == "" {
		code = a.generateCode()
	} else {
		if !validAlias(code) {
			http.Error(w, `{"error":"Alias must be 3-20 characters and contain only letters, numbers, - or _"}`, http.StatusBadRequest)
			return
		}

		if _, exists := a.db.Links[code]; exists {
			http.Error(w, `{"error":"This alias is already taken"}`, http.StatusConflict)
			return
		}
	}

	a.db.Links[code] = &Link{
		Code:        code,
		OriginalURL: req.URL,
		Clicks:      0,
		CreatedAt:   time.Now(),
	}

	if err := a.saveDatabase(); err != nil {
		delete(a.db.Links, code)
		http.Error(w, `{"error":"Could not save URL"}`, http.StatusInternalServerError)
		return
	}

	response := ShortenResponse{
		ShortURL: "http://localhost:8080/" + code,
		Code:     code,
	}

	json.NewEncoder(w).Encode(response)
}

func (a *App) redirect(w http.ResponseWriter, r *http.Request) {
	code := strings.TrimPrefix(r.URL.Path, "/")

	if code == "" {
		http.ServeFile(w, r, "public/index.html")
		return
	}

	if strings.HasPrefix(code, "api/") {
		http.NotFound(w, r)
		return
	}

	a.mu.Lock()

	link, exists := a.db.Links[code]
	if exists {
		link.Clicks++
		_ = a.saveDatabase()
	}

	a.mu.Unlock()

	if !exists {
		http.Error(w, "Short URL not found", http.StatusNotFound)
		return
	}

	http.Redirect(w, r, link.OriginalURL, http.StatusFound)
}

func (a *App) stats(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	if r.Method != http.MethodGet {
		http.Error(w, `{"error":"Method not allowed"}`, http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	code := strings.TrimPrefix(r.URL.Path, "/api/stats/")

	a.mu.RLock()
	link, exists := a.db.Links[code]

	if !exists {
		a.mu.RUnlock()
		http.Error(w, `{"error":"Short URL not found"}`, http.StatusNotFound)
		return
	}

	result := *link
	a.mu.RUnlock()

	json.NewEncoder(w).Encode(result)
}

func health(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)

	w.Header().Set("Content-Type", "application/json")

	json.NewEncoder(w).Encode(map[string]string{
		"status":  "ok",
		"message": "URL Shortener API is running",
	})
}

func main() {
	app := &App{
		db: loadDatabase(),
	}

	http.HandleFunc("/shorten", app.shorten)
	http.HandleFunc("/health", health)
	http.HandleFunc("/api/stats/", app.stats)
	http.HandleFunc("/", app.redirect)

	fmt.Println("====================================")
	fmt.Println("       URL SHORTENER API")
	fmt.Println("====================================")
	fmt.Println("Server: http://localhost:8080")
	fmt.Println("Frontend: http://localhost:8080")
	fmt.Println("API: POST /shorten")
	fmt.Println("Stats: GET /api/stats/{code}")
	fmt.Println("====================================")

	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Println("Server error:", err)
	}
}
