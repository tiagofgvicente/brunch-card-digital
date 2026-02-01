package api

import (
	"brunch-card-digital/internal/database"
	"encoding/json"
	"log"
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
