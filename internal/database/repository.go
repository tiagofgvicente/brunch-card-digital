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

// GetCardByID retrieves a specific card by its UUID
func (r *CardRepository) GetCardByID(id string) (*models.BrunchCard, error) {
	var c models.BrunchCard
	var email, phone, nif, design sql.NullString

	query := `SELECT id, member_number, customer_id, last_name, email, phone, nif, 
              stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design 
              FROM brunch_cards WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&c.ID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif,
		&c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design,
	)
	if err != nil {
		return nil, err
	}

	c.Email, c.Phone, c.NIF, c.Design = email.String, phone.String, nif.String, design.String
	return &c, nil
}

// UseReward increments redeemed bonuses and updates availability
func (r *CardRepository) UseReward(id string) error {
	query := `
        UPDATE brunch_cards 
        SET total_redeemed_bonuses = total_redeemed_bonuses + 1,
            is_reward_ready = CASE WHEN (total_stamps / 10) - (total_redeemed_bonuses + 1) > 0 THEN TRUE ELSE FALSE END,
            updated_at = NOW() 
        WHERE id = $1 AND (total_stamps / 10) > total_redeemed_bonuses`

	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}

	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("customer has no available bonuses to redeem")
	}
	return nil
}

// AddStamp adds a single visit stamp and logic for reward eligibility
func (r *CardRepository) AddStamp(id string) (*models.BrunchCard, error) {
	query := `
        UPDATE brunch_cards 
        SET 
            total_stamps = total_stamps + 1,
            stamps_count = CASE WHEN stamps_count >= 10 THEN 1 ELSE stamps_count + 1 END,
            is_reward_ready = CASE WHEN stamps_count = 9 OR (stamps_count = 10 AND is_reward_ready = true) THEN TRUE ELSE FALSE END,
            updated_at = NOW()
        WHERE id = $1`

	_, err := r.db.Exec(query, id)
	if err != nil {
		return nil, err
	}
	return r.GetCardByID(id)
}

// GetAllCards fetches all customers for the Admin Dashboard
func (r *CardRepository) GetAllCards() ([]models.BrunchCard, error) {
	query := `SELECT id, member_number, customer_id, last_name, email, phone, nif, 
              stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design 
              FROM brunch_cards ORDER BY member_number DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		var email, phone, nif, design sql.NullString

		err := rows.Scan(
			&c.ID, &c.MemberNumber, &c.CustomerID, &c.LastName,
			&email, &phone, &nif, &c.StampsCount, &c.TotalStamps,
			&c.TotalRedeemedBonuses, &c.Is_reward_ready, &design,
		)
		if err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}

		c.Email = email.String
		c.Phone = phone.String
		c.NIF = nif.String
		c.Design = design.String

		cards = append(cards, c)
	}
	return cards, nil
}

// SaveCard creates a new card record, inheriting the active shop design if empty
func (r *CardRepository) SaveCard(card models.BrunchCard) error {
	if card.Design == "" || card.Design == "modern" || card.Design == "default" {
		var activeDesign string
		err := r.db.QueryRow("SELECT design FROM brunch_cards WHERE design IS NOT NULL AND design != '' ORDER BY updated_at DESC LIMIT 1").Scan(&activeDesign)
		if err == nil && activeDesign != "" {
			card.Design = activeDesign
		}
	}

	query := `
        INSERT INTO brunch_cards (
            id, customer_id, last_name, email, phone, 
            nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design, 
            updated_at
        )
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 0, $9, $10, NOW())
    `

	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}

	_, err := r.db.Exec(query,
		card.ID, card.CustomerID, card.LastName,
		toNull(card.Email), toNull(card.Phone), toNull(card.NIF),
		card.StampsCount, card.TotalStamps, card.Is_reward_ready, card.Design,
	)
	return err
}

// UpdateCard updates basic profile info of a customer
func (r *CardRepository) UpdateCard(card models.BrunchCard) error {
	query := `
        UPDATE brunch_cards 
        SET customer_id = $1, last_name = $2, email = $3, phone = $4, nif = $5, updated_at = NOW()
        WHERE id = $6`

	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}

	_, err := r.db.Exec(query, card.CustomerID, card.LastName, toNull(card.Email), toNull(card.Phone), toNull(card.NIF), card.ID)
	return err
}

// ResetCard resets loyalty progress but keeps the profile
func (r *CardRepository) ResetCard(id string) error {
	query := `UPDATE brunch_cards SET stamps_count = 0, total_stamps = 0, total_redeemed_bonuses = 0, is_reward_ready = false, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

// SearchCards finds cards by multiple criteria
func (r *CardRepository) SearchCards(term string) ([]models.BrunchCard, error) {
	query := `
        SELECT id, member_number, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready 
        FROM brunch_cards 
        WHERE customer_id ILIKE $1 OR last_name ILIKE $1 OR email ILIKE $1 OR phone ILIKE $1 OR nif ILIKE $1
        ORDER BY member_number DESC LIMIT 15`

	searchTerm := "%" + term + "%"
	rows, err := r.db.Query(query, searchTerm)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		var email, phone, nif sql.NullString
		err := rows.Scan(&c.ID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif, &c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready)
		if err != nil {
			return nil, err
		}
		c.Email, c.Phone, c.NIF = email.String, phone.String, nif.String
		cards = append(cards, c)
	}
	return cards, nil
}

// UpdateGlobalDesign updates the visual theme for all existing cards
func (r *CardRepository) UpdateGlobalDesign(design string) error {
	query := `UPDATE brunch_cards SET design = $1, updated_at = NOW()`
	_, err := r.db.Exec(query, design)
	return err
}

// GetSettings fetches the global store configuration
func (r *CardRepository) GetSettings() (models.StoreConfig, error) {
	var cfg models.StoreConfig
	query := `SELECT store_name, store_logo, theme_mode, primary_color FROM system_settings WHERE id = 1`

	// Certifica-te que a ordem do Scan coincide com o SELECT
	err := r.db.QueryRow(query).Scan(&cfg.Name, &cfg.Logo, &cfg.ThemeMode, &cfg.PrimaryColor)
	if err != nil {
		return cfg, err
	}
	return cfg, nil
}

// UpdateSettings updates the global store configuration
func (r *CardRepository) UpdateSettings(cfg models.StoreConfig) error {
	query := `UPDATE system_settings SET store_name=$1, store_logo=$2, theme_mode=$3, primary_color=$4, updated_at=NOW() WHERE id = 1`
	_, err := r.db.Exec(query, cfg.Name, cfg.Logo, cfg.ThemeMode, cfg.PrimaryColor)
	return err
}
