package models

import "time"

// CardDesign defines the visual style of the card
type CardDesign string

const (
	Minimalist CardDesign = "minimalist"
	Tropical   CardDesign = "tropical"
	Retro      CardDesign = "retro"
	Modern     CardDesign = "modern"
	Classic    CardDesign = "classic"
	Holiday    CardDesign = "holiday"
	Valentine  CardDesign = "valentine"
)

// BrunchCard represents the customer's digital loyalty card
type BrunchCard struct {
	ID                   string    `json:"id"`
	MemberNumber         int       `json:"member_number"`
	CustomerID           string    `json:"customer_id"`
	LastName             string    `json:"last_name"`
	Email                string    `json:"email"`
	Phone                string    `json:"phone"`
	NIF                  string    `json:"nif"`
	StampsCount          int       `json:"stamps_count"`
	TotalStamps          int       `json:"total_stamps"`
	TotalRedeemedBonuses int       `json:"total_redeemed_bonuses"`
	Is_reward_ready      bool      `json:"is_reward_ready"`
	Design               string    `json:"design"`
	RgpdAccepted         bool      `json:"rgpd_accepted"`
	MarketingAccepted    bool      `json:"marketing_accepted"`
	ConsentDate          time.Time `json:"consent_date"`
	UpdatedAt            time.Time `json:"updated_at"`
}

type CreateCardRequest struct {
	CustomerID        string `json:"customer_id"` // First Name
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	NIF               string `json:"nif"`
	Design            string `json:"design"`
	RgpdAccepted      bool   `json:"rgpd_accepted"`
	MarketingAccepted bool   `json:"marketing_accepted"`
}

// StoreConfig represents the global system settings
type StoreConfig struct {
	Name            string `json:"name"`
	Logo            string `json:"logo"`
	ThemeMode       string `json:"themeMode"`
	PrimaryColor    string `json:"primaryColor"`
	BronzeThreshold int    `json:"bronzeThreshold"`
	SilverThreshold int    `json:"silverThreshold"`
	GoldThreshold   int    `json:"goldThreshold"`
	AdminPassword   string `json:"-"`
}

type Store struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Slug          string    `json:"slug"`
	LogoURL       string    `json:"logo_url"`
	PrimaryColor  string    `json:"primary_color"`
	StampIcon     string    `json:"stamp_icon"`
	ThemeMode     string    `json:"theme_mode"`
	Bronze        int       `json:"bronze_threshold"`
	Silver        int       `json:"silver_threshold"`
	Gold          int       `json:"gold_threshold"`
	AdminPassword string    `json:"admin_password"`
	CreatedAt     time.Time `json:"created_at"`
}
