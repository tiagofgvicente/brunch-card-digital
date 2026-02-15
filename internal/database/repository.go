package database

import (
	"database/sql"
	"fmt"

	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

func (r *CardRepository) GetStoreByLogin(identifier string) (*models.Store, error) {
	var s models.Store
	var logoURL sql.NullString

	query := `
		SELECT id, name, slug, admin_username, admin_email, admin_password, logo_url 
		FROM stores 
		WHERE admin_username = $1 OR admin_email = $1
	`

	err := r.db.QueryRow(query, identifier).Scan(
		&s.ID, &s.Name, &s.Slug,
		&s.AdminUsername, &s.AdminEmail, &s.AdminPassword,
		&logoURL,
	)
	if err != nil {
		return nil, err
	}
	s.LogoURL = logoURL.String
	return &s, nil
}

func (r *CardRepository) GetStoreSlugByCustomerEmail(identifier string) (string, string, error) {
	var slug string
	var cardID string

	// Procura por Email OU Telefone nas tabelas de cartões
	query := `
		SELECT s.slug, c.id 
		FROM loyalty_cards c
		JOIN stores s ON c.store_id = s.id
		WHERE c.email = $1 OR c.phone = $1
		LIMIT 1
	`
	err := r.db.QueryRow(query, identifier).Scan(&slug, &cardID)
	return slug, cardID, err
}

// --- CARTÕES (Clientes) ---

func (r *CardRepository) GetCardByID(id string) (*models.BrunchCard, error) {
	var c models.BrunchCard
	var email, phone, nif, design sql.NullString
	query := `SELECT id, store_id, member_number, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design, rgpd_accepted, marketing_accepted FROM loyalty_cards WHERE id = $1`
	err := r.db.QueryRow(query, id).Scan(&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif, &c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design, &c.RgpdAccepted, &c.MarketingAccepted)
	if err != nil {
		return nil, err
	}
	c.Email, c.Phone, c.NIF, c.Design = email.String, phone.String, nif.String, design.String
	return &c, nil
}

func (r *CardRepository) GetCardByEmailOrPhone(storeID, identifier string) (*models.BrunchCard, error) {
	var c models.BrunchCard
	var email, phone, nif, design sql.NullString
	query := `SELECT id, store_id, member_number, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design FROM loyalty_cards WHERE (email = $1 OR phone = $1) AND store_id = $2 LIMIT 1`
	err := r.db.QueryRow(query, identifier, storeID).Scan(&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif, &c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design)
	if err != nil {
		return nil, err
	}
	c.Email, c.Phone, c.NIF, c.Design = email.String, phone.String, nif.String, design.String
	return &c, nil
}

func (r *CardRepository) UseReward(id string) error {
	query := `UPDATE loyalty_cards SET total_redeemed_bonuses = total_redeemed_bonuses + 1, is_reward_ready = CASE WHEN (total_stamps / 10) - (total_redeemed_bonuses + 1) > 0 THEN TRUE ELSE FALSE END, updated_at = NOW() WHERE id = $1 AND (total_stamps / 10) > total_redeemed_bonuses`
	result, err := r.db.Exec(query, id)
	if err != nil {
		return err
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("no bonuses available")
	}
	return nil
}

func (r *CardRepository) AddStamp(id string) (*models.BrunchCard, error) {
	query := `UPDATE loyalty_cards SET total_stamps = total_stamps + 1, stamps_count = CASE WHEN stamps_count >= 10 THEN 1 ELSE stamps_count + 1 END, is_reward_ready = CASE WHEN stamps_count = 9 OR (stamps_count = 10 AND is_reward_ready = true) THEN TRUE ELSE FALSE END, updated_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return nil, err
	}
	return r.GetCardByID(id)
}

func (r *CardRepository) GetAllCards(storeID string) ([]models.BrunchCard, error) {
	query := `SELECT id, store_id, member_number, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design, rgpd_accepted, marketing_accepted FROM loyalty_cards WHERE store_id = $1 ORDER BY member_number DESC`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		var email, phone, nif, design sql.NullString
		err := rows.Scan(&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif, &c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design, &c.RgpdAccepted, &c.MarketingAccepted)
		if err != nil {
			continue
		}
		c.Email, c.Phone, c.NIF, c.Design = email.String, phone.String, nif.String, design.String
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) SaveCard(card models.BrunchCard) error {
	if card.Design == "" {
		card.Design = "default"
	}
	query := `INSERT INTO loyalty_cards (id, store_id, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design, rgpd_accepted, marketing_accepted, consent_date, updated_at) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, $10, $11, $12, $13, NOW(), NOW())`
	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}
	_, err := r.db.Exec(query, card.ID, card.StoreID, card.CustomerID, card.LastName, toNull(card.Email), toNull(card.Phone), toNull(card.NIF), card.StampsCount, card.TotalStamps, card.Is_reward_ready, card.Design, card.RgpdAccepted, card.MarketingAccepted)
	return err
}

func (r *CardRepository) UpdateCard(card models.BrunchCard) error {
	query := `UPDATE loyalty_cards SET customer_id=$1, last_name=$2, email=$3, phone=$4, nif=$5, updated_at=NOW() WHERE id=$6`
	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}
	_, err := r.db.Exec(query, card.CustomerID, card.LastName, toNull(card.Email), toNull(card.Phone), toNull(card.NIF), card.ID)
	return err
}

func (r *CardRepository) ResetCard(id string) error {
	query := `UPDATE loyalty_cards SET stamps_count=0, total_stamps=0, total_redeemed_bonuses=0, is_reward_ready=false, updated_at=NOW() WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CardRepository) SearchCards(storeID, term string) ([]models.BrunchCard, error) {
	query := `SELECT id, store_id, member_number, customer_id, last_name, email, phone, nif, stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, rgpd_accepted, marketing_accepted FROM loyalty_cards WHERE store_id = $1 AND (customer_id ILIKE $2 OR last_name ILIKE $2 OR email ILIKE $2 OR phone ILIKE $2 OR nif ILIKE $2) ORDER BY member_number DESC LIMIT 15`
	rows, err := r.db.Query(query, storeID, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.BrunchCard
	for rows.Next() {
		var c models.BrunchCard
		var email, phone, nif sql.NullString
		rows.Scan(&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif, &c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &c.RgpdAccepted, &c.MarketingAccepted)
		c.Email, c.Phone, c.NIF = email.String, phone.String, nif.String
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) GetPublicStats(storeID string) (map[string]int, error) {
	var totalCards, totalStamps, totalRedeems int
	err := r.db.QueryRow(`SELECT COUNT(*), COALESCE(SUM(total_stamps), 0), COALESCE(SUM(total_redeemed_bonuses), 0) FROM loyalty_cards WHERE store_id = $1`, storeID).Scan(&totalCards, &totalStamps, &totalRedeems)
	if err != nil {
		return nil, err
	}
	return map[string]int{"total_cards": totalCards, "total_stamps": totalStamps, "total_redeems": totalRedeems}, nil
}

func (r *CardRepository) UpdateConsent(id string, rgpd *bool, marketing *bool) error {
	if rgpd != nil {
		r.db.Exec("UPDATE loyalty_cards SET rgpd_accepted=$1 WHERE id=$2", *rgpd, id)
	}
	if marketing != nil {
		r.db.Exec("UPDATE loyalty_cards SET marketing_accepted=$1 WHERE id=$2", *marketing, id)
	}
	return nil
}

// --- LOJAS (Stores) ---

// GetStoreBySlug: A FUNÇÃO QUE ESTAVA A FALTAR!
func (r *CardRepository) GetStoreBySlug(slug string) (*models.Store, error) {
	var s models.Store
	var logoURL sql.NullString

	query := `
		SELECT 
            id, name, slug, admin_username, admin_email, admin_password,
            logo_url, primary_color, stamp_icon, 
            card_skin, theme_mode, 
            bronze_threshold, silver_threshold, gold_threshold
		FROM stores 
		WHERE slug = $1
	`
	err := r.db.QueryRow(query, slug).Scan(
		&s.ID, &s.Name, &s.Slug, &s.AdminUsername, &s.AdminEmail, &s.AdminPassword,
		&logoURL, &s.PrimaryColor, &s.StampIcon,
		&s.CardSkin, &s.ThemeMode, &s.Bronze, &s.Silver, &s.Gold,
	)
	if err != nil {
		return nil, err
	}
	s.LogoURL = logoURL.String
	return &s, nil
}

func (r *CardRepository) CreateStore(s models.Store) error {
	query := `
        INSERT INTO stores (
            id, name, slug, 
            admin_username, admin_email, admin_password, 
            logo_url, primary_color, stamp_icon, card_skin, theme_mode,
            bronze_threshold, silver_threshold, gold_threshold, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, 15, 40, 100, NOW())`

	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	// Defaults
	if s.CardSkin == "" {
		s.CardSkin = "custom"
	}
	if s.ThemeMode == "" {
		s.ThemeMode = "dark"
	}
	if s.PrimaryColor == "" {
		s.PrimaryColor = "#00a896"
	}
	if s.StampIcon == "" {
		s.StampIcon = "🍳"
	}

	var logoURL interface{} = s.LogoURL
	if s.LogoURL == "" {
		logoURL = nil
	}

	_, err := r.db.Exec(query,
		s.ID, s.Name, s.Slug,
		s.AdminUsername, s.AdminEmail, s.AdminPassword,
		logoURL, s.PrimaryColor, s.StampIcon, s.CardSkin, s.ThemeMode)
	return err
}

func (r *CardRepository) GetAllStores() ([]models.Store, error) {
	// Agora retorna também Username e Email para o Master ver
	query := `
		SELECT id, name, slug, admin_username, admin_email, admin_password, created_at 
		FROM stores ORDER BY created_at DESC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var s models.Store
		// Nota: Não estamos a ler logo_url aqui para ser mais rápido na lista
		rows.Scan(&s.ID, &s.Name, &s.Slug, &s.AdminUsername, &s.AdminEmail, &s.AdminPassword, &s.CreatedAt)
		stores = append(stores, s)
	}
	return stores, nil
}

func (r *CardRepository) UpdateSkin(storeID, skin string) error {
	_, err := r.db.Exec("UPDATE stores SET card_skin = $1 WHERE id = $2", skin, storeID)
	return err
}

func (r *CardRepository) UpdateSettings(s models.Store) error {
	// Agora gravamos TUDO: Nome, Logo, Cor, Icon, Tema e Limites
	query := `
        UPDATE stores 
        SET name = $1, 
            logo_url = $2, 
            primary_color = $3, 
            stamp_icon = $4, 
            theme_mode = $5,
            bronze_threshold = $6, 
            silver_threshold = $7, 
            gold_threshold = $8
        WHERE id = $9`

	_, err := r.db.Exec(query,
		s.Name,
		s.LogoURL,
		s.PrimaryColor,
		s.StampIcon,
		s.ThemeMode,
		s.Bronze,
		s.Silver,
		s.Gold,
		s.ID,
	)
	return err
}

func (r *CardRepository) UpdatePassword(storeID, newPass string) error {
	_, err := r.db.Exec("UPDATE stores SET admin_password=$1 WHERE id=$2", newPass, storeID)
	return err
}

func (r *CardRepository) VerifyPassword(pass string) bool { return true }
