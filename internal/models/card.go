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
	ID            string     `json:"id"`
	CustomerID    string     `json:"customer_id"`
	StampsCount   int        `json:"stamps_count"`    // Current stamps (0-9)
	IsRewardReady bool       `json:"is_reward_ready"` // True if 10th is free
	Design        CardDesign `json:"design"`          // Chosen by user
	CreatedAt     time.Time  `json:"created_at"`
}
