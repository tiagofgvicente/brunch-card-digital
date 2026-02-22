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
type LoyaltyCard struct {
	ID                   string    `json:"id"`
	StoreID              string    `json:"store_id"`
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

	// --- NOVOS CAMPOS DE ÂMBITO (SCOPE) ---
	ScopeID       string `json:"scope_id"`
	ScopeName     string `json:"scope_name"`
	ScopeIcon     string `json:"scope_icon"`
	ScopeIsActive bool   `json:"scope_is_active"`
}

type CreateCardRequest struct {
	CustomerID        string `json:"customer_id"`
	LastName          string `json:"last_name"`
	Email             string `json:"email"`
	Phone             string `json:"phone"`
	NIF               string `json:"nif"`
	Design            string `json:"design"`
	RgpdAccepted      bool   `json:"rgpd_accepted"`
	MarketingAccepted bool   `json:"marketing_accepted"`
	ScopeID           string `json:"scope_id"`
}

// Store define as configurações da loja
type Store struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`

	// CREDENCIAIS
	AdminUsername string `json:"admin_username"`
	AdminEmail    string `json:"admin_email"`
	AdminPassword string `json:"admin_password"`

	// Campo usado APENAS para receber nova password no Update
	NewPassword string `json:"new_password,omitempty"`

	// BRANDING
	LogoURL      string `json:"logo_url"`
	PrimaryColor string `json:"primary_color"`
	StampIcon    string `json:"stamp_icon"`
	CardSkin     string `json:"card_skin"`
	ThemeMode    string `json:"theme_mode"`
	CardScope    string `json:"card_scope"`

	// --- NOVOS CAMPOS DE DESIGN ---
	TextColor     string `json:"text_color"`
	BorderColor   string `json:"border_color"`
	CardImageUrl  string `json:"card_image_url"`
	CardImageZoom int    `json:"card_image_zoom"`
	CardImagePosX int    `json:"card_image_pos_x"`
	CardImagePosY int    `json:"card_image_pos_y"`

	// SOCIAL MEDIA
	SocialInstagram string `json:"social_instagram"`
	SocialFacebook  string `json:"social_facebook"`
	SocialTwitter   string `json:"social_twitter"`
	SocialWhatsapp  string `json:"social_whatsapp"`
	SocialTiktok    string `json:"social_tiktok"`
	SocialYoutube   string `json:"social_youtube"`
	SocialWebsite   string `json:"social_website"`
	MenuUrl         string `json:"menu_url"`
	LocationUrl     string `json:"location_url"`

	// THRESHOLDS
	Bronze int `json:"bronze_threshold"`
	Silver int `json:"silver_threshold"`
	Gold   int `json:"gold_threshold"`

	CreatedAt time.Time `json:"created_at"`

	// SaaS & Tiers
	IsActive         bool      `json:"isActive" db:"is_active"`
	Status           string    `json:"status"`
	Tier             string    `json:"tier"`
	TierExpiration   time.Time `json:"tier_expiration"`
	BillingCycle     string    `json:"billing_cycle"`
	MaxUsers         int       `json:"max_users"`
	AccountActivated bool      `json:"account_activated"`

	// Métricas
	TotalMembers int `json:"totalMembers" db:"total_members"`
}

type Skin struct {
	ID          string     `json:"id" db:"id"`
	Name        string     `json:"name" db:"name"`
	Type        string     `json:"type" db:"type"`
	ImageData   string     `json:"image" db:"image_data"`
	ColorBg     string     `json:"colorBg" db:"color_bg"`
	ColorText   string     `json:"colorText" db:"color_text"`
	ColorBorder string     `json:"colorBorder" db:"color_border"`
	IsGlobal    bool       `json:"isGlobal" db:"is_global"`
	StoreID     *string    `json:"storeId" db:"store_id"`
	StartDate   *time.Time `json:"start,omitempty" db:"start_date"`
	EndDate     *time.Time `json:"end,omitempty" db:"end_date"`
}

type RegisterStoreRequest struct {
	Name     string `json:"name"`
	Slug     string `json:"slug"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"identifier"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Token    string `json:"token"`
	Redirect string `json:"redirect"`
}

type GlobalUser struct {
	ID                string    `json:"id"`
	FirstName         string    `json:"first_name"`
	LastName          string    `json:"last_name"`
	Email             string    `json:"email"`
	Phone             string    `json:"phone"`
	Password          string    `json:"-"`
	RgpdAccepted      bool      `json:"rgpd_accepted"`
	MarketingAccepted bool      `json:"marketing_accepted"`
	CreatedAt         time.Time `json:"created_at"`
}

type RegisterGlobalUserRequest struct {
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
	Phone     string `json:"phone"`
	Password  string `json:"password"`
	Rgpd      bool   `json:"rgpd"`
	Marketing bool   `json:"marketing"`
}

type StoreScope struct {
	ID        string    `json:"id"`
	StoreID   string    `json:"store_id"`
	Name      string    `json:"name"`
	StampIcon string    `json:"stamp_icon"`
	IsMain    bool      `json:"is_main"`
	IsActive  bool      `json:"is_active"`
	CreatedAt time.Time `json:"created_at"`
}

type WalletNotification struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	StoreID   string    `json:"store_id"`
	StoreName string    `json:"store_name"`
	StoreLogo string    `json:"store_logo"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

type StoreNotification struct {
	ID        string    `json:"id"`
	StoreID   string    `json:"store_id"`
	Title     string    `json:"title"`
	Message   string    `json:"message"`
	Type      string    `json:"type"`
	ImageData string    `json:"image_data"`
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}
