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
		log.Printf("Migration warning: %v", err)
	}

	// 4. REPOSITORY
	repo := database.NewCardRepository(db)

	// 5. ROUTING
	mux := http.NewServeMux()

	// --- UI ENDPOINTS ---
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/index.html")
	})
	mux.HandleFunc("/card", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/card.html") })
	mux.HandleFunc("/skins.html", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/skins.html") })
	mux.HandleFunc("/settings.html", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/settings.html") })

	// ROTA ADMIN PROTEGIDA PELA BASE DE DADOS
	mux.HandleFunc("/admin", basicAuth(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/admin.html")
	}, repo))

	// --- API ENDPOINTS ---
	mux.HandleFunc("/health", healthHandler)

	// Public
	mux.HandleFunc("/api/v1/auth/login", makeHandler(api.LoginHandler, repo))
	mux.HandleFunc("/api/v1/auth/logout", makeHandler(api.LogoutHandler, repo))
	mux.HandleFunc("/api/v1/public/stats", makeHandler(api.PublicStatsHandler, repo))
	mux.HandleFunc("/api/v1/cards", makeHandler(api.CreateCardHandler, repo))
	mux.HandleFunc("/api/v1/cards/status", makeHandler(api.GetStatusHandler, repo))
	mux.HandleFunc("/api/v1/cards/stamp", makeHandler(api.StampHandler, repo))
	mux.HandleFunc("/api/v1/cards/use-reward", makeHandler(api.UseRewardHandler, repo))
	mux.HandleFunc("/api/v1/qrcode", makeHandler(api.GetQRCodeHandler, repo))
	mux.HandleFunc("/api/v1/cards/search", makeHandler(api.SearchHandler, repo))
	mux.HandleFunc("/api/v1/system/config", makeHandler(api.GetSettingsHandler, repo))

	// --- NOVOS ENDPOINTS ---
	mux.HandleFunc("/api/v1/admin/verify-password", makeHandler(api.VerifyPasswordHandler, repo))

	// Protected Admin API
	mux.HandleFunc("/api/v1/admin/cards", basicAuth(makeHandler(api.ListAllCardsHandler, repo), repo))
	mux.HandleFunc("/api/v1/admin/reset", basicAuth(makeHandler(api.AdminResetHandler, repo), repo))
	mux.HandleFunc("/api/v1/admin/update", basicAuth(makeHandler(api.UpdateCardHandler, repo), repo))
	mux.HandleFunc("/api/v1/admin/update-skin", basicAuth(makeHandler(api.UpdateSkinHandler, repo), repo))
	mux.HandleFunc("/api/v1/admin/settings", basicAuth(makeHandler(api.UpdateSettingsHandler, repo), repo))

	// NOVOS PROTEGIDOS
	mux.HandleFunc("/api/v1/admin/update-consent", basicAuth(makeHandler(api.UpdateConsentHandler, repo), repo))
	mux.HandleFunc("/api/v1/admin/update-password", basicAuth(makeHandler(api.UpdatePasswordHandler, repo), repo))

	// 6. START SERVER
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

// ATUALIZADO: basicAuth agora verifica a password na Base de Dados
func basicAuth(next http.HandlerFunc, repo *database.CardRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. VERIFICAR SE EXISTE COOKIE VÁLIDO
		cookie, err := r.Cookie("session_token")
		if err == nil && cookie.Value == "authenticated_admin" {
			next.ServeHTTP(w, r)
			return
		}

		// 2. FALLBACK PARA BASIC AUTH (Se o cookie não existir)
		user, pass, ok := r.BasicAuth()
		if !ok || user != "admin" || !repo.VerifyPassword(pass) {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r)
	}
}
