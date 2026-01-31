package main

import (
	"fmt"
	"log"
	"net/http"

	"brunch-card-digital/internal/api"
	"brunch-card-digital/internal/database"
)

// Main entry point for the Brunch Card API
func main() {
	// 1. Database Configuration
	// These values match the postgres-db.yaml setup in Kubernetes
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

	// 2. Run Database Migrations
	// Path relative to the WORKDIR in Dockerfile (/root/)
	migrationPath := "internal/database/migrations.sql"
	err = database.RunMigrations(db, migrationPath)
	if err != nil {
		log.Printf("Migration warning: %v", err)
	}

	// 3. Initialize Repository
	repo := database.NewCardRepository(db)

	// --- ROUTES ---

	// Health check for Kubernetes liveness/readiness probes
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Brunch Card API is running!")
	})

	// API Route to create a new digital loyalty card
	http.HandleFunc("/api/v1/cards", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.CreateCardHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// API Route to get the card status (stamps count, reward status)
	http.HandleFunc("/api/v1/cards/status", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetStatusHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// API Route to generate the QR Code image
	http.HandleFunc("/api/v1/qrcode", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			api.GetQRCodeHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// API Route to add a stamp to a card
	http.HandleFunc("/api/v1/cards/stamp", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			api.StampHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Serve static files (CSS, JS) from the /web directory
	fs := http.FileServer(http.Dir("web"))
	http.Handle("/web/", http.StripPrefix("/web/", fs))

	// Main UI Route to view the digital card in the browser
	http.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Serves the HTML file for the loyalty card
		http.ServeFile(w, r, "web/card.html")
	})

	fmt.Println("Server starting on port 8080...")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
