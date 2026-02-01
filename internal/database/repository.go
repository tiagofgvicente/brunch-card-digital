package database

import (
	"database/sql"
	"fmt"

	"brunch-card-digital/internal/models"
)

// CardRepository handles database operations for Brunch Cards
type CardRepository struct {
	db *sql.DB
}

// NewCardRepository creates a new repository instance
func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// SaveCard inserts a new card into the PostgreSQL database
func (r *CardRepository) SaveCard(card models.BrunchCard) error {
	query := `
		INSERT INTO brunch_cards (id, customer_id, stamps_count, is_reward_ready, design, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, NOW(), NOW())
	`
	_, err := r.db.Exec(query, card.ID, card.CustomerID, card.StampsCount, card.IsRewardReady, card.Design)
	if err != nil {
		return fmt.Errorf("failed to insert card: %w", err)
	}
	return nil
}

// / AddStamp handles the atomic increment of both current and total stamps
func (r *CardRepository) AddStamp(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard
	// Logic: If current is 10, next one resets it to 1. Total always grows.
	query := `
		UPDATE brunch_cards 
		SET stamps_count = CASE WHEN stamps_count >= 10 THEN 1 ELSE stamps_count + 1 END,
		    total_stamps = total_stamps + 1,
		    is_reward_ready = (CASE WHEN stamps_count >= 9 OR (stamps_count = 10) THEN true ELSE false END),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, customer_id, stamps_count, total_stamps, is_reward_ready;
	`
	err := r.db.QueryRow(query, id).Scan(&card.ID, &card.CustomerID, &card.StampsCount, &card.TotalStamps, &card.IsRewardReady)
	return &card, err
}

// UseReward subtracts 10 stamps from the total (consumes one side-reward)
func (r *CardRepository) UseReward(id string) error {
	query := `UPDATE brunch_cards SET total_stamps = total_stamps - 10 WHERE id = $1 AND total_stamps >= 10`
	_, err := r.db.Exec(query, id)
	return err
}

// GetCardByID retrieves a specific card by its UUID
func (r *CardRepository) GetCardByID(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard

	query := `
		SELECT id, customer_id, stamps_count, is_reward_ready, design 
		FROM brunch_cards 
		WHERE id = $1
	`

	err := r.db.QueryRow(query, id).Scan(
		&card.ID,
		&card.CustomerID,
		&card.StampsCount,
		&card.IsRewardReady,
		&card.Design,
	)

	if err != nil {
		return nil, err
	}

	return &card, nil
}

// GetAllCards retorna todos os clientes para o Dashboard
func (r *CardRepository) GetAllCards() ([]models.BrunchCard, error) {
	query := `SELECT id, customer_id, stamps_count, total_stamps, is_reward_ready FROM brunch_cards ORDER BY updated_at DESC`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		if err := rows.Scan(&c.ID, &c.CustomerID, &c.StampsCount, &c.TotalStamps, &c.IsRewardReady); err != nil {
			return nil, err
		}
		cards = append(cards, c)
	}
	return cards, nil
}

// ResetCard zera os contadores de um cliente específico (Ação de Admin)
func (r *CardRepository) ResetCard(id string) error {
	query := `UPDATE brunch_cards SET stamps_count = 0, total_stamps = 0, is_reward_ready = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
