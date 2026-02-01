package database

import (
	"database/sql"
	"fmt"

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

	_, err := r.db.Exec(query,
		card.ID,            // $1
		card.CustomerID,    // $2
		card.LastName,      // $3
		toNull(card.Email), // $4
		toNull(card.Phone), // $5
		toNull(card.NIF),   // $6
		card.StampsCount,   // $7
		card.TotalStamps,   // $8
		card.IsRewardReady, // $9
		card.Design,        // $10
	)
	if err != nil {
		return fmt.Errorf("failed to insert card: %w", err)
	}
	return nil
}

// AddStamp CORRIGIDO: Retorna o objeto completo para o Vue não "limpar" os nomes
func (r *CardRepository) AddStamp(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard
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
		&card.ID, &card.CustomerID, &card.LastName, &card.Email, &card.Phone,
		&card.NIF, &card.StampsCount, &card.TotalStamps, &card.IsRewardReady, &card.Design,
	)
	return &card, err
}

func (r *CardRepository) UseReward(id string) error {
	query := `UPDATE brunch_cards SET total_stamps = total_stamps - 10 WHERE id = $1 AND total_stamps >= 10`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CardRepository) GetCardByID(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard
	query := `
        SELECT id, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, is_reward_ready, design 
        FROM brunch_cards 
        WHERE id = $1
    `
	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.CustomerID, &card.LastName, &card.Email, &card.Phone,
		&card.NIF, &card.StampsCount, &card.TotalStamps, &card.IsRewardReady, &card.Design,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) GetAllCards() ([]models.BrunchCard, error) {
	query := `SELECT id, customer_id, last_name, stamps_count, total_stamps, is_reward_ready FROM brunch_cards ORDER BY updated_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		if err := rows.Scan(&c.ID, &c.CustomerID, &c.LastName, &c.StampsCount, &c.TotalStamps, &c.IsRewardReady); err != nil {
			return nil, err
		}
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
		err := rows.Scan(
			&c.ID, &c.CustomerID, &c.LastName, &c.Email,
			&c.Phone, &c.NIF, &c.StampsCount, &c.TotalStamps, &c.IsRewardReady,
		)
		if err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}
