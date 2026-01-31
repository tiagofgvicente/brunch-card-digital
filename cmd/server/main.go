package main

import (
	"fmt"
	"log"
	"net/http"

	// Replace 'youruser' with your actual GitHub username
	"brunch-card-digital/internal/api"
	"brunch-card-digital/internal/database"
)

// Main entry point for the Brunch Card API
func main() {
	// 1. Database Configuration
	// These values match our postgres-db.yaml setup
	dbHost := "postgres-service"
	dbUser := "admin"
	dbPass := "brunch_pass"
	dbName := "loyalty_db"

	log.Println("Connecting to PostgreSQL...")
	db, err := database.ConnectDB(dbHost, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("Critical: Database connection failed: %v", err)
	}
	defer db.Close()

	// 2. Run Migrations
	// The path here is relative to the WORKDIR in the Dockerfile (/root/)
	migrationPath := "internal/database/migrations.sql"
	err = database.RunMigrations(db, migrationPath)
	if err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// 3. Initialize Repository and Handlers
	repo := database.NewCardRepository(db)

	// --- ROUTES ---

	// Health check for Kubernetes liveness/readiness probes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Brunch Card API is running!")
	})

	// API Route to create a new digital card
	http.HandleFunc("/api/v1/cards", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.CreateCardHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Route to get the QR Code image
	http.HandleFunc("/api/v1/qrcode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetQRCodeHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/api/v1/cards/stamp", func(w http.ResponseWriter, r *http.Request) {
		api.StampHandler(w, r, repo)
	})

	// Serve ficheiros estáticos da pasta /web
	http.Handle("/web/", http.StripPrefix("/web/", http.FileServer(http.Dir("web"))))

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
