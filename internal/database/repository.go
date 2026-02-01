package database

import (
	"database/sql"
	"fmt"
	"log"

	"brunch-card-digital/internal/models"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// SaveCard CORRIGIDO: 11 colunas e 11 placeholders ($1 a $11)
func (r *CardRepository) SaveCard(card models.BrunchCard) error {
	query := `
        INSERT INTO brunch_cards (
            id, customer_id, last_name, email, phone, 
            nif, stamps_count, total_stamps, is_reward_ready, design, 
            updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
    `

	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}

	// VERIFICA ESTA ORDEM:
	_, err := r.db.Exec(query,
		card.ID,            // $1
		card.CustomerID,    // $2 (Primeiro Nome)
		card.LastName,      // $3 (Apelido)
		toNull(card.Email), // $4
		toNull(card.Phone), // $5
		toNull(card.NIF),   // $6
		card.StampsCount,   // $7
		card.TotalStamps,   // $8
		card.IsRewardReady, // $9
		card.Design,        // $10
	)
	return err
}

// AddStamp CORRIGIDO: Retorna o objeto completo para o Vue não "limpar" os nomes
func (r *CardRepository) AddStamp(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard
	var email, phone, nif, design sql.NullString

	query := `
        UPDATE brunch_cards 
        SET stamps_count = CASE WHEN stamps_count >= 10 THEN 1 ELSE stamps_count + 1 END,
            total_stamps = total_stamps + 1,
            is_reward_ready = (CASE WHEN stamps_count >= 9 OR (stamps_count = 10) THEN true ELSE false END),
            updated_at = NOW()
        WHERE id = $1
        RETURNING id, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, is_reward_ready, design;
    `
	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.CustomerID, &card.LastName, &email, &phone, &nif,
		&card.StampsCount, &card.TotalStamps, &card.IsRewardReady, &design,
	)

	if err == nil {
		card.Email = email.String
		card.Phone = phone.String
		card.NIF = nif.String
		card.Design = design.String
	}

	return &card, err
}

func (r *CardRepository) UseReward(id string) error {
	query := `UPDATE brunch_cards SET total_stamps = total_stamps - 10 WHERE id = $1 AND total_stamps >= 10`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CardRepository) GetCardByID(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard

	// Criamos variáveis temporárias que suportam NULL
	var email, phone, nif, design sql.NullString

	query := `
        SELECT id, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, is_reward_ready, design 
        FROM brunch_cards 
        WHERE id = $1
    `

	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.CustomerID,
		&card.LastName,
		&email, // Scan para NullString
		&phone, // Scan para NullString
		&nif,   // Scan para NullString
		&card.StampsCount,
		&card.TotalStamps,
		&card.IsRewardReady,
		&design, // Scan para NullString
	)

	if err != nil {
		return nil, err
	}

	// Convertemos de volta para o modelo (o campo .String extrai o valor ou fica "")
	card.Email = email.String
	card.Phone = phone.String
	card.NIF = nif.String
	card.Design = design.String

	return &card, nil
}

// No ficheiro internal/database/repository.go
func (r *CardRepository) GetAllCards() ([]models.BrunchCard, error) {
	// 1. Pedir TODAS as colunas
	query := `SELECT id, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, is_reward_ready FROM brunch_cards ORDER BY updated_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		var email, phone, nif sql.NullString // IMPORTANTE para campos que podem ser NULL

		// 2. O SCAN tem de ter 9 campos para bater com o SELECT
		err := rows.Scan(
			&c.ID, &c.CustomerID, &c.LastName,
			&email, &phone, &nif,
			&c.StampsCount, &c.TotalStamps, &c.IsRewardReady,
		)
		if err != nil {
			log.Printf("Erro no Scan: %v", err)
			return nil, err
		}

		c.Email = email.String
		c.Phone = phone.String
		c.NIF = nif.String

		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) ResetCard(id string) error {
	query := `UPDATE brunch_cards SET stamps_count = 0, total_stamps = 0, is_reward_ready = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CardRepository) SearchCards(term string) ([]models.BrunchCard, error) {
	// 1. Definimos a query com todas as colunas necessárias
	query := `
        SELECT id, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, is_reward_ready 
        FROM brunch_cards 
        WHERE customer_id ILIKE $1 
           OR last_name ILIKE $1 
           OR email ILIKE $1 
           OR phone ILIKE $1 
           OR nif ILIKE $1
        ORDER BY updated_at DESC 
        LIMIT 15`

	searchTerm := "%" + term + "%"
	rows, err := r.db.Query(query, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		// Usamos NullString para campos que podem estar vazios na BD
		var email, phone, nif sql.NullString

		// O Scan tem de bater exatamente com a ordem do SELECT acima (9 campos)
		err := rows.Scan(
			&c.ID,
			&c.CustomerID,
			&c.LastName,
			&email,
			&phone,
			&nif,
			&c.StampsCount,
			&c.TotalStamps,
			&c.IsRewardReady,
		)
		if err != nil {
			return nil, err
		}

		// Mapeamos os valores de volta para a struct
		c.Email = email.String
		c.Phone = phone.String
		c.NIF = nif.String

		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) UpdateCard(card models.BrunchCard) error {
	query := `
        UPDATE brunch_cards 
        SET customer_id = $1, last_name = $2, email = $3, phone = $4, nif = $5, updated_at = NOW()
        WHERE id = $6`

	// Se o frontend enviar string vazia, gravamos NULL para não quebrar os CHECKs
	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}

	result, err := r.db.Exec(query,
		card.CustomerID,    // $1
		card.LastName,      // $2
		toNull(card.Email), // $3
		toNull(card.Phone), // $4
		toNull(card.NIF),   // $5
		card.ID,            // $6
	)
	if err != nil {
		return err
	}

	// Verifica se alguma linha foi realmente afetada
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("nenhum cartão encontrado com o ID: %s", card.ID)
	}
	return nil
}
