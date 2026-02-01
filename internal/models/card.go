package models

import "time"

// CardDesign defines the visual style of the card
type CardDesign string

const (
	Minimalist CardDesign = "minimalist"
	Tropical   CardDesign = "tropical"
	Retro      CardDesign = "retro"
)

// BrunchCard represents the customer's digital loyalty card
type BrunchCard struct {
	ID            string    `json:"id"`
	CustomerID    string    `json:"customer_id"`
	StampsCount   int       `json:"stamps_count"`
	TotalStamps   int       `json:"total_stamps"` // ADICIONA ESTA LINHA
	IsRewardReady bool      `json:"is_reward_ready"`
	Design        string    `json:"design"`
	UpdatedAt     time.Time `json:"updated_at"`
}
