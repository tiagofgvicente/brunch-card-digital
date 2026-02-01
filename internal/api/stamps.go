package api

import (
	"encoding/json"
	"net/http"

	"brunch-card-digital/internal/database"
)

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

// UseRewardHandler handles the deduction of 10 points for a reward
func UseRewardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if err := repo.UseReward(cardID); err != nil {
		http.Error(w, "Failed to use reward", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

