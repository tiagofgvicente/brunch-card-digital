package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"brunch-card-digital/internal/api"
	"brunch-card-digital/internal/database"

	"github.com/joho/godotenv"
)

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: Environment variable %s is not set.", key)
	}
	return value
}

func main() {
	_ = godotenv.Load()

	dbHost := getEnvOrPanic("DB_HOST")
	dbUser := getEnvOrPanic("DB_USER")
	dbPass := getEnvOrPanic("DB_PASS")
	dbName := getEnvOrPanic("DB_NAME")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	masterPass := getEnvOrPanic("MASTER_PASSWORD")

	db, err := database.ConnectDB(dbHost, dbUser, dbPass, dbName)
	if err != nil {
		log.Fatalf("DB Connection Error: %v", err)
	}
	defer db.Close()

	if err := database.RunMigrations(db, "internal/database/migrations.sql"); err != nil {
		log.Printf("Migration Error: %v", err)
	}

	repo := database.NewCardRepository(db)
	mux := http.NewServeMux()

	// --- MIDDLEWARES ---

	masterAuth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok || user != "master" || pass != masterPass {
				w.Header().Set("WWW-Authenticate", `Basic realm="Master Control"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
	}

	// ESTE É O MIDDLEWARE IMPORTANTE (Agora com o Log corrigido)
	tenantMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Tenta ler do URL (?store=xyz)
			slug := r.URL.Query().Get("store")

			// 2. Se estiver vazio (acesso à raiz /), força "brunch"
			if slug == "" {
				slug = "brunch"
			}

			// 3. Vai buscar à BD
			store, err := repo.GetStoreBySlug(slug)
			if err != nil {
				// AQUI ESTÁ A CORREÇÃO: %v para vermos o erro técnico!
				log.Printf("❌ Critical: Default store '%s' not found. DETALHE DO ERRO: %v", slug, err)
				http.Error(w, "System Error: Store Not Found ("+slug+")", 404)
				return
			}

			// 4. Injeta e segue
			ctx := context.WithValue(r.Context(), "current_store", store)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	storeAuth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			cookie, err := r.Cookie("session_token")
			if err == nil && cookie.Value == "authenticated_admin" {
				next.ServeHTTP(w, r)
				return
			}

			user, pass, ok := r.BasicAuth()
			if !ok || (user != "admin" && user != "loja") || !repo.VerifyPassword(pass) {
				w.Header().Set("WWW-Authenticate", `Basic realm="Store Admin"`)
				http.Error(w, "Unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(w, r)
		}
	}

	// --- ROTAS ---

	mux.HandleFunc("/health", healthHandler)

	mux.HandleFunc("/master", masterAuth(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/master.html")
	}))
	mux.HandleFunc("/api/v1/master/stores", masterAuth(makeHandler(api.MasterListStoresHandler, repo)))

	mux.HandleFunc("/", tenantMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, "web/index.html")
	}))

	mux.HandleFunc("/card", tenantMiddleware(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/card.html")
	}))

	mux.HandleFunc("/api/v1/auth/login", tenantMiddleware(makeHandler(api.LoginHandler, repo)))
	mux.HandleFunc("/api/v1/auth/logout", tenantMiddleware(makeHandler(api.LogoutHandler, repo)))
	mux.HandleFunc("/api/v1/public/stats", tenantMiddleware(makeHandler(api.PublicStatsHandler, repo)))

	mux.HandleFunc("/api/v1/cards", tenantMiddleware(makeHandler(api.CreateCardHandler, repo)))
	mux.HandleFunc("/api/v1/cards/status", tenantMiddleware(makeHandler(api.GetStatusHandler, repo)))
	mux.HandleFunc("/api/v1/cards/stamp", tenantMiddleware(makeHandler(api.StampHandler, repo)))
	mux.HandleFunc("/api/v1/cards/use-reward", tenantMiddleware(makeHandler(api.UseRewardHandler, repo)))
	mux.HandleFunc("/api/v1/qrcode", tenantMiddleware(makeHandler(api.GetQRCodeHandler, repo)))
	mux.HandleFunc("/api/v1/cards/search", tenantMiddleware(makeHandler(api.SearchHandler, repo)))
	mux.HandleFunc("/api/v1/system/config", tenantMiddleware(makeHandler(api.GetSettingsHandler, repo)))

	mux.HandleFunc("/admin", tenantMiddleware(storeAuth(func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "web/admin.html")
	})))

	mux.HandleFunc("/skins.html", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/skins.html") })
	mux.HandleFunc("/settings.html", func(w http.ResponseWriter, r *http.Request) { http.ServeFile(w, r, "web/settings.html") })

	mux.HandleFunc("/api/v1/admin/verify-password", tenantMiddleware(makeHandler(api.VerifyPasswordHandler, repo)))
	mux.HandleFunc("/api/v1/admin/cards", tenantMiddleware(storeAuth(makeHandler(api.ListAllCardsHandler, repo))))
	mux.HandleFunc("/api/v1/admin/reset", tenantMiddleware(storeAuth(makeHandler(api.AdminResetHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update", tenantMiddleware(storeAuth(makeHandler(api.UpdateCardHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update-skin", tenantMiddleware(storeAuth(makeHandler(api.UpdateSkinHandler, repo))))
	mux.HandleFunc("/api/v1/admin/settings", tenantMiddleware(storeAuth(makeHandler(api.UpdateSettingsHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update-consent", tenantMiddleware(storeAuth(makeHandler(api.UpdateConsentHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update-password", tenantMiddleware(storeAuth(makeHandler(api.UpdatePasswordHandler, repo))))

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      loggingMiddleware(mux),
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("🚀 Brunch SaaS Platform started on port %s\n", port)
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
