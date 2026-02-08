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
		CustomerID:           req.CustomerID, // First Name
		LastName:             req.LastName,
		Email:                req.Email,
		Phone:                req.Phone,
		NIF:                  req.NIF,
		Design:               req.Design,
		StampsCount:          0,
		TotalStamps:          0,
		TotalRedeemedBonuses: 0,
		Is_reward_ready:      false,
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
		log.Printf("Error decoding update request: %v", err)
		http.Error(w, "Invalid data format", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "Customer ID is required", http.StatusBadRequest)
		return
	}

	if err := repo.UpdateCard(req); err != nil {
		log.Printf("Error updating card %s: %v", req.ID, err)
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

	var req struct {
		Design string `json:"design"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.Design == "" {
		http.Error(w, "Design name is required", http.StatusBadRequest)
		return
	}

	err := repo.UpdateGlobalDesign(req.Design)
	if err != nil {
		http.Error(w, "Error updating global skin: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Skin updated successfully"))
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
		log.Printf("Error decoding settings update: %v", err)
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
// SearchHandler searches for cards using a general term (name, email, phone, nif)
func SearchHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Get the 'q' parameter from the URL
	query := r.URL.Query().Get("q")

	// 2. Minimum 2 characters to trigger search to avoid overloading the DB
	if len(query) < 2 {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}

	// 3. Call the repository search function
	cards, err := repo.SearchCards(query)
	if err != nil {
		http.Error(w, "Search failed", http.StatusInternalServerError)
		return
	}

	// 4. Return results as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// GetStatusHandler returns card info for public view
// GetStatusHandler fetches the current state of a loyalty card from the database
func GetStatusHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Get the card ID from the query string
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Card ID is required", http.StatusBadRequest)
		return
	}

	// 2. Fetch the card details using the repository
	card, err := repo.GetCardByID(cardID)
	if err != nil {
		// Logamos o erro real no servidor para sabermos se o SQL falhou
		log.Printf("Error fetching card status for ID %s: %v", cardID, err)

		// Se o erro for do SQL por colunas em falta, o err não será nulo
		http.Error(w, "Card not found or database error", http.StatusNotFound)
		return
	}

	// 3. Return the card object as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(card); err != nil {
		log.Printf("Error encoding card JSON: %v", err)
	}
}

// StampHandler processes visit validation
// StampHandler handles the logic of adding a new stamp to a customer's card
func StampHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Validate Method: Only POST (from UI) or GET (for quick browser testing) are allowed
	if r.Method != http.MethodPost && r.Method != http.MethodGet {
		http.Error(w, "Method not allowed. Use POST or GET.", http.StatusMethodNotAllowed)
		return
	}

	// 2. Extract Card ID from query parameters (?id=UUID)
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Card ID is required", http.StatusBadRequest)
		return
	}

	// 3. Update Database: Increment stamps and check for rewards
	// The AddStamp method in the repository handles the logic atomically
	updatedCard, err := repo.AddStamp(cardID)
	if err != nil {
		// If the ID doesn't exist or database is down
		http.Error(w, "Failed to update card: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 4. Return the updated card as JSON
	// This allows the Vue.js frontend to update the UI immediately
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedCard)
}

// UseRewardHandler handles the redemption of a free reward
// UseRewardHandler handles the deduction of 10 points for a reward
func UseRewardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if err := repo.UseReward(cardID); err != nil {
		http.Error(w, "Failed to use reward", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// GetQRCodeHandler generates a QR code image (Placeholder or logic)
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Extrair o ID dos parâmetros da query (?id=UUID)
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// 2. Validar se o cartão existe na base de dados
	_, err := repo.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card ID not found", http.StatusNotFound)
		return
	}

	// 3. Construir o URL que o cliente vai usar
	// detetamos o protocolo (http/https) e o host (localhost:8080 ou domínio real)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	// Usamos o r.Host para que o QR Code funcione tanto em localhost como no Kubernetes
	targetURL := fmt.Sprintf("%s://%s/card?id=%s", scheme, r.Host, cardID)

	// 4. Gerar os bytes do QR Code com o URL completo
	// Nível de recuperação Médio e tamanho de 256px
	png, err := qrcode.Encode(targetURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR Code", http.StatusInternalServerError)
		return
	}

	// 5. Enviar a imagem com os headers corretos
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(png)
}
