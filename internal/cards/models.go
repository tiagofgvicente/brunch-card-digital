package cards

import "time"

// CardDesign represents the visual theme selected by the user
type CardDesign string

const (
	Minimalist CardDesign = "minimalist"
	Tropical   CardDesign = "tropical"
	Retro      CardDesign = "retro"
)

// BrunchCard represents the digital loyalty card for a customer
type BrunchCard struct {
	ID            string     `json:"id"`              // Unique identifier (UUID)
	CustomerID    string     `json:"customer_id"`     // Linked to the user
	StampsCount   int        `json:"stamps_count"`    // Number of stamps (0 to 9)
	IsRewardReady bool       `json:"is_reward_ready"` // True if 10th brunch is free
	Design        CardDesign `json:"design"`          // Chosen visual template
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}
