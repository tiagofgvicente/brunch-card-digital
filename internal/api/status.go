package api

import (
	"brunch-card-digital/internal/database" // Adjust to your module name
	"encoding/json"
	"net/http"
)

// GetStatusHandler returns the current state of a card
func GetStatusHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Missing ID", http.StatusBadRequest)
		return
	}

	// Logic to fetch card from DB
	// We need to add 'GetCardByID' to our repository
	card, err := repo.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}
