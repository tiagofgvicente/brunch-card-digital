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
	CustomerID    string    `json:"customer_id"` // Usaremos como "First Name"
	LastName      string    `json:"last_name"`
	Email         string    `json:"email"`
	Phone         string    `json:"phone"`
	NIF           string    `json:"nif"`
	StampsCount   int       `json:"stamps_count"`
	TotalStamps   int       `json:"total_stamps"`
	IsRewardReady bool      `json:"is_reward_ready"`
	Design        string    `json:"design"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type CreateCardRequest struct {
	CustomerID string `json:"customer_id"` // First Name
	LastName   string `json:"last_name"`
	Email      string `json:"email"`
	Phone      string `json:"phone"`
	NIF        string `json:"nif"`
	Design     string `json:"design"`
}
