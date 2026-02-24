package database

import (
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"fmt"
	"log"
	"strings"
	"time"

	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

type CardRepository struct {
	db *sql.DB
}

func NewCardRepository(db *sql.DB) *CardRepository {
	return &CardRepository{db: db}
}

// --- LOJAS (Stores) ---

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

func (r *CardRepository) GetStoreBySlug(slug string) (*models.Store, error) {
	var s models.Store
	var logoURL, textCol, borderCol, cardImgUrl, cardScope sql.NullString
	var socInsta, socFb, socTwit, socWhats, socTik, socYt, socWeb sql.NullString
	var menuUrl, locUrl sql.NullString
	var imgZoom, imgPosX, imgPosY sql.NullInt32

	query := `
        SELECT 
            id, name, slug, admin_username, admin_email, admin_password,
            logo_url, primary_color, stamp_icon, card_skin, theme_mode, 
            bronze_threshold, silver_threshold, gold_threshold,
            tier, tier_expiration, billing_cycle, max_users, account_activated, status, is_active,
            text_color, border_color, card_image_url, card_image_zoom, card_image_pos_x, card_image_pos_y,
            card_scope, social_instagram, social_facebook, social_twitter, social_whatsapp, social_tiktok, social_youtube, social_website,
            menu_url, location_url
        FROM stores 
        WHERE slug = $1
    `

	err := r.db.QueryRow(query, slug).Scan(
		&s.ID, &s.Name, &s.Slug, &s.AdminUsername, &s.AdminEmail, &s.AdminPassword,
		&logoURL, &s.PrimaryColor, &s.StampIcon, &s.CardSkin, &s.ThemeMode, &s.Bronze, &s.Silver, &s.Gold,
		&s.Tier, &s.TierExpiration, &s.BillingCycle, &s.MaxUsers, &s.AccountActivated, &s.Status, &s.IsActive,
		&textCol, &borderCol, &cardImgUrl, &imgZoom, &imgPosX, &imgPosY,
		&cardScope, &socInsta, &socFb, &socTwit, &socWhats, &socTik, &socYt, &socWeb,
		&menuUrl, &locUrl,
	)
	if err != nil {
		return nil, err
	}

	s.LogoURL = logoURL.String
	if textCol.Valid && textCol.String != "" {
		s.TextColor = textCol.String
	} else {
		s.TextColor = "#ffffff"
	}
	if borderCol.Valid && borderCol.String != "" {
		s.BorderColor = borderCol.String
	} else {
		s.BorderColor = "#ffffff"
	}
	s.CardImageUrl = cardImgUrl.String
	if imgZoom.Valid {
		s.CardImageZoom = int(imgZoom.Int32)
	} else {
		s.CardImageZoom = 100
	}
	if imgPosX.Valid {
		s.CardImagePosX = int(imgPosX.Int32)
	}
	if imgPosY.Valid {
		s.CardImagePosY = int(imgPosY.Int32)
	}
	if cardScope.Valid && cardScope.String != "" {
		s.CardScope = cardScope.String
	} else {
		s.CardScope = "Geral"
	}

	s.SocialInstagram = socInsta.String
	s.SocialFacebook = socFb.String
	s.SocialTwitter = socTwit.String
	s.SocialWhatsapp = socWhats.String
	s.SocialTiktok = socTik.String
	s.SocialYoutube = socYt.String
	s.SocialWebsite = socWeb.String
	s.MenuUrl = menuUrl.String
	s.LocationUrl = locUrl.String

	return &s, nil
}

func (r *CardRepository) CreateStore(s models.Store) error {
	query := `
        INSERT INTO stores (
            id, name, slug, 
            admin_username, admin_email, admin_password, 
            logo_url, primary_color, stamp_icon, card_skin, theme_mode,
            tier, tier_expiration, billing_cycle, max_users,
            bronze_threshold, silver_threshold, gold_threshold, created_at
        ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, 15, 40, 100, CURRENT_TIMESTAMP)`

	if s.ID == "" {
		s.ID = uuid.New().String()
	}
	if s.CardSkin == "" {
		s.CardSkin = "default"
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
	if s.Tier == "" {
		s.Tier = "free_trial"
	}
	if s.BillingCycle == "" {
		s.BillingCycle = "monthly"
	}

	var logoURL interface{} = nil
	if s.LogoURL != "" {
		logoURL = s.LogoURL
	}

	_, err := r.db.Exec(query,
		s.ID, s.Name, s.Slug,
		s.AdminUsername, s.AdminEmail, s.AdminPassword,
		logoURL, s.PrimaryColor, s.StampIcon, s.CardSkin, s.ThemeMode,
		s.Tier, s.TierExpiration, s.BillingCycle, s.MaxUsers)
	return err
}

func (r *CardRepository) GetAllStores() ([]models.Store, error) {
	query := `
        SELECT 
            s.id, s.name, s.slug, s.admin_username, s.admin_email, 
            COALESCE(s.logo_url, ''), s.primary_color, s.stamp_icon, s.card_skin, 
            s.is_active, s.tier, s.tier_expiration, s.billing_cycle, 
            s.max_users, s.account_activated, s.status, s.created_at,
            (SELECT COUNT(*) FROM loyalty_cards WHERE store_id = s.id) as total_members
        FROM stores s 
        ORDER BY s.name ASC`

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stores []models.Store
	for rows.Next() {
		var s models.Store
		if err := rows.Scan(
			&s.ID, &s.Name, &s.Slug, &s.AdminUsername, &s.AdminEmail,
			&s.LogoURL, &s.PrimaryColor, &s.StampIcon, &s.CardSkin,
			&s.IsActive, &s.Tier, &s.TierExpiration, &s.BillingCycle,
			&s.MaxUsers, &s.AccountActivated, &s.Status, &s.CreatedAt,
			&s.TotalMembers,
		); err != nil {
			log.Printf("Scan error: %v", err)
			continue
		}
		s.AdminPassword = ""
		stores = append(stores, s)
	}
	return stores, nil
}

func (r *CardRepository) UpdateSkin(storeID, skin string) error {
	_, err := r.db.Exec("UPDATE stores SET card_skin = $1 WHERE id = $2", skin, storeID)
	return err
}

func (r *CardRepository) UpdatePassword(storeID, newPass string) error {
	_, err := r.db.Exec("UPDATE stores SET admin_password=$1 WHERE id=$2", newPass, storeID)
	return err
}

func (r *CardRepository) VerifyPassword(pass string) bool { return true }

// --- SKINS (Global Assets) ---

func (r *CardRepository) GetAllSkins() ([]models.Skin, error) {
	query := "SELECT id, name, type, image_data, color_bg, color_text, color_border, is_global, store_id, start_date, end_date FROM skins ORDER BY created_at DESC"
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skins []models.Skin
	for rows.Next() {
		var s models.Skin
		var img, cBg, cTxt, cBorder, sId sql.NullString
		var start, end sql.NullTime

		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &img, &cBg, &cTxt, &cBorder, &s.IsGlobal, &sId, &start, &end); err != nil {
			continue
		}

		s.ImageData, s.ColorBg, s.ColorText, s.ColorBorder = img.String, cBg.String, cTxt.String, cBorder.String
		if sId.Valid {
			temp := sId.String
			s.StoreID = &temp
		}
		if start.Valid {
			s.StartDate = &start.Time
		}
		if end.Valid {
			s.EndDate = &end.Time
		}
		skins = append(skins, s)
	}
	return skins, nil
}

func (r *CardRepository) GetAvailableSkins() ([]models.Skin, error) {
	query := `
        SELECT id, name, type, image_data, color_bg, color_text, color_border, is_global, start_date, end_date 
        FROM skins 
        WHERE is_global = TRUE 
        AND (type = 'standard' OR (type = 'seasonal' AND start_date <= CURRENT_TIMESTAMP AND end_date >= CURRENT_TIMESTAMP))`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var skins []models.Skin
	for rows.Next() {
		var s models.Skin
		var img, cBg, cTxt, cBorder sql.NullString
		var start, end sql.NullTime
		if err := rows.Scan(&s.ID, &s.Name, &s.Type, &img, &cBg, &cTxt, &cBorder, &s.IsGlobal, &start, &end); err != nil {
			continue
		}
		s.ImageData, s.ColorBg, s.ColorText, s.ColorBorder = img.String, cBg.String, cTxt.String, cBorder.String
		if start.Valid {
			s.StartDate = &start.Time
		}
		if end.Valid {
			s.EndDate = &end.Time
		}
		skins = append(skins, s)
	}
	return skins, nil
}

func (r *CardRepository) SaveSkin(s models.Skin) error {
	// 1. Garantir que tem um ID antes de procurar
	if s.ID == "" {
		s.ID = uuid.New().String()
	}

	var exists bool
	var oldStoreID sql.NullString

	// 2. Verifica se a Skin já existia
	err := r.db.QueryRow("SELECT store_id FROM skins WHERE id=$1", s.ID).Scan(&oldStoreID)
	if err == sql.ErrNoRows {
		exists = false
	} else if err == nil {
		exists = true
	}

	query := `
        INSERT INTO skins (id, name, type, image_data, color_bg, color_text, color_border, is_global, store_id, start_date, end_date, created_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, CURRENT_TIMESTAMP)
        ON CONFLICT(id) DO UPDATE SET
            name=excluded.name, type=excluded.type, image_data=excluded.image_data,
            color_bg=excluded.color_bg, color_text=excluded.color_text, color_border=excluded.color_border,
            is_global=excluded.is_global, store_id=excluded.store_id, 
            start_date=excluded.start_date, end_date=excluded.end_date;`

	var storeID interface{} = nil
	if s.StoreID != nil && *s.StoreID != "" {
		storeID = *s.StoreID
	}

	_, err = r.db.Exec(query, s.ID, s.Name, s.Type, s.ImageData, s.ColorBg, s.ColorText, s.ColorBorder, s.IsGlobal, storeID, s.StartDate, s.EndDate)
	if err != nil {
		log.Printf("❌ Erro SQL ao guardar Skin: %v", err)
		return err
	}

	// --- AUTOMAÇÃO DE NOTIFICAÇÕES ---
	visual := s.ImageData
	if visual == "" {
		visual = s.ColorBg
	}

	log.Printf("ℹ️ DETEÇÃO SKIN: ID=%s | Nome=%s | Nova=%v | Global=%v | Tipo=%s", s.ID, s.Name, !exists, s.IsGlobal, s.Type)

	if !exists {
		if s.IsGlobal {
			log.Println("📢 A disparar Broadcast de Skin 100% NOVA!")
			r.BroadcastStoreNotification("🎨 Novo Cartão: "+s.Name, "Lançámos um novo design na plataforma! Vá ao menu 'Estilo do Cartão' para o ativar.", "success", visual)
		} else if storeID != nil {
			r.SendStoreNotification(*s.StoreID, "🎁 Design Exclusivo: "+s.Name, "Foi-lhe atribuído um novo estilo de cartão exclusivo. Verifique o menu 'Estilo do Cartão'.", "success", visual)
		}
	} else {
		if storeID != nil && (!oldStoreID.Valid || oldStoreID.String != *s.StoreID) {
			r.SendStoreNotification(*s.StoreID, "🎁 Novo Design Exclusivo: "+s.Name, "A equipa Volto acabou de lhe atribuir um estilo de cartão exclusivo!", "success", visual)
		}

		// Aceita minúsculas ou maiúsculas ("seasonal" ou "Seasonal")
		if s.IsGlobal && strings.ToLower(s.Type) == "seasonal" {
			log.Println("📢 A disparar Broadcast de Skin SAZONAL atualizada!")
			r.BroadcastStoreNotification("🗓️ Época Especial: "+s.Name, "O cartão sazonal de época está disponível para uso na sua loja!", "info", visual)
		}
	}

	return nil
}

func (r *CardRepository) DeleteSkin(id string) error {
	if id == "default" || id == "black" {
		return fmt.Errorf("cannot delete system skin")
	}
	_, err := r.db.Exec("DELETE FROM skins WHERE id = $1", id)
	return err
}

// --- CARTÕES (Clientes) ---

func (r *CardRepository) GetCardByID(id string) (*models.LoyaltyCard, error) {
	var c models.LoyaltyCard
	var email, phone, nif, design sql.NullString

	// NOVAS VARIÁVEIS PARA O ÂMBITO
	var scopeID, scopeName, scopeIcon sql.NullString
	var scopeIsActive sql.NullBool

	query := `
        SELECT c.id, c.store_id, c.member_number, c.customer_id, c.last_name, c.email, c.phone, c.nif, 
               c.stamps_count, c.total_stamps, c.total_redeemed_bonuses, c.is_reward_ready, c.design, 
               c.rgpd_accepted, c.marketing_accepted,
               c.scope_id, sc.name, sc.stamp_icon, sc.is_active
        FROM loyalty_cards c
        LEFT JOIN store_scopes sc ON c.scope_id = sc.id
        WHERE c.id = $1`

	err := r.db.QueryRow(query, id).Scan(
		&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif,
		&c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design,
		&c.RgpdAccepted, &c.MarketingAccepted,
		&scopeID, &scopeName, &scopeIcon, &scopeIsActive,
	)
	if err != nil {
		return nil, err
	}

	c.Email, c.Phone, c.NIF, c.Design = email.String, phone.String, nif.String, design.String
	c.ScopeID, c.ScopeName, c.ScopeIcon = scopeID.String, scopeName.String, scopeIcon.String
	if scopeIsActive.Valid {
		c.ScopeIsActive = scopeIsActive.Bool
	} else {
		c.ScopeIsActive = true
	}

	return &c, nil
}

func (r *CardRepository) GetCardByEmailOrPhone(storeID, identifier string) (*models.LoyaltyCard, error) {
	var c models.LoyaltyCard
	var email, phone, nif, design, scopeID sql.NullString

	query := `
        SELECT id, store_id, member_number, customer_id, last_name, email, phone, nif, 
               stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, design, scope_id 
        FROM loyalty_cards 
        WHERE (email = $1 OR phone = $1) AND store_id = $2 LIMIT 1`

	err := r.db.QueryRow(query, identifier, storeID).Scan(
		&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif,
		&c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design, &scopeID,
	)
	if err != nil {
		return nil, err
	}
	c.Email, c.Phone, c.NIF, c.Design, c.ScopeID = email.String, phone.String, nif.String, design.String, scopeID.String
	return &c, nil
}

func (r *CardRepository) UseReward(id string) error {
	query := `UPDATE loyalty_cards SET total_redeemed_bonuses = total_redeemed_bonuses + 1, is_reward_ready = CASE WHEN (total_stamps / 10) - (total_redeemed_bonuses + 1) > 0 THEN TRUE ELSE FALSE END, updated_at = CURRENT_TIMESTAMP WHERE id = $1 AND (total_stamps / 10) > total_redeemed_bonuses`
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

func (r *CardRepository) AddStamp(id string) (*models.LoyaltyCard, error) {
	query := `UPDATE loyalty_cards SET total_stamps = total_stamps + 1, stamps_count = CASE WHEN stamps_count >= 10 THEN 1 ELSE stamps_count + 1 END, is_reward_ready = CASE WHEN stamps_count = 9 OR (stamps_count = 10 AND is_reward_ready = true) THEN TRUE ELSE FALSE END, updated_at = CURRENT_TIMESTAMP WHERE id = $1`
	_, err := r.db.Exec(query, id)
	if err != nil {
		return nil, err
	}
	return r.GetCardByID(id)
}

func (r *CardRepository) GetAllCards(storeID string) ([]models.LoyaltyCard, error) {
	// ATUALIZADO: Fomos buscar também o sc.stamp_icon
	query := `
        SELECT c.id, c.store_id, c.member_number, c.customer_id, c.last_name, c.email, c.phone, c.nif, 
               c.stamps_count, c.total_stamps, c.total_redeemed_bonuses, c.is_reward_ready, c.design, 
               c.rgpd_accepted, c.marketing_accepted, c.scope_id, sc.name, sc.stamp_icon
        FROM loyalty_cards c
        LEFT JOIN store_scopes sc ON c.scope_id = sc.id
        WHERE c.store_id = $1 
        ORDER BY c.member_number DESC`

	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.LoyaltyCard
	for rows.Next() {
		var c models.LoyaltyCard
		var email, phone, nif, design, scopeID, scopeName, scopeIcon sql.NullString
		err := rows.Scan(
			&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif,
			&c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready, &design,
			&c.RgpdAccepted, &c.MarketingAccepted, &scopeID, &scopeName, &scopeIcon,
		)
		if err != nil {
			continue
		}
		c.Email, c.Phone, c.NIF, c.Design = email.String, phone.String, nif.String, design.String
		c.ScopeID, c.ScopeName, c.ScopeIcon = scopeID.String, scopeName.String, scopeIcon.String
		cards = append(cards, c)
	}
	return cards, nil
}

func (r *CardRepository) SaveCard(card models.LoyaltyCard) error {
	if card.Design == "" {
		card.Design = "default"
	}

	// ATUALIZADO: Inserir scope_id
	query := `
	    INSERT INTO loyalty_cards (
	        id, store_id, customer_id, last_name, email, phone, nif, 
	        stamps_count, total_stamps, total_redeemed_bonuses, is_reward_ready, 
	        design, rgpd_accepted, marketing_accepted, scope_id, consent_date, updated_at
	    ) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, 0, $10, $11, $12, $13, NULLIF($14, ''), CURRENT_TIMESTAMP, CURRENT_TIMESTAMP)`

	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}

	_, err := r.db.Exec(query,
		card.ID, card.StoreID, card.CustomerID, card.LastName,
		toNull(card.Email), toNull(card.Phone), toNull(card.NIF),
		card.StampsCount, card.TotalStamps, card.Is_reward_ready,
		card.Design, card.RgpdAccepted, card.MarketingAccepted, card.ScopeID)
	return err
}

func (r *CardRepository) UpdateCard(card models.LoyaltyCard) error {
	query := `UPDATE loyalty_cards SET customer_id=$1, last_name=$2, email=$3, phone=$4, nif=$5, scope_id=NULLIF($6, ''), updated_at=CURRENT_TIMESTAMP WHERE id=$7`
	toNull := func(s string) interface{} {
		if s == "" {
			return nil
		}
		return s
	}
	_, err := r.db.Exec(query, card.CustomerID, card.LastName, toNull(card.Email), toNull(card.Phone), toNull(card.NIF), card.ScopeID, card.ID)
	return err
}

func (r *CardRepository) ResetCard(id string) error {
	query := `UPDATE loyalty_cards SET stamps_count=0, total_stamps=0, total_redeemed_bonuses=0, is_reward_ready=false, updated_at=CURRENT_TIMESTAMP WHERE id=$1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *CardRepository) SearchCards(storeID, term string) ([]models.LoyaltyCard, error) {
	query := `
	    SELECT c.id, c.store_id, c.member_number, c.customer_id, c.last_name, c.email, c.phone, c.nif, 
	           c.stamps_count, c.total_stamps, c.total_redeemed_bonuses, c.is_reward_ready, 
               c.rgpd_accepted, c.marketing_accepted, sc.name, sc.stamp_icon
	    FROM loyalty_cards c
        LEFT JOIN store_scopes sc ON c.scope_id = sc.id
	    WHERE c.store_id = $1 AND (c.customer_id LIKE $2 OR c.last_name LIKE $2 OR c.email LIKE $2 OR c.phone LIKE $2 OR c.nif LIKE $2) 
	    ORDER BY c.member_number DESC LIMIT 15`

	rows, err := r.db.Query(query, storeID, "%"+term+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var cards []models.LoyaltyCard
	for rows.Next() {
		var c models.LoyaltyCard
		var email, phone, nif, scopeName, scopeIcon sql.NullString
		rows.Scan(&c.ID, &c.StoreID, &c.MemberNumber, &c.CustomerID, &c.LastName, &email, &phone, &nif,
			&c.StampsCount, &c.TotalStamps, &c.TotalRedeemedBonuses, &c.Is_reward_ready,
			&c.RgpdAccepted, &c.MarketingAccepted, &scopeName, &scopeIcon)
		c.Email, c.Phone, c.NIF, c.ScopeName, c.ScopeIcon = email.String, phone.String, nif.String, scopeName.String, scopeIcon.String
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

func (r *CardRepository) ToggleStoreStatus(id string, status bool) error {
	_, err := r.db.Exec("UPDATE stores SET is_active = $1 WHERE id = $2", status, id)
	return err
}

func (repo *CardRepository) UpdateStore(s models.Store) error {

	if s.NewPassword != "" {
		hashed, err := bcrypt.GenerateFromPassword([]byte(s.NewPassword), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		query := `
            UPDATE stores 
            SET name=$1, tier=$2, tier_expiration=$3, billing_cycle=$4, 
                logo_url=$5, card_skin=$6, primary_color=$7, stamp_icon=$8, 
                admin_password=$9, updated_at=CURRENT_TIMESTAMP
            WHERE id=$10
        `
		_, err = repo.db.Exec(query, s.Name, s.Tier, s.TierExpiration, s.BillingCycle, s.LogoURL, s.CardSkin, s.PrimaryColor, s.StampIcon, string(hashed), s.ID)
		return err
	}

	query := `
        UPDATE stores 
        SET name=$1, tier=$2, tier_expiration=$3, billing_cycle=$4, 
            logo_url=$5, card_skin=$6, primary_color=$7, stamp_icon=$8, 
            updated_at=CURRENT_TIMESTAMP
        WHERE id=$9
    `
	_, err := repo.db.Exec(query, s.Name, s.Tier, s.TierExpiration, s.BillingCycle, s.LogoURL, s.CardSkin, s.PrimaryColor, s.StampIcon, s.ID)
	return err
}

func (repo *CardRepository) RegisterStore(req models.RegisterStoreRequest) (string, string, error) {
	hashed, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	id := uuid.New().String()
	expiration := time.Now().AddDate(0, 0, 30)

	// 👇 Gera Token (Usa a função generateSecureToken() que já lá tens)
	token := generateSecureToken()
	tokenExpires := time.Now().Add(30 * time.Minute)

	// Inserimos account_activated = FALSE e os dados do token
	query := `
        INSERT INTO stores (
            id, name, slug, 
            admin_username, admin_email, admin_password, 
            tier, tier_expiration, billing_cycle, max_users, 
            account_activated, is_active, 
            card_skin, primary_color, stamp_icon, theme_mode,
            verification_token, token_expires_at
        ) VALUES (
            $1, $2, $3, 
            $4, $4, $5, 
            'free_trial', $6, 'monthly', 1, 
            FALSE, TRUE, 
            'default', '#00a896', '🍳', 'dark',
            $7, $8
        )
    `
	_, err = repo.db.Exec(query,
		id, req.Name, req.Slug,
		req.Email, string(hashed),
		expiration, token, tokenExpires,
	)

	if err != nil {
		return "", "", err
	}

	// Auto-cria o âmbito principal para a nova loja
	scopeQuery := `INSERT INTO store_scopes (id, store_id, name, stamp_icon, is_main, is_active) VALUES ($1, $2, 'Geral', '💳', TRUE, TRUE)`
	repo.db.Exec(scopeQuery, uuid.New().String(), id)

	return id, token, nil
}

func (repo *CardRepository) AuthenticateStore(email, password string) (*models.Store, error) {
	var store models.Store
	var hashedPassword string
	//var isActivated bool

	// Lemos também o account_activated
	query := `
        SELECT id, slug, admin_password, tier, is_active, account_activated 
        FROM stores 
        WHERE admin_email = $1
    `

	err := repo.db.QueryRow(query, email).Scan(
		&store.ID,
		&store.Slug,
		&hashedPassword,
		&store.Tier,
		&store.IsActive,
		&store.AccountActivated,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("email not found")
		}
		return nil, err
	}

	// 👇 PROTEÇÃO CONTRA LOJAS NÃO VERIFICADAS 👇
	// if !isActivated {
	// 	return nil, fmt.Errorf("unverified")
	// }

	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	return &store, nil
}

func (r *CardRepository) UpdateSettings(s models.Store) error {
	query := `
        UPDATE stores 
        SET 
            name = $1, primary_color = $2, stamp_icon = $3,
            bronze_threshold = $4, silver_threshold = $5, gold_threshold = $6,
            logo_url = $7, theme_mode = $8, text_color = $9, border_color = $10,
            card_image_url = $11, card_image_zoom = $12,
            card_image_pos_x = $13, card_image_pos_y = $14, card_scope = $15,
            social_instagram = $16, social_facebook = $17, social_twitter = $18, 
            social_whatsapp = $19, social_tiktok = $20, social_youtube = $21, social_website = $22,
            menu_url = $23, location_url = $24
        WHERE id = $25
    `
	_, err := r.db.Exec(query,
		s.Name, s.PrimaryColor, s.StampIcon, s.Bronze, s.Silver, s.Gold,
		s.LogoURL, s.ThemeMode, s.TextColor, s.BorderColor, s.CardImageUrl,
		s.CardImageZoom, s.CardImagePosX, s.CardImagePosY, s.CardScope,
		s.SocialInstagram, s.SocialFacebook, s.SocialTwitter, s.SocialWhatsapp, s.SocialTiktok, s.SocialYoutube, s.SocialWebsite,
		s.MenuUrl, s.LocationUrl,
		s.ID,
	)
	return err
}

// --- GESTÃO DE UTILIZADORES GLOBAIS (VOLTO WALLET) ---

func generateSecureToken() string {
	b := make([]byte, 16) // 32 caracteres em Hex
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (repo *CardRepository) CreateGlobalUser(user models.GlobalUser) (string, error) {
	// 👇 1. ENCRIPTAR A PASSWORD DO CLIENTE
	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	// 2. Gerar o token de verificação de email
	token := generateSecureToken()
	tokenExpires := time.Now().Add(30 * time.Minute)

	// 3. Guardar na BD com a password já encriptada (string(hashed))
	query := `
		INSERT INTO global_users (
			id, first_name, last_name, email, phone, 
			password, is_verified, verification_token, token_expires_at, 
			rgpd_accepted, marketing_accepted
		) VALUES ($1, $2, $3, $4, $5, $6, FALSE, $7, $8, $9, $10)
	`

	_, err = repo.db.Exec(query,
		user.ID, user.FirstName, user.LastName, user.Email, user.Phone,
		string(hashed), // 🔐 PASSWORD ENCRIPTADA AQUI
		token, tokenExpires, user.RgpdAccepted, user.MarketingAccepted,
	)

	if err != nil {
		return "", err
	}

	return token, nil
}

func (r *CardRepository) VerifyGlobalUser(token string) error {
	// Procura o utilizador pelo token e verifica se não expirou
	query := `
        UPDATE global_users 
        SET is_verified = TRUE, verification_token = NULL, token_expires_at = NULL 
        WHERE verification_token = $1 AND token_expires_at > CURRENT_TIMESTAMP
    `
	res, err := r.db.Exec(query, token)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("token inválido ou expirado")
	}

	return nil
}

func (repo *CardRepository) AuthenticateGlobalUser(email, password string) (*models.GlobalUser, error) {
	var user models.GlobalUser
	var hashedPassword string
	//var isVerified bool

	query := `
		SELECT id, first_name, last_name, email, password, is_verified 
		FROM global_users 
		WHERE email = $1
	`

	err := repo.db.QueryRow(query, email).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&hashedPassword,
		&user.IsVerified,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}

	// 👇 1. Verifica se já clicou no link do email
	// if !isVerified {
	// 	return nil, fmt.Errorf("unverified")
	// }

	// 👇 2. COMPARA A PASSWORD ENCRIPTADA COM A QUE ELE ESCREVEU
	err = bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	if err != nil {
		return nil, fmt.Errorf("invalid password")
	}

	// 3. Limpa a password da struct antes de devolver (por segurança na memória)
	user.Password = ""

	return &user, nil
}

func (r *CardRepository) GetGlobalUserByID(id string) (*models.GlobalUser, error) {
	var u models.GlobalUser
	query := `SELECT id, first_name, last_name, email, phone, is_verified FROM global_users WHERE id = $1`
    err := r.db.QueryRow(query, id).Scan(&u.ID, &u.FirstName, &u.LastName, &u.Email, &u.Phone, &u.IsVerified)
    return &u, err
}

func (r *CardRepository) GetMyWalletCards(email string) ([]map[string]interface{}, error) {
	query := `
        SELECT c.id, c.total_stamps, c.total_redeemed_bonuses, s.name, s.slug, s.logo_url, s.primary_color,
               sc.name, sc.stamp_icon, sc.is_active
        FROM loyalty_cards c
        JOIN stores s ON c.store_id = s.id
        LEFT JOIN store_scopes sc ON c.scope_id = sc.id
        WHERE c.email = $1
    `
	rows, err := r.db.Query(query, email)
	if err != nil {
		fmt.Println("❌ ERRO SQL GetMyWalletCards:", err)
		return nil, err
	}
	defer rows.Close()

	var myCards []map[string]interface{}
	for rows.Next() {
		var id, storeName, storeSlug string
		var totalStamps, totalRedeemed int
		var logoUrl, primaryColor, scopeName, scopeIcon sql.NullString
		var scopeIsActive sql.NullBool

		if err := rows.Scan(&id, &totalStamps, &totalRedeemed, &storeName, &storeSlug, &logoUrl, &primaryColor, &scopeName, &scopeIcon, &scopeIsActive); err != nil {
			fmt.Println("⚠️ Erro a ler linha do cartão:", err)
			continue
		}

		scopeActive := true
		if scopeIsActive.Valid {
			scopeActive = scopeIsActive.Bool
		}

		myCards = append(myCards, map[string]interface{}{
			"id":                     id,
			"total_stamps":           totalStamps,
			"total_redeemed_bonuses": totalRedeemed,
			"store_name":             storeName,
			"store_slug":             storeSlug,
			"logo_url":               logoUrl.String,
			"primary_color":          primaryColor.String,
			"scope_name":             scopeName.String,
			"scope_icon":             scopeIcon.String,
			"scope_is_active":        scopeActive,
		})
	}
	return myCards, nil
}

func (r *CardRepository) GetStoreScopes(storeID string) ([]models.StoreScope, error) {
	query := `SELECT id, store_id, name, stamp_icon, is_main, is_active FROM store_scopes WHERE store_id = $1 ORDER BY is_main DESC, created_at ASC`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var scopes []models.StoreScope
	for rows.Next() {
		var s models.StoreScope
		if err := rows.Scan(&s.ID, &s.StoreID, &s.Name, &s.StampIcon, &s.IsMain, &s.IsActive); err == nil {
			scopes = append(scopes, s)
		}
	}
	return scopes, nil
}

func (r *CardRepository) CreateStoreScope(scope models.StoreScope) error {
	query := `INSERT INTO store_scopes (id, store_id, name, stamp_icon, is_main, is_active) VALUES ($1, $2, $3, $4, $5, $6)`
	_, err := r.db.Exec(query, scope.ID, scope.StoreID, scope.Name, scope.StampIcon, scope.IsMain, scope.IsActive)

	// UPSELL AUTOMÁTICO: Se não for o cartão principal, avisa todos os clientes desta loja!
	if err == nil && !scope.IsMain {
		rows, _ := r.db.Query(`SELECT DISTINCT email FROM loyalty_cards WHERE store_id = $1 AND email IS NOT NULL AND email != ''`, scope.StoreID)
		var emails []string
		for rows.Next() {
			var e string
			rows.Scan(&e)
			emails = append(emails, e)
		}
		rows.Close()

		title := "✨ Novo Cartão Disponível!"
		msg := fmt.Sprintf("Acabámos de lançar o cartão '%s %s'. Peça para aderir na sua próxima visita e ganhe prémios!", scope.StampIcon, scope.Name)
		r.SendNotificationToEmails(emails, scope.StoreID, title, msg, "success")
	}
	return err
}

func (r *CardRepository) ToggleScopeStatus(scopeID string, storeID string, isActive bool) error {
	// Faz o update e devolve o nome do âmbito para usarmos na mensagem
	query := `UPDATE store_scopes SET is_active = $1 WHERE id = $2 AND store_id = $3 AND is_main = FALSE RETURNING name`
	var scopeName string
	err := r.db.QueryRow(query, isActive, scopeID, storeID).Scan(&scopeName)

	if err == nil {
		rows, _ := r.db.Query(`SELECT DISTINCT email FROM loyalty_cards WHERE scope_id = $1 AND email IS NOT NULL AND email != ''`, scopeID)
		var emails []string
		for rows.Next() {
			var e string
			rows.Scan(&e)
			emails = append(emails, e)
		}
		rows.Close()

		if !isActive {
			r.SendNotificationToEmails(emails, storeID, "⏸️ Cartão Pausado", fmt.Sprintf("O programa '%s' foi temporariamente pausado. Os seus selos estão salvaguardados.", scopeName), "warning")
		} else {
			r.SendNotificationToEmails(emails, storeID, "▶️ Cartão Reativado", fmt.Sprintf("O programa '%s' voltou a estar ativo. Já pode voltar a ganhar selos!", scopeName), "info")
		}
	}
	return err
}

func (r *CardRepository) UpdateStoreScope(id, storeID, name, icon string) error {
	query := `UPDATE store_scopes SET name = $1, stamp_icon = $2 WHERE id = $3 AND store_id = $4`
	_, err := r.db.Exec(query, name, icon, id, storeID)
	return err
}

func (r *CardRepository) DeleteStoreScope(scopeID string, storeID string) error {
	// COMO VAMOS APAGAR, TEMOS DE IR BUSCAR OS EMAILS *ANTES* DO CASCADE DESTRUIR OS CARTÕES!
	var scopeName string
	r.db.QueryRow(`SELECT name FROM store_scopes WHERE id = $1`, scopeID).Scan(&scopeName)

	rows, _ := r.db.Query(`SELECT DISTINCT email FROM loyalty_cards WHERE scope_id = $1 AND email IS NOT NULL AND email != ''`, scopeID)
	var emails []string
	for rows.Next() {
		var e string
		rows.Scan(&e)
		emails = append(emails, e)
	}
	rows.Close()

	query := `DELETE FROM store_scopes WHERE id = $1 AND store_id = $2 AND is_main = FALSE`
	res, err := r.db.Exec(query, scopeID, storeID)

	if err == nil {
		rowsAff, _ := res.RowsAffected()
		if rowsAff > 0 && scopeName != "" {
			r.SendNotificationToEmails(emails, storeID, "🔴 Programa Encerrado", fmt.Sprintf("O cartão '%s' foi descontinuado. Este cartão já não aparecerá na sua carteira.", scopeName), "error")
		}
	}
	return err
}

// --- NOTIFICAÇÕES DA WALLET ---

func (r *CardRepository) GetUserNotifications(email string) ([]models.WalletNotification, error) {
	query := `
		SELECT n.id, n.email, n.store_id, s.name, COALESCE(s.logo_url, ''), n.title, n.message, n.type, n.is_read, n.created_at
		FROM wallet_notifications n
		JOIN stores s ON n.store_id = s.id
		WHERE n.email = $1
		ORDER BY n.created_at DESC LIMIT 50
	`
	rows, err := r.db.Query(query, email)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.WalletNotification
	for rows.Next() {
		var n models.WalletNotification
		rows.Scan(&n.ID, &n.Email, &n.StoreID, &n.StoreName, &n.StoreLogo, &n.Title, &n.Message, &n.Type, &n.IsRead, &n.CreatedAt)
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *CardRepository) MarkNotificationsAsRead(email string) error {
	_, err := r.db.Exec(`UPDATE wallet_notifications SET is_read = TRUE WHERE email = $1`, email)
	return err
}

func (r *CardRepository) SendNotificationToEmails(emails []string, storeID, title, message, nType string) {
	for _, email := range emails {
		if email == "" {
			continue
		}
		id := uuid.New().String()
		r.db.Exec(`INSERT INTO wallet_notifications (id, email, store_id, title, message, type) VALUES ($1, $2, $3, $4, $5, $6)`, id, email, storeID, title, message, nType)
	}
}

// --- MENSAGENS PARA OS LOJISTAS (ADMIN NOTIFICATIONS) ---

func (r *CardRepository) GetStoreNotifications(storeID string) ([]models.StoreNotification, error) {
	query := `SELECT id, store_id, title, message, type, is_read, created_at, COALESCE(image_data, '') FROM store_notifications WHERE store_id = $1 ORDER BY created_at DESC LIMIT 50`
	rows, err := r.db.Query(query, storeID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifs []models.StoreNotification
	for rows.Next() {
		var n models.StoreNotification
		rows.Scan(&n.ID, &n.StoreID, &n.Title, &n.Message, &n.Type, &n.IsRead, &n.CreatedAt, &n.ImageData)
		notifs = append(notifs, n)
	}
	return notifs, nil
}

func (r *CardRepository) MarkStoreNotificationsAsRead(storeID string) error {
	_, err := r.db.Exec(`UPDATE store_notifications SET is_read = TRUE WHERE store_id = $1`, storeID)
	return err
}

func (r *CardRepository) SendStoreNotification(storeID, title, message, nType, imageData string) error {
	id := uuid.New().String()
	_, err := r.db.Exec(`INSERT INTO store_notifications (id, store_id, title, message, type, image_data) VALUES ($1, $2, $3, $4, $5, NULLIF($6, ''))`, id, storeID, title, message, nType, imageData)

	// SE FALHAR NA BD, VAI GRITAR NO TERMINAL!
	if err != nil {
		log.Printf("❌ ERRO SQL NA NOTIFICAÇÃO (Loja: %s): %v", storeID, err)
	}
	return err
}

func (r *CardRepository) BroadcastStoreNotification(title, message, nType, imageData string) error {
	rows, err := r.db.Query(`SELECT id FROM stores`)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var storeID string
		if err := rows.Scan(&storeID); err == nil {
			r.SendStoreNotification(storeID, title, message, nType, imageData)
		}
	}
	return nil
}

// CleanupUnverifiedUsers apaga utilizadores que não confirmaram o email após 30 mins
func (r *CardRepository) CleanupUnverifiedUsers() {
	// Apaga Wallet Users
	res1, _ := r.db.Exec(`DELETE FROM global_users WHERE is_verified = FALSE AND token_expires_at < CURRENT_TIMESTAMP`)

	// Apaga Lojas
	res2, _ := r.db.Exec(`DELETE FROM stores WHERE account_activated = FALSE AND token_expires_at < CURRENT_TIMESTAMP`)

	rows1, _ := res1.RowsAffected()
	rows2, _ := res2.RowsAffected()
	if rows1 > 0 || rows2 > 0 {
		log.Printf("🧹 Limpeza automática: %d Wallets e %d Lojas não verificadas foram removidas.", rows1, rows2)
	}
}

// StartCleanupWorker arranca a Goroutine que corre de 5 em 5 minutos
func (r *CardRepository) StartCleanupWorker() {
	go func() {
		for {
			r.CleanupUnverifiedUsers()
			// Dorme durante 5 minutos antes de voltar a tentar
			time.Sleep(5 * time.Minute)
		}
	}()
}

func (repo *CardRepository) VerifyStoreEmail(token string) error {
	query := `
        UPDATE stores 
        SET account_activated = TRUE, verification_token = NULL, token_expires_at = NULL 
        WHERE verification_token = $1 AND token_expires_at > CURRENT_TIMESTAMP
    `
	res, err := repo.db.Exec(query, token)
	if err != nil {
		return err
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("token inválido ou expirado")
	}

	return nil
}
