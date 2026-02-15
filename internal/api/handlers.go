package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

// Helper para obter a loja atual do Contexto (Injetado pelo Middleware)
func getCurrentStore(r *http.Request) *models.Store {
	store, ok := r.Context().Value("current_store").(*models.Store)
	if !ok {
		return nil
	}
	return store
}

// --- AUTENTICAÇÃO E LOGIN ---

// LoginHandler gere a entrada (Admin da Loja ou Cliente da Loja)
func LoginHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	if store == nil {
		http.Error(w, "Loja não identificada", http.StatusInternalServerError)
		return
	}

	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", 400)
		return
	}

	// 1. ADMIN LOGIN (Verifica a password desta loja específica)
	if req.Identifier == "admin" || req.Identifier == "loja" {
		if req.Password == store.AdminPassword {
			// CRIA O COOKIE
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "authenticated_admin",
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(24 * time.Hour),
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"role":     "admin",
				"redirect": fmt.Sprintf("/admin?store=%s", store.Slug),
			})
			return
		}
	}

	// 2. CUSTOMER LOGIN (Filtra pelo StoreID para garantir que o cliente é desta loja)
	card, err := repo.GetCardByEmailOrPhone(store.ID, req.Identifier)
	if err == nil && card != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"role":     "customer",
			"redirect": fmt.Sprintf("/card?store=%s&id=%s", store.Slug, card.ID),
		})
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

// PublicStatsHandler fornece números para o ecrã de login desta loja
func PublicStatsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	if store == nil {
		http.Error(w, "Store context missing", 500)
		return
	}

	stats, err := repo.GetPublicStats(store.ID)
	if err != nil {
		http.Error(w, "Error fetching stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// VerifyPasswordHandler (Usado para zonas protegidas manuais dentro do Admin)
func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Password == store.AdminPassword {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

// --- GESTÃO DE CARTÕES ---

// CreateCardHandler generates a new loyalty card for a customer
func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	if store == nil {
		http.Error(w, "Store error", 500)
		return
	}

	var req models.CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	newCard := models.BrunchCard{
		ID:                   uuid.New().String(),
		StoreID:              store.ID, // <--- ASSOCIA À LOJA ATUAL
		CustomerID:           req.CustomerID,
		LastName:             req.LastName,
		Email:                req.Email,
		Phone:                req.Phone,
		NIF:                  req.NIF,
		Design:               req.Design,
		StampsCount:          0,
		TotalStamps:          0,
		TotalRedeemedBonuses: 0,
		Is_reward_ready:      false,
		RgpdAccepted:         req.RgpdAccepted,
		MarketingAccepted:    req.MarketingAccepted,
	}

	if err := repo.SaveCard(newCard); err != nil {
		log.Printf("Error saving card: %v", err)
		http.Error(w, "Error saving to database (Email/Phone may exist)", http.StatusInternalServerError)
		return
	}

	// Retorna o cartão criado
	savedCard, err := repo.GetCardByID(newCard.ID)
	if err != nil {
		json.NewEncoder(w).Encode(newCard) // Fallback
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(savedCard)
}

// UpdateCardHandler edits existing card details
func UpdateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.BrunchCard
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}
	if req.ID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}
	if err := repo.UpdateCard(req); err != nil {
		http.Error(w, "Database update error", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdateConsentHandler altera as permissões de RGPD/Mkt via Admin
func UpdateConsentHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		ID   string `json:"id"`
		Rgpd *bool  `json:"rgpd_accepted"`
		Mkt  *bool  `json:"marketing_accepted"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if err := repo.UpdateConsent(req.ID, req.Rgpd, req.Mkt); err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// ListAllCardsHandler lists all cards for the CURRENT Store
func ListAllCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	if store == nil {
		http.Error(w, "Store missing", 500)
		return
	}

	// Passa o ID da loja para filtrar
	cards, err := repo.GetAllCards(store.ID)
	if err != nil {
		http.Error(w, "Erro ao listar cartões", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// AdminResetHandler handles total reset of customer history
func AdminResetHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if err := repo.ResetCard(cardID); err != nil {
		http.Error(w, "Erro ao resetar cartão", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// SearchHandler handles customer search via AJAX (Scoped by Store)
func SearchHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	query := r.URL.Query().Get("q")

	if len(query) < 2 {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// Passa o StoreID para não encontrar clientes de outras lojas
	cards, err := repo.SearchCards(store.ID, query)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// --- OPERAÇÕES DO CARTÃO ---

// GetStatusHandler returns card info
func GetStatusHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	card, err := repo.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}

// StampHandler processes visit validation
func StampHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	updatedCard, err := repo.AddStamp(cardID)
	if err != nil {
		http.Error(w, "Failed to update card", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCard)
}

// UseRewardHandler handles the redemption
func UseRewardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if err := repo.UseReward(cardID); err != nil {
		http.Error(w, "Failed to use reward: "+err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetQRCodeHandler generates a QR code image
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	cardID := r.URL.Query().Get("id")

	scheme := "http"
	if r.TLS != nil || r.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	// QR Code agora inclui o slug da loja para garantir que abre no sítio certo
	targetURL := fmt.Sprintf("%s://%s/card?store=%s&id=%s", scheme, r.Host, store.Slug, cardID)

	png, err := qrcode.Encode(targetURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR Code", 500)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

// --- SETTINGS & CONFIGURAÇÃO ---

// UpdateSkinHandler updates the CARD SKIN for the STORE
func UpdateSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Design string `json:"design"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", 400)
		return
	}

	// Atualiza o Skin na tabela de Lojas
	if err := repo.UpdateSkin(store.ID, req.Design); err != nil {
		http.Error(w, "Error updating skin", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetSettingsHandler returns current global store configuration
func GetSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	if store == nil {
		http.Error(w, "Store missing", 500)
		return
	}
	// Retorna o objeto Store diretamente, que tem tudo o que o frontend precisa
	json.NewEncoder(w).Encode(store)
}

// UpdateSettingsHandler saves configuration
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)

	// Lemos os dados novos para uma struct temporária
	var input models.Store
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid payload", 400)
		return
	}

	// Atualizamos os campos do objeto store atual com o input
	store.Name = input.Name
	store.LogoURL = input.LogoURL
	store.PrimaryColor = input.PrimaryColor
	store.ThemeMode = input.ThemeMode
	store.Bronze = input.Bronze
	store.Silver = input.Silver
	store.Gold = input.Gold

	// Guardamos na BD
	if err := repo.UpdateSettings(*store); err != nil {
		log.Printf("Error saving settings: %v", err)
		http.Error(w, "Failed to save settings", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdatePasswordHandler changes the admin key for THIS store
func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	if req.Old != store.AdminPassword {
		http.Error(w, "Wrong old password", http.StatusUnauthorized)
		return
	}

	if err := repo.UpdatePassword(store.ID, req.New); err != nil {
		http.Error(w, "DB Error", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// --- MASTER HANDLERS (Não dependem do middleware de loja) ---

func MasterCreateStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.Store
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid", 400)
		return
	}
	if err := repo.CreateStore(req); err != nil {
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
