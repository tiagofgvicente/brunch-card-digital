package api

import (
	"brunch-card-digital/internal/database"
	"encoding/json"
	"net/http"
)

// ListAllCardsHandler devolve a lista completa de clientes
func ListAllCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cards, err := repo.GetAllCards()
	if err != nil {
		http.Error(w, "Erro ao listar cartões", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

// AdminResetHandler força o reset de um cartão
func AdminResetHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if err := repo.ResetCard(cardID); err != nil {
		http.Error(w, "Erro ao resetar cartão", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
