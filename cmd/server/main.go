package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"brunch-card-digital/internal/api"
	"brunch-card-digital/internal/database"
)

func main() {
	// 1. DATABASE CONFIGURATION
	// Using the service names defined in your Kubernetes manifests
	dbConfig := struct {
		host, user, pass, name string
	}{
		host: "postgres-service",
		user: "admin",
		pass: "brunch_pass",
		name: "loyalty_db",
	}

	// 2. INITIALIZE DATABASE
	log.Printf("Connecting to database at %s...", dbConfig.host)
	db, err := database.ConnectDB(dbConfig.host, dbConfig.user, dbConfig.pass, dbConfig.name)
	if err != nil {
		log.Fatalf("Critical Error: Failed to connect to DB: %v", err)
	}
	defer db.Close()

	// 3. RUN MIGRATIONS
	// Ensure the table structure exists before handling requests
	migrationPath := "internal/database/migrations.sql"
	if err := database.RunMigrations(db, migrationPath); err != nil {
		log.Printf("Migration warning (continuing...): %v", err)
	}

	// 4. REPOSITORY & HANDLERS INJECTION
	repo := database.NewCardRepository(db)

	// 5. DEFINE HTTP HANDLERS (ROUTING)
	mux := http.NewServeMux()

	// API Endpoints
	mux.HandleFunc("/health", healthHandler)
	mux.HandleFunc("/api/v1/cards", makeHandler(api.CreateCardHandler, repo))
	mux.HandleFunc("/api/v1/cards/status", makeHandler(api.GetStatusHandler, repo))
	mux.HandleFunc("/api/v1/cards/stamp", makeHandler(api.StampHandler, repo))
	mux.HandleFunc("/api/v1/qrcode", makeHandler(api.GetQRCodeHandler, repo))

	// UI Endpoints (Vue.js App)
	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/card.html")
	})

	// 6. SERVER CONFIGURATION
	server := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux), // Adding a simple logger
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("🚀 Brunch API Server started on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

// --- HELPER FUNCTIONS & MIDDLEWARE ---

// healthHandler for K8s Liveness/Readiness probes
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

// makeHandler is a wrapper to inject the repository into handlers without repeating logic
func makeHandler(fn func(http.ResponseWriter, *http.Request, *database.CardRepository), repo *database.CardRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, repo)
	}
}

// loggingMiddleware prints basic info for every request to help debugging
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}
