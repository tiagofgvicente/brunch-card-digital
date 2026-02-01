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

	// --- UI ENDPOINTS ---

	// Landing Page / Registration Form
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		// Only handle exactly "/" to avoid matching all sub-paths
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/index.html")
	})

	// Digital Card View (Vue.js App)
	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/card.html")
	})

	// Static Files Support (if needed in the future)
	mux.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	// --- API ENDPOINTS ---

	// Health check for Kubernetes probes
	mux.HandleFunc("/health", healthHandler)

	// Create a new digital card
	mux.HandleFunc("/api/v1/cards", makeHandler(api.CreateCardHandler, repo))

	// Get card status and stamp count
	mux.HandleFunc("/api/v1/cards/status", makeHandler(api.GetStatusHandler, repo))

	// Add a stamp to a card
	mux.HandleFunc("/api/v1/cards/stamp", makeHandler(api.StampHandler, repo))

	// Consume a 10-stamp side reward (Gifts)
	mux.HandleFunc("/api/v1/cards/use-reward", makeHandler(api.UseRewardHandler, repo))

	// Generate QR Code image
	mux.HandleFunc("/api/v1/qrcode", makeHandler(api.GetQRCodeHandler, repo))

	// Nas tuas rotas, protege o admin:
	mux.HandleFunc("/admin", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/admin.html")
	}))

	// API Admin
	// Rota para listar todos os cartões (usada pelo Dashboard)
	mux.HandleFunc("/api/v1/admin/cards", basicAuth(makeHandler(api.ListAllCardsHandler, repo)))
	mux.HandleFunc("/api/v1/admin/reset", basicAuth(makeHandler(api.AdminResetHandler, repo)))

	// Balcão (Público/Staff)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/index.html") })
	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/card.html") })

	// Dashboard Admin (Protegido por Basic Auth)
	mux.HandleFunc("/admin", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/admin.html")
	}))

	// 6. SERVER CONFIGURATION
	server := &http.Server{
		Addr:         ":8080",
		Handler:      loggingMiddleware(mux), // Request logging
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

// No main.go, adiciona esta função
func basicAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()
		// Define aqui o teu login e password de admin
		if !ok || user != "admin" || pass != "brunch2026" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
