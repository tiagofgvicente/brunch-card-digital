package api

import (
	"brunch-card-digital/internal/database"
	"encoding/json"
	"net/http"
)

// StampHandler handles adding a stamp to a card
func StampHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Missing card ID", http.StatusBadRequest)
		return
	}

	// Logic: Increment stamp in DB and return updated card
	// We'll create this method in the repository next
	updatedCard, err := repo.AddStamp(cardID)
	if err != nil {
		http.Error(w, "Failed to add stamp: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updatedCard)
}
