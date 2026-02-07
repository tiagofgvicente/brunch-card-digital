package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
)

// CreateCardHandler generates a new loyalty card for a customer
func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.CreateCardRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao descodificar criação: %v", err)
		http.Error(w, "Dados de entrada inválidos", http.StatusBadRequest)
		return
	}

	newCard := models.BrunchCard{
		ID:                   uuid.New().String(),
		CustomerID:           req.CustomerID, // Fisrt Name
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
		log.Printf("Erro ao salvar cartão: %v", err)
		http.Error(w, "Erro ao salvar na base de dados", http.StatusInternalServerError)
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

// UpdateCardHandler edits existing card details (except stamps) - for Admin Dashboard
func UpdateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.BrunchCard

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao descodificar update: %v", err)
		http.Error(w, "Formato de dados inválido", http.StatusBadRequest)
		return
	}

	if req.ID == "" {
		http.Error(w, "ID do cliente é obrigatório", http.StatusBadRequest)
		return
	}

	if err := repo.UpdateCard(req); err != nil {
		log.Printf("Erro ao atualizar cartão %s: %v", req.ID, err)
		http.Error(w, "Erro ao atualizar base de dados", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Cliente atualizado com sucesso")
}

// ListCardsHandler list all cards - for Admin Dashboard
func ListCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cards, err := repo.GetAllCards()
	if err != nil {
		log.Printf("Erro ao listar cartões: %v", err)
		http.Error(w, "Erro ao obter lista de clientes", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(cards)
}

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
		http.Error(w, "Design is required", http.StatusBadRequest)
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
