package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

func getCurrentStore(r *http.Request) *models.Store {
	store, ok := r.Context().Value("current_store").(*models.Store)
	if !ok {
		return nil
	}
	return store
}

// --- MASTER HANDLERS (Create Store) ---

func MasterCreateStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.Store
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	log.Printf("🛠️ Criando loja: %s (User: %s)", req.Name, req.AdminUsername)

	if req.AdminUsername == "" || req.AdminEmail == "" {
		http.Error(w, "Username and Email are required", 400)
		return
	}

	if err := repo.CreateStore(req); err != nil {
		log.Printf("❌ Falha ao criar loja: %v", err)
		http.Error(w, "DB Error: "+err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func MasterListStoresHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	stores, err := repo.GetAllStores()
	if err != nil {
		http.Error(w, "DB Error", 500)
		return
	}
	json.NewEncoder(w).Encode(stores)
}

// --- LOGIN GLOBAL INTELIGENTE ---

func LoginHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Ler o JSON (Identifier = Email ou Username)
	fmt.Println("!!! DEBUG: O HANDLER DE LOGIN FOI CHAMADO !!!")
	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	// Se o JSON estiver mal formatado, abortamos logo
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato inválido", http.StatusBadRequest)
		return
	}

	log.Printf("🌍 GLOBAL LOGIN: Tentativa para '%s'", req.Identifier)

	// ---------------------------------------------------------
	// 2. É O MASTER ADMIN? (Backdoor para ti)
	// ---------------------------------------------------------
	if req.Identifier == "master" && req.Password == "superadmin2026" {
		log.Println("✅ Login Master")

		// Criar cookie de sessão master
		http.SetCookie(w, &http.Cookie{
			Name: "session_token", Value: "master_access", Path: "/", HttpOnly: true, Expires: time.Now().Add(12 * time.Hour),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"role": "master", "redirect": "/master", // Redireciona para o painel secreto
		})
		return
	}

	// ---------------------------------------------------------
	// 3. É UM DONO DE LOJA? (Verifica Email + Password Encriptada)
	// ---------------------------------------------------------
	// Usamos a função AuthenticateStore que criámos no repository.go
	store, err := repo.AuthenticateStore(req.Identifier, req.Password)

	if err == nil && store != nil {
		log.Printf("✅ Login Store Admin: '%s' -> %s", req.Identifier, store.Slug)

		// Criar Cookie de Sessão (Guardamos o ID da loja no cookie)
		http.SetCookie(w, &http.Cookie{
			Name:     "session_token",
			Value:    store.ID, // Guardamos o ID real para saber quem é
			Path:     "/",
			HttpOnly: true,
			Expires:  time.Now().Add(24 * time.Hour),
		})

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"role": "admin",
			// IMPORTANTE: O landing.js vai enviar o user para aqui
			// Se o teu dashboard estiver em /?store=slug, mantém assim:
			"redirect": fmt.Sprintf("/admin?store=%s", store.Slug),
		})
		return
	}

	// ---------------------------------------------------------
	// 4. É UM CLIENTE FINAL? (Carteira / Wallet)
	// ---------------------------------------------------------
	// Nota: Esta parte mantém-se igual à tua lógica antiga para clientes
	slug, cardID, err := repo.GetStoreSlugByCustomerEmail(req.Identifier)
	if err == nil && slug != "" {
		log.Printf("✅ Login Cliente: Redirecionar para %s", slug)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"role":     "customer",
			"redirect": fmt.Sprintf("/card?store=%s&id=%s", slug, cardID),
		})
		return
	}

	// ---------------------------------------------------------
	// 5. FALHA TOTAL
	// ---------------------------------------------------------
	log.Println("❌ Falha de Login: Credenciais incorretas")
	http.Error(w, "Email ou Password incorretos", http.StatusUnauthorized)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
	w.WriteHeader(http.StatusOK)
}

// --- RESTO DA API (Com suporte a Âmbitos) ---

func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	newCard := models.LoyaltyCard{
		ID:                uuid.New().String(),
		StoreID:           store.ID,
		CustomerID:        req.CustomerID,
		LastName:          req.LastName,
		Email:             req.Email,
		Phone:             req.Phone,
		NIF:               req.NIF,
		Design:            req.Design,
		ScopeID:           req.ScopeID, // Mapeado diretamente do JSON
		RgpdAccepted:      req.RgpdAccepted,
		MarketingAccepted: req.MarketingAccepted,
	}

	// SISTEMA DE SEGURANÇA: Se a interface falhar e não enviar um ScopeID
	// O backend procura o "Geral/Principal" dessa loja e atribui automaticamente.
	if newCard.ScopeID == "" {
		scopes, _ := repo.GetStoreScopes(store.ID)
		for _, s := range scopes {
			if s.IsMain {
				newCard.ScopeID = s.ID
				break
			}
		}
	}

	if err := repo.SaveCard(newCard); err != nil {
		// Este erro vai disparar se o SQL UNIQUE(store_id, scope_id, email) for violado
		log.Printf("Erro ao gravar cartão: %v", err)
		http.Error(w, "Cliente já possui este cartão", 400)
		return
	}

	savedCard, _ := repo.GetCardByID(newCard.ID)
	json.NewEncoder(w).Encode(savedCard)
}

func ListAllCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	cards, _ := repo.GetAllCards(store.ID)
	json.NewEncoder(w).Encode(cards)
}

func SearchHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	cards, _ := repo.SearchCards(store.ID, query)
	json.NewEncoder(w).Encode(cards)
}

func GetStatusHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	id := r.URL.Query().Get("id")

	// 1. Tenta procurar um cartão de loja normal
	card, err := repo.GetCardByID(id)
	if err == nil {
		json.NewEncoder(w).Encode(card)
		return
	}

	// 2. Se falhar, procura na base de dados Global (Wallet)
	gu, err := repo.GetGlobalUserByID(id)
	if err == nil {
		// Devolve os dados formatados como se fossem um "Cartão" sem selos
		json.NewEncoder(w).Encode(map[string]interface{}{
			"id":           gu.ID,
			"customer_id":  gu.FirstName,
			"last_name":    gu.LastName,
			"email":        gu.Email,
			"phone":        gu.Phone,
			"total_stamps": 0,
			"stamps_count": 0,
		})
		return
	}

	http.Error(w, "Not found", 404)
}

func StampHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	updatedCard, _ := repo.AddStamp(cardID)
	json.NewEncoder(w).Encode(updatedCard)
}

func UseRewardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	if err := repo.UseReward(r.URL.Query().Get("id")); err != nil {
		http.Error(w, "Error", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	cardID := r.URL.Query().Get("id")
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	targetURL := fmt.Sprintf("%s://%s/card?store=%s&id=%s", scheme, r.Host, store.Slug, cardID)
	png, _ := qrcode.Encode(targetURL, qrcode.Medium, 256)
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func PublicStatsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	stats, _ := repo.GetPublicStats(store.ID)
	json.NewEncoder(w).Encode(stats)
}

func GetSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	json.NewEncoder(w).Encode(store)
}

func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)

	var input models.Store
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	store.Name = input.Name
	store.LogoURL = input.LogoURL
	store.PrimaryColor = input.PrimaryColor
	store.StampIcon = input.StampIcon
	store.ThemeMode = input.ThemeMode
	store.Bronze = input.Bronze
	store.Silver = input.Silver
	store.Gold = input.Gold
	store.TextColor = input.TextColor
	store.BorderColor = input.BorderColor
	store.CardImageUrl = input.CardImageUrl
	store.CardImageZoom = input.CardImageZoom
	store.CardImagePosX = input.CardImagePosX
	store.CardImagePosY = input.CardImagePosY
	store.CardScope = input.CardScope

	store.SocialInstagram = input.SocialInstagram
	store.SocialFacebook = input.SocialFacebook
	store.SocialTwitter = input.SocialTwitter
	store.SocialWhatsapp = input.SocialWhatsapp
	store.SocialTiktok = input.SocialTiktok
	store.SocialYoutube = input.SocialYoutube
	store.SocialWebsite = input.SocialWebsite
	store.MenuUrl = input.MenuUrl
	store.LocationUrl = input.LocationUrl

	if err := repo.UpdateSettings(*store); err != nil {
		log.Printf("Erro ao atualizar settings: %v", err)
		http.Error(w, "DB Error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "success"})
}

func UpdateSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Design string `json:"design"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	repo.UpdateSkin(store.ID, req.Design)
	w.WriteHeader(http.StatusOK)
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Old != store.AdminPassword {
		http.Error(w, "Wrong old password", 401)
		return
	}
	repo.UpdatePassword(store.ID, req.New)
	w.WriteHeader(http.StatusOK)
}

func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Password == store.AdminPassword {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Unauthorized", 401)
	}
}

func UpdateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.LoyaltyCard
	json.NewDecoder(r.Body).Decode(&req)
	repo.UpdateCard(req)
	w.WriteHeader(http.StatusOK)
}

func UpdateConsentHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		ID   string `json:"id"`
		Rgpd *bool  `json:"rgpd_accepted"`
		Mkt  *bool  `json:"marketing_accepted"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	repo.UpdateConsent(req.ID, req.Rgpd, req.Mkt)
	w.WriteHeader(http.StatusOK)
}

func AdminResetHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	repo.ResetCard(r.URL.Query().Get("id"))
	w.WriteHeader(http.StatusOK)
}

// Master: Get All Skins
func GetMasterSkinsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	skins, err := repo.GetAllSkins()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(skins)
}

// Master: Save Skin
func SaveSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var s models.Skin
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	if err := repo.SaveSkin(s); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Master: Delete Skin
func DeleteSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	id := r.URL.Query().Get("id")
	if err := repo.DeleteSkin(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// System: Get Available Skins
func GetAvailableSkinsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	skins, err := repo.GetAvailableSkins()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(skins)
}

func MasterToggleStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		ID       string `json:"id"`
		IsActive bool   `json:"isActive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if err := repo.ToggleStoreStatus(req.ID, req.IsActive); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func MasterUpdateStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var store models.Store
	if err := json.NewDecoder(r.Body).Decode(&store); err != nil {
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	if store.ID == "" {
		http.Error(w, "Store ID is required", http.StatusBadRequest)
		return
	}

	if err := repo.UpdateStore(store); err != nil {
		log.Printf("Error updating store: %v", err)
		http.Error(w, "Failed to update store", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Store updated successfully"})
}

func PublicRegisterHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.RegisterStoreRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Nome, Email e Password são obrigatórios", http.StatusBadRequest)
		return
	}

	baseSlug := generateSlug(req.Name)
	req.Slug = baseSlug

	id, err := repo.RegisterStore(req)
	if err != nil {
		req.Slug = fmt.Sprintf("%s-%d", baseSlug, rand.Intn(9999))
		id, err = repo.RegisterStore(req)

		if err != nil {
			log.Printf("Register Error: %v", err)
			http.Error(w, "Erro ao criar conta. O Email pode já estar em uso.", http.StatusConflict)
			return
		}
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    id,
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now().Add(24 * time.Hour),
	})

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Loja criada! A entrar...",
		"redirect": "/admin?store=" + req.Slug,
	})
}

func generateSlug(name string) string {
	slug := strings.ToLower(name)
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	if slug == "" {
		slug = fmt.Sprintf("store-%d", rand.Intn(100000))
	}
	return slug
}

// --- UTILIZADORES GLOBAIS (VOLTO WALLET) ---

func WalletRegisterHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.RegisterGlobalUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato inválido", 400)
		return
	}

	newUser := models.GlobalUser{
		ID:                uuid.New().String(),
		FirstName:         req.FirstName,
		LastName:          req.LastName,
		Email:             req.Email,
		Phone:             req.Phone,
		Password:          req.Password,
		RgpdAccepted:      req.Rgpd,
		MarketingAccepted: req.Marketing,
	}

	if err := repo.CreateGlobalUser(newUser); err != nil {
		http.Error(w, "Email já existe", 409)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: newUser.ID, Path: "/", HttpOnly: true, Expires: time.Now().Add(24 * time.Hour)})

	json.NewEncoder(w).Encode(map[string]string{
		"redirect": "/card?id=" + newUser.ID + "#global_qr",
	})
}

func WalletLoginHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct{ Email, Password string }
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Formato inválido", 400)
		return
	}

	user, err := repo.AuthenticateGlobalUser(req.Email, req.Password)
	if err != nil {
		http.Error(w, "Credenciais inválidas", 401)
		return
	}

	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: user.ID, Path: "/", HttpOnly: true, Expires: time.Now().Add(24 * time.Hour)})

	json.NewEncoder(w).Encode(map[string]string{
		"redirect": "/card?id=" + user.ID,
	})
}

func WalletProfileHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	id := r.URL.Query().Get("id")
	gu, err := repo.GetGlobalUserByID(id)
	if err != nil {
		http.Error(w, "User not found", 404)
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"id":           gu.ID,
		"customer_id":  gu.FirstName,
		"last_name":    gu.LastName,
		"email":        gu.Email,
		"phone":        gu.Phone,
		"total_stamps": 0,
		"stamps_count": 0,
	})
}

func WalletMyCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	email := r.URL.Query().Get("email")
	cards, err := repo.GetMyWalletCards(email)
	if err != nil {
		fmt.Println("❌ ERRO SQL GetMyWalletCards:", err)
		http.Error(w, "Erro BD: "+err.Error(), 500)
		return
	}
	if cards == nil {
		cards = []map[string]interface{}{}
	}
	json.NewEncoder(w).Encode(cards)
}

// --- GESTÃO DE ÂMBITOS (ADMIN) ---

func GetScopesHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// AGORA É 100% SEGURO
	store := getCurrentStore(r)
	if store == nil {
		http.Error(w, "Unauthorized", 401)
		return
	}

	scopes, err := repo.GetStoreScopes(store.ID)
	if err != nil {
		http.Error(w, "Error fetching scopes", 500)
		return
	}
	json.NewEncoder(w).Encode(scopes)
}

func CreateScopeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)

	var req models.StoreScope
	json.NewDecoder(r.Body).Decode(&req)

	req.ID = uuid.New().String()
	req.StoreID = store.ID
	req.IsMain = false
	req.IsActive = true

	err := repo.CreateStoreScope(req)
	if err != nil {
		http.Error(w, "Erro ao criar âmbito (nome duplicado?)", 400)
		return
	}
	json.NewEncoder(w).Encode(req)
}

func ToggleScopeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)

	var req struct {
		ID       string `json:"id"`
		IsActive bool   `json:"is_active"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Payload", 400)
		return
	}

	// Faz o Update usando o ID validado
	err := repo.ToggleScopeStatus(req.ID, store.ID, req.IsActive)
	if err != nil {
		http.Error(w, "Erro ao atualizar", 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func UpdateScopeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)

	var req struct {
		ID        string `json:"id"`
		Name      string `json:"name"`
		StampIcon string `json:"stamp_icon"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	err := repo.UpdateStoreScope(req.ID, store.ID, req.Name, req.StampIcon)
	if err != nil {
		http.Error(w, "Erro ao atualizar âmbito", 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func DeleteScopeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)

	var req struct {
		ID string `json:"id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Payload", 400)
		return
	}

	err := repo.DeleteStoreScope(req.ID, store.ID)
	if err != nil {
		http.Error(w, "Erro ao apagar âmbito", 500)
		return
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// --- NOTIFICAÇÕES (WALLET) ---

func GetNotificationsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	email := r.URL.Query().Get("email")
	if email == "" {
		http.Error(w, "Email required", 400)
		return
	}
	notifs, err := repo.GetUserNotifications(email)
	if err != nil {
		http.Error(w, "Erro BD", 500)
		return
	}
	if notifs == nil {
		notifs = []models.WalletNotification{}
	}
	json.NewEncoder(w).Encode(notifs)
}

func MarkNotificationsReadHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		Email string `json:"email"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	repo.MarkNotificationsAsRead(req.Email)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// --- ADMIN NOTIFICATIONS ---

func GetAdminNotificationsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	notifs, err := repo.GetStoreNotifications(store.ID)
	if err != nil {
		http.Error(w, "Erro BD", 500)
		return
	}
	if notifs == nil {
		notifs = []models.StoreNotification{}
	}
	json.NewEncoder(w).Encode(notifs)
}

func MarkAdminNotificationsReadHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	repo.MarkStoreNotificationsAsRead(store.ID)
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func MasterSendNotificationHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		StoreID string `json:"store_id"` // "all" para todas as lojas
		Title   string `json:"title"`
		Message string `json:"message"`
		Type    string `json:"type"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.StoreID == "all" {
		repo.BroadcastStoreNotification(req.Title, req.Message, req.Type)
	} else {
		repo.SendStoreNotification(req.StoreID, req.Title, req.Message, req.Type)
	}
	json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}
