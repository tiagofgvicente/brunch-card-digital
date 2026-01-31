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

func (r *CardRepository) AddStamp(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard

	// Atomic update: increment stamps and check if reward is ready
	query := `
		UPDATE brunch_cards 
		SET stamps_count = stamps_count + 1,
		    is_reward_ready = (stamps_count + 1 >= 10),
		    updated_at = NOW()
		WHERE id = $1
		RETURNING id, customer_id, stamps_count, is_reward_ready, design;
	`

	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.CustomerID, &card.StampsCount, &card.IsRewardReady, &card.Design,
	)

	if err != nil {
		return nil, err
	}
	return &card, nil
}

func (r *CardRepository) GetCardByID(id string) (*models.BrunchCard, error) {
	var card models.BrunchCard
	query := `SELECT id, customer_id, stamps_count, is_reward_ready, design FROM brunch_cards WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&card.ID, &card.CustomerID, &card.StampsCount, &card.IsRewardReady, &card.Design,
	)
	if err != nil {
		return nil, err
	}
	return &card, nil
}
