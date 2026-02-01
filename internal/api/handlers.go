package api

import (
	"encoding/json"
	"net/http"
	"time"

	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
)

// CreateCardRequest defines the payload for creating a new card
type CreateCardRequest struct {
	CustomerID string            `json:"customer_id"`
	Design     models.CardDesign `json:"design"`
}

// CreateCardHandler manages the card creation and persistence
// It now receives the repository as an argument
func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req CreateCardRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	// Initialize the new card model
	newCard := models.BrunchCard{
		ID:            uuid.New().String(),
		CustomerID:    req.CustomerID,
		StampsCount:   0,
		TotalStamps:   0, // Ensure this field exists in your struct
		IsRewardReady: false,
		// Conversion: string(req.Design) ensures the types match
		Design: string(req.Design),
		// Use UpdatedAt if CreatedAt is not in your struct
		UpdatedAt: time.Now(),
	}

	// Persist the card to PostgreSQL
	if err := repo.SaveCard(newCard); err != nil {
		http.Error(w, "Failed to save card: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Respond with the created card details
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(newCard)
}
