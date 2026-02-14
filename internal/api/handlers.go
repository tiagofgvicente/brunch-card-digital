package api

import (
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "brunch-card-digital/internal/database"
    "brunch-card-digital/internal/models"

    "github.com/google/uuid"
    "github.com/skip2/go-qrcode"
)

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
        // NOVOS CAMPOS
        RgpdAccepted:         req.RgpdAccepted,
        MarketingAccepted:    req.MarketingAccepted,
    }

    if err := repo.SaveCard(newCard); err != nil {
        log.Printf("Error saving card: %v", err)
        http.Error(w, "Error saving to database", http.StatusInternalServerError)
        return
    }

    savedCard, err := repo.GetCardByID(newCard.ID)
    if err != nil {
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

// UpdateSkinHandler updates the design theme globally
func UpdateSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
    if r.Method != http.MethodPost {
        http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
        return
    }
    var req struct { Design string `json:"design"` }
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
        http.Error(w, "Failed to save system settings", http.StatusInternalServerError)
        return
    }
    w.WriteHeader(http.StatusOK)
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
    if r.TLS != nil { scheme = "https" }
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

// --- NOVOS HANDLERS DE SEGURANÇA E RGPD ---

// UpdateConsentHandler toggles permissions from Admin
func UpdateConsentHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
    var req struct {
        ID string `json:"id"`
        Rgpd *bool `json:"rgpd_accepted"`
        Mkt  *bool `json:"marketing_accepted"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid body", 400); return
    }
    if err := repo.UpdateConsent(req.ID, req.Rgpd, req.Mkt); err != nil {
        http.Error(w, "DB Error", 500)
    } else {
        w.WriteHeader(http.StatusOK)
    }
}

// VerifyPasswordHandler checks password for Kiosk/Admin entry
func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
    var req struct { Password string `json:"password"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid body", 400); return
    }
    if repo.VerifyPassword(req.Password) {
        w.WriteHeader(http.StatusOK)
    } else {
        http.Error(w, "Unauthorized", 401)
    }
}

// UpdatePasswordHandler changes the admin key
func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
    var req struct { Old string `json:"old"`; New string `json:"new"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid body", 400); return
    }
    if repo.UpdatePassword(req.Old, req.New) {
        w.WriteHeader(http.StatusOK)
    } else {
        http.Error(w, "Wrong password", 401)
    }
}