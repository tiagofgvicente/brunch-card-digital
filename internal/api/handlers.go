package api

import (
	"encoding/json"
	"log"
	"net/http"

	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
)

// CreateCardRequest defines the payload for creating a new card
type CreateCardRequest struct {
	CustomerID string `json:"customer_id"`
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	NIF        string `json:"nif"`
	Design     string `json:"design"`
}

// CreateCardHandler manages the card creation and persistence
// It now receives the repository as an argument
func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.CreateCardRequest // Usa a struct de Request
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Criamos o modelo final para a BD transferindo os dados do Request
	newCard := models.BrunchCard{
		ID:            uuid.New().String(),
		CustomerID:    req.CustomerID, // Primeiro Nome
		LastName:      req.LastName,   // Apelido
		Email:         req.Email,
		Phone:         req.Phone,
		NIF:           req.NIF,
		Design:        req.Design,
		StampsCount:   0,
		TotalStamps:   0,
		IsRewardReady: false,
	}

	if err := repo.SaveCard(newCard); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(newCard)
}

func UpdateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.BrunchCard
	json.NewDecoder(r.Body).Decode(&req)
	log.Printf("Recebido para Update: ID=%s, Nome=%s, Apelido=%s", req.ID, req.CustomerID, req.LastName)
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid input", http.StatusBadRequest)
		return
	}

	if err := repo.UpdateCard(req); err != nil {
		http.Error(w, "Failed to update: "+err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
