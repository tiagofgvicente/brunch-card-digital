package api

import (
	"brunch-card-digital/internal/database"
	"encoding/json"
	"net/http"
)

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
		// If the card doesn't exist, return a 404
		http.Error(w, "Card not found", http.StatusNotFound)
		return
	}

	// 3. Return the card object as JSON
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(card)
}
