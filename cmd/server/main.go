package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"brunch-card-digital/internal/api"
	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/joho/godotenv"
)

func getEnvOrPanic(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Error: Environment variable %s is not set.", key)
	}
	return value
}

func getStaticDir() string {
	// Tenta no diretório atual
	if _, err := os.Stat("./static"); err == nil {
		return "./static"
	}
	// Tenta subir dois níveis (caso estejas a correr de cmd/server)
	if _, err := os.Stat("../../static"); err == nil {
		return "../../static"
	}
	// Retorna vazio se falhar (vai dar erro, mas permite logar)
	return ""
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
	repo.StartCleanupWorker()
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

	tenantMiddleware := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			slug := r.URL.Query().Get("store")
			if slug == "" {
				slug = "brunch"
			}

			store, err := repo.GetStoreBySlug(slug)
			if err != nil {
				log.Printf("❌ Critical: Default store '%s' not found. DETALHE DO ERRO: %v", slug, err)
				http.Error(w, "System Error: Store Not Found ("+slug+")", 404)
				return
			}

			// --- LÓGICA DE BLOQUEIO (TRIAL/PAGAMENTO) ---
			// Verifica se a data de expiração já passou.
			isExpired := time.Now().After(store.TierExpiration)

			// Se estiver expirado E o status não for "lifetime" ou "paid", bloqueamos a API
			if isExpired && store.Status != "paid" && store.Status != "lifetime" {
				// Se a chamada for para modificar dados (/admin/update, /cards, etc), bloqueia com 403
				if r.Method != http.MethodGet && strings.Contains(r.URL.Path, "/api/v1/admin/") {
					http.Error(w, "Subscrição expirada. Por favor regularize o pagamento.", http.StatusForbidden)
					return
				}

				// Se for um GET (ex: carregar as settings), alteramos temporariamente o status para o Frontend saber que tem de mostrar o Modal
				store.Status = "expired"
			}

			ctx := context.WithValue(r.Context(), "current_store", store)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}

	storeAuth := func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 1. Ler o cookie de sessão
			cookie, err := r.Cookie("session_token")
			if err != nil {
				// Se não tem cookie, manda para login
				http.Redirect(w, r, "/?error=login_required", http.StatusSeeOther)
				return
			}

			// 2. Ir buscar a loja do Contexto (que o tenantMiddleware lá pôs)
			storeCtx := r.Context().Value("current_store")
			if storeCtx == nil {
				http.Error(w, "Store context missing", http.StatusInternalServerError)
				return
			}

			// CORREÇÃO AQUI: Usamos models.Store em vez de database.Store
			currentStore := storeCtx.(*models.Store)

			// 3. SEGURANÇA MÁXIMA (IDOR PROTECTION)
			// Comparamos o ID guardado no Cookie (Login) com o ID da Loja atual (URL)
			if cookie.Value != currentStore.ID {
				log.Printf("⛔ SEGURANÇA: Tentativa de acesso cruzado. User: %s -> Loja: %s", cookie.Value, currentStore.Slug)

				// Redireciona para o login se tentar aceder à loja errada
				http.Redirect(w, r, "/?error=access_denied", http.StatusSeeOther)
				return
			}

			// 4. Tudo bate certo, pode passar
			next.ServeHTTP(w, r)
		}
	}

	// --- CONFIGURAÇÃO INTELIGENTE DE FICHEIROS ESTÁTICOS ---
	// Verifica se a pasta "static" está na diretoria atual (raiz) ou acima (cmd/server)
	staticDir := getStaticDir()
	if staticDir == "" {
		log.Fatal("CRITICAL ERROR: Pasta 'static' não encontrada. Corre o projeto na raiz.")
	}

	log.Printf("📂 Serving static files from: %s", staticDir)
	fs := http.FileServer(http.Dir(staticDir))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// --- ROTAS ---

	mux.HandleFunc("/health", healthHandler)

	// 1. MASTER ROUTES
	mux.HandleFunc("/master", masterAuth(func(w http.ResponseWriter, r *http.Request) {
		htmlPath := "web/master.html"
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			htmlPath = "../../web/master.html"
		}
		http.ServeFile(w, r, htmlPath)
	}))

	// API Master Stores
	mux.HandleFunc("/api/v1/master/stores", masterAuth(makeHandler(func(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
		if r.Method == http.MethodPost {
			api.MasterCreateStoreHandler(w, r, repo)
		} else {
			api.MasterListStoresHandler(w, r, repo)
		}
	}, repo)))

	mux.HandleFunc("/api/v1/master/stores/status", masterAuth(makeHandler(func(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
		if r.Method == http.MethodPost {
			api.MasterToggleStoreHandler(w, r, repo)
		}
	}, repo)))

	mux.HandleFunc("/api/v1/master/stores/update", masterAuth(makeHandler(func(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
		if r.Method == http.MethodPost {
			api.MasterUpdateStoreHandler(w, r, repo)
		}
	}, repo)))

	// API Master Skins (NOVO)
	mux.HandleFunc("/api/v1/master/skins", masterAuth(makeHandler(func(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
		switch r.Method {
		case http.MethodGet:
			api.GetMasterSkinsHandler(w, r, repo)
		case http.MethodPost:
			api.SaveSkinHandler(w, r, repo)
		case http.MethodDelete:
			api.DeleteSkinHandler(w, r, repo)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}, repo)))

	mux.HandleFunc("/api/v1/master/notifications/send", masterAuth(makeHandler(func(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
		if r.Method == http.MethodPost {
			api.MasterSendNotificationHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}, repo)))

	// 2. PUBLIC / TENANT ROUTES
	mux.HandleFunc("/", tenantMiddleware(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}

		htmlPath := "web/index.html"
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			htmlPath = "../../web/index.html"
		}
		http.ServeFile(w, r, htmlPath)
	}))

	mux.HandleFunc("/card", tenantMiddleware(func(w http.ResponseWriter, r *http.Request) {
		htmlPath := "web/card.html"
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			htmlPath = "../../web/card.html"
		}
		http.ServeFile(w, r, htmlPath)
	}))

	// Auth & Stats
	mux.HandleFunc("/api/v1/auth/login", tenantMiddleware(makeHandler(api.LoginHandler, repo)))
	mux.HandleFunc("/api/v1/auth/logout", tenantMiddleware(makeHandler(api.LogoutHandler, repo)))
	mux.HandleFunc("/api/v1/public/stats", tenantMiddleware(makeHandler(api.PublicStatsHandler, repo)))
	mux.HandleFunc("/api/v1/public/register", func(w http.ResponseWriter, r *http.Request) {
		// Não usamos middleware aqui porque é um acesso público (ainda não tem token)
		if r.Method == http.MethodPost {
			api.PublicRegisterHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})
	mux.HandleFunc("/api/v1/public/wallet-register", func(w http.ResponseWriter, r *http.Request) {
		// Acesso público (sem middleware) para registar o cliente Global
		if r.Method == http.MethodPost {
			api.WalletRegisterHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/verify", func(w http.ResponseWriter, r *http.Request) {
		api.WalletVerifyEmailHandler(w, r, repo)
	})

	mux.HandleFunc("/verify-store", func(w http.ResponseWriter, r *http.Request) {
		api.StoreVerifyEmailHandler(w, r, repo)
	})

	mux.HandleFunc("/api/v1/public/wallet-login", func(w http.ResponseWriter, r *http.Request) {
		// Acesso público (sem middleware) para login do cliente Global
		if r.Method == http.MethodPost {
			api.WalletLoginHandler(w, r, repo)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/api/v1/public/wallet-profile", func(w http.ResponseWriter, r *http.Request) {
		api.WalletProfileHandler(w, r, repo)
	})
	mux.HandleFunc("/api/v1/public/wallet-mycards", func(w http.ResponseWriter, r *http.Request) {
		api.WalletMyCardsHandler(w, r, repo)
	})

	mux.HandleFunc("/api/v1/public/wallet-notifications", makeHandler(api.GetNotificationsHandler, repo))
	mux.HandleFunc("/api/v1/public/wallet-notifications/read", makeHandler(api.MarkNotificationsReadHandler, repo))

	// Card Operations
	mux.HandleFunc("/api/v1/cards", tenantMiddleware(makeHandler(api.CreateCardHandler, repo)))
	mux.HandleFunc("/api/v1/cards/status", tenantMiddleware(makeHandler(api.GetStatusHandler, repo)))
	mux.HandleFunc("/api/v1/cards/stamp", tenantMiddleware(makeHandler(api.StampHandler, repo)))
	mux.HandleFunc("/api/v1/cards/use-reward", tenantMiddleware(makeHandler(api.UseRewardHandler, repo)))
	mux.HandleFunc("/api/v1/qrcode", tenantMiddleware(makeHandler(api.GetQRCodeHandler, repo)))
	mux.HandleFunc("/api/v1/cards/search", tenantMiddleware(makeHandler(api.SearchHandler, repo)))

	// System Config & Skins
	mux.HandleFunc("/api/v1/system/config", tenantMiddleware(makeHandler(api.GetSettingsHandler, repo)))
	mux.HandleFunc("/api/v1/system/skins", tenantMiddleware(makeHandler(api.GetAvailableSkinsHandler, repo)))

	// 3. ADMIN ROUTES (STORE SPECIFIC)
	mux.HandleFunc("/admin", tenantMiddleware(storeAuth(func(w http.ResponseWriter, r *http.Request) {
		htmlPath := "web/admin.html"
		if _, err := os.Stat(htmlPath); os.IsNotExist(err) {
			htmlPath = "../../web/admin.html"
		}
		http.ServeFile(w, r, htmlPath)
	})))

	// API Admin
	mux.HandleFunc("/api/v1/admin/verify-password", tenantMiddleware(makeHandler(api.VerifyPasswordHandler, repo)))
	mux.HandleFunc("/api/v1/admin/cards", tenantMiddleware(storeAuth(makeHandler(api.ListAllCardsHandler, repo))))
	mux.HandleFunc("/api/v1/admin/reset", tenantMiddleware(storeAuth(makeHandler(api.AdminResetHandler, repo))))
	mux.HandleFunc("/api/v1/admin/notifications", tenantMiddleware(storeAuth(makeHandler(api.GetAdminNotificationsHandler, repo))))
	mux.HandleFunc("/api/v1/admin/notifications/read", tenantMiddleware(storeAuth(makeHandler(api.MarkAdminNotificationsReadHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update", tenantMiddleware(storeAuth(makeHandler(api.UpdateCardHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update-skin", tenantMiddleware(storeAuth(makeHandler(api.UpdateSkinHandler, repo))))
	mux.HandleFunc("/api/v1/admin/settings", tenantMiddleware(storeAuth(makeHandler(api.UpdateSettingsHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update-consent", tenantMiddleware(storeAuth(makeHandler(api.UpdateConsentHandler, repo))))
	mux.HandleFunc("/api/v1/admin/update-password", tenantMiddleware(storeAuth(makeHandler(api.UpdatePasswordHandler, repo))))
	mux.HandleFunc("/api/v1/admin/scopes", tenantMiddleware(makeHandler(api.GetScopesHandler, repo)))
	mux.HandleFunc("/api/v1/admin/scopes/create", tenantMiddleware(makeHandler(api.CreateScopeHandler, repo)))
	mux.HandleFunc("/api/v1/admin/scopes/toggle", tenantMiddleware(makeHandler(api.ToggleScopeHandler, repo)))
	mux.HandleFunc("/api/v1/admin/scopes/update", tenantMiddleware(makeHandler(api.UpdateScopeHandler, repo)))
	mux.HandleFunc("/api/v1/admin/scopes/delete", tenantMiddleware(makeHandler(api.DeleteScopeHandler, repo)))

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
