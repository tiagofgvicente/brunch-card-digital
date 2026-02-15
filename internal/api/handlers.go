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

// --- AUTENTICAÇÃO E LOGIN (NOVO) ---

// LoginHandler gere a entrada única (Admin ou Cliente)
func LoginHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		Identifier string `json:"identifier"`
		Password   string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	// ADMIN LOGIN
	if req.Identifier == "admin" || req.Identifier == "loja" {
		if repo.VerifyPassword(req.Password) {
			// CRIA O COOKIE PARA EVITAR DUPLO LOGIN
			http.SetCookie(w, &http.Cookie{
				Name:     "session_token",
				Value:    "authenticated_admin",
				Path:     "/",
				HttpOnly: true,
				Expires:  time.Now().Add(24 * time.Hour),
			})
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{"role": "admin", "redirect": "/admin"})
			return
		}
	}

	// CUSTOMER LOGIN
	card, err := repo.GetCardByEmailOrPhone(req.Identifier)
	if err == nil && card != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"role": "customer", "redirect": fmt.Sprintf("/card?id=%s", card.ID)})
		return
	}

	http.Error(w, "Invalid credentials", http.StatusUnauthorized)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// Apaga o cookie
	http.SetCookie(w, &http.Cookie{
		Name:     "session_token",
		Value:    "",
		Path:     "/",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
	})
	w.WriteHeader(http.StatusOK)
}

// PublicStatsHandler fornece números para o ecrã de login
func PublicStatsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	stats, err := repo.GetPublicStats()
	if err != nil {
		http.Error(w, "Error fetching stats", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stats)
}

// VerifyPasswordHandler verifica a password (usado para zonas protegidas manuais)
func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		Password string `json:"password"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if repo.VerifyPassword(req.Password) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}
}

// --- GESTÃO DE CARTÕES ---

// CreateCardHandler generates a new loyalty card for a customer
func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.CreateCardRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding creation request: %v", err)
		http.Error(w, "Invalid input data", http.StatusBadRequest)
		return
	}

	newCard := models.BrunchCard{
		ID:                   uuid.New().String(),
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
		// NOVOS CAMPOS RGPD
		RgpdAccepted:      req.RgpdAccepted,
		MarketingAccepted: req.MarketingAccepted,
	}

	if err := repo.SaveCard(newCard); err != nil {
		log.Printf("Error saving card: %v", err)
		http.Error(w, "Error saving to database", http.StatusInternalServerError)
		return
	}

	savedCard, err := repo.GetCardByID(newCard.ID)
	if err != nil {
		// Fallback se o Get falhar, devolve o objeto criado
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newCard)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(savedCard)
}

// UpdateCardHandler edits existing card details (Admin Dashboard)
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
	fmt.Fprint(w, "Customer updated successfully")
}

// UpdateConsentHandler altera as permissões de RGPD/Mkt via Admin
func UpdateConsentHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		ID   string `json:"id"`
		Rgpd *bool  `json:"rgpd_accepted"`
		Mkt  *bool  `json:"marketing_accepted"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if err := repo.UpdateConsent(req.ID, req.Rgpd, req.Mkt); err != nil {
		http.Error(w, "DB Error", http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}
}

// ListAllCardsHandler lists all cards for the Admin Dashboard
func ListAllCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cards, err := repo.GetAllCards()
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

// SearchHandler handles customer search via AJAX
func SearchHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	cards, err := repo.SearchCards(query)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// --- OPERAÇÕES DO CARTÃO (Stamp, Reward, Status) ---

// GetStatusHandler returns card info for public view
func GetStatusHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Card ID is required", http.StatusBadRequest)
		return
	}
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
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Card ID is required", http.StatusBadRequest)
		return
	}
	updatedCard, err := repo.AddStamp(cardID)
	if err != nil {
		http.Error(w, "Failed to update card", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCard)
}

// UseRewardHandler handles the redemption of a free reward
func UseRewardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if err := repo.UseReward(cardID); err != nil {
		http.Error(w, "Failed to use reward", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetQRCodeHandler generates a QR code image
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	targetURL := fmt.Sprintf("%s://%s/card?id=%s", scheme, r.Host, cardID)
	png, err := qrcode.Encode(targetURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR Code", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(png)
}

// --- SETTINGS & CONFIGURAÇÃO ---

// UpdateSkinHandler updates the design theme globally
func UpdateSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var req struct {
		Design string `json:"design"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	if err := repo.UpdateGlobalDesign(req.Design); err != nil {
		http.Error(w, "Error updating global skin", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetSettingsHandler returns current global store configuration
func GetSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cfg, err := repo.GetSettings()
	if err != nil {
		log.Printf("Error loading settings: %v", err)
		http.Error(w, "Failed to load system settings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cfg)
}

// UpdateSettingsHandler saves new global system configuration
func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var cfg models.StoreConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		http.Error(w, "Invalid payload", http.StatusBadRequest)
		return
	}
	if err := repo.UpdateSettings(cfg); err != nil {
		log.Printf("Error saving settings: %v", err)
		http.Error(w, "Failed to save system settings", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// UpdatePasswordHandler changes the admin key
func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid body", http.StatusBadRequest)
		return
	}
	if repo.UpdatePassword(req.Old, req.New) {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Wrong password", http.StatusUnauthorized)
	}
}
