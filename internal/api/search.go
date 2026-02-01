package api

import (
	"brunch-card-digital/internal/database"
	"encoding/json"
	"net/http"
)

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
