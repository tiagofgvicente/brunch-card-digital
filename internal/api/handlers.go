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

// CreateCardHandler gere a criação de novos cartões de sócio
func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// Usamos o modelo definido em models para manter a consistência
	var req models.CreateCardRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Erro ao descodificar criação: %v", err)
		http.Error(w, "Dados de entrada inválidos", http.StatusBadRequest)
		return
	}

	// Criamos o modelo final.
	// Nota: MemberNumber é gerado automaticamente pelo Postgres (SERIAL)
	newCard := models.BrunchCard{
		ID:                   uuid.New().String(),
		CustomerID:           req.CustomerID, // Primeiro Nome
		LastName:             req.LastName,   // Apelido
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

	// Após salvar, o Postgres gera o MemberNumber.
	// Para devolvermos o número real, fazemos um fetch rápido.
	savedCard, err := repo.GetCardByID(newCard.ID)
	if err != nil {
		// Se falhar o fetch, devolvemos o que temos mas o ideal é o salvo
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newCard)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(savedCard)
}

// UpdateCardHandler permite editar os dados de contacto do sócio
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

// ListCardsHandler para o Dashboard Admin
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
