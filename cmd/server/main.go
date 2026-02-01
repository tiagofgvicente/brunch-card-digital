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
	migrationPath := "internal/database/migrations.sql"
	if err := database.RunMigrations(db, migrationPath); err != nil {
		log.Printf("Migration warning (continuing...): %v", err)
	}

	// 4. REPOSITORY & HANDLERS INJECTION
	repo := database.NewCardRepository(db)

	// 5. DEFINE HTTP HANDLERS (ROUTING)
	mux := http.NewServeMux()

	// --- UI ENDPOINTS (Frontend) ---
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/index.html")
	})

	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/card.html")
	})

	mux.HandleFunc("/admin", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/admin.html")
	}))

	// --- API ENDPOINTS (Backend) ---
	mux.HandleFunc("/health", healthHandler)

	// Public API
	mux.HandleFunc("/api/v1/cards", makeHandler(api.CreateCardHandler, repo))
	mux.HandleFunc("/api/v1/cards/status", makeHandler(api.GetStatusHandler, repo))
	mux.HandleFunc("/api/v1/cards/stamp", makeHandler(api.StampHandler, repo))
	mux.HandleFunc("/api/v1/cards/use-reward", makeHandler(api.UseRewardHandler, repo))
	mux.HandleFunc("/api/v1/qrcode", makeHandler(api.GetQRCodeHandler, repo))
	mux.HandleFunc("/api/v1/cards/search", makeHandler(api.SearchHandler, repo))

	// Protected Admin API
	mux.HandleFunc("/api/v1/admin/cards", basicAuth(makeHandler(api.ListAllCardsHandler, repo)))
	mux.HandleFunc("/api/v1/admin/reset", basicAuth(makeHandler(api.AdminResetHandler, repo)))
	mux.HandleFunc("/api/v1/admin/update", basicAuth(makeHandler(api.UpdateCardHandler, repo)))

	// 6. SERVER CONFIGURATION
	server := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Println("🚀 Brunch API Server started on port 8080")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server startup failed: %v", err)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	fmt.Fprintf(w, "OK")
}

func makeHandler(fn func(http.ResponseWriter, *http.Request, *database.CardRepository), repo *database.CardRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fn(w, r, repo)
	}
}

func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %v", r.Method, r.URL.Path, time.Since(start))
	})
}

func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || pass != "brunch2026" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
