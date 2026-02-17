package api

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"regexp"
	"strings"
	"time"

	"brunch-card-digital/internal/database"
	"brunch-card-digital/internal/models"

	"github.com/google/uuid"
	"github.com/skip2/go-qrcode"
)

func getCurrentStore(r *http.Request) *models.Store {
	store, ok := r.Context().Value("current_store").(*models.Store)
	if !ok {
		return nil
	}
	return store
}

// --- MASTER HANDLERS (Create Store) ---

func MasterCreateStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.Store
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	log.Printf("🛠️ Criando loja: %s (User: %s)", req.Name, req.AdminUsername)

	if req.AdminUsername == "" || req.AdminEmail == "" {
		http.Error(w, "Username and Email are required", 400)
		return
	}

	if err := repo.CreateStore(req); err != nil {
		log.Printf("❌ Falha ao criar loja: %v", err)
		http.Error(w, "DB Error: "+err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func MasterListStoresHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	stores, err := repo.GetAllStores()
	if err != nil {
		http.Error(w, "DB Error", 500)
		return
	}
	json.NewEncoder(w).Encode(stores)
}

// --- LOGIN GLOBAL INTELIGENTE ---

func LoginHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		Identifier string `json:"identifier"` // Username ou Email
		Password   string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)

	log.Printf("🌍 GLOBAL LOGIN: Tentativa para '%s'", req.Identifier)

	// 1. É O MASTER?
	if req.Identifier == "master" && req.Password == "superadmin2026" {
		log.Println("✅ Login Master")
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"role": "master", "redirect": "/master",
		})
		return
	}

	// 2. É UM DONO DE LOJA? (Verifica user OU email)
	store, err := repo.GetStoreByLogin(req.Identifier)
	if err == nil && store != nil {
		if req.Password == store.AdminPassword {
			log.Printf("✅ Login Store Admin: '%s' -> %s", req.Identifier, store.Slug)

			http.SetCookie(w, &http.Cookie{
				Name: "session_token", Value: "authenticated_admin", Path: "/", HttpOnly: true, Expires: time.Now().Add(24 * time.Hour),
			})

			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]string{
				"role":     "admin",
				"redirect": fmt.Sprintf("/admin?store=%s", store.Slug),
			})
			return
		}
	}

	// 3. É UM CLIENTE FINAL?
	// Nota: Cliente só entra com nº telemóvel ou email
	slug, cardID, err := repo.GetStoreSlugByCustomerEmail(req.Identifier)
	if err == nil && slug != "" {
		log.Printf("✅ Login Cliente: Redirecionar para %s", slug)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"role":     "customer",
			"redirect": fmt.Sprintf("/card?store=%s&id=%s", slug, cardID),
		})
		return
	}

	log.Println("❌ Falha de Login")
	http.Error(w, "Credenciais inválidas", http.StatusUnauthorized)
}

func LogoutHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	http.SetCookie(w, &http.Cookie{Name: "session_token", Value: "", Path: "/", Expires: time.Unix(0, 0), HttpOnly: true})
	w.WriteHeader(http.StatusOK)
}

// --- RESTO DA API (Sem alterações de lógica, só a compilar com o novo models) ---

func CreateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req models.CreateCardRequest
	json.NewDecoder(r.Body).Decode(&req)

	newCard := models.BrunchCard{
		ID: uuid.New().String(), StoreID: store.ID, CustomerID: req.CustomerID, LastName: req.LastName, Email: req.Email, Phone: req.Phone, NIF: req.NIF, Design: req.Design, RgpdAccepted: req.RgpdAccepted, MarketingAccepted: req.MarketingAccepted,
	}

	if err := repo.SaveCard(newCard); err != nil {
		http.Error(w, "Error saving", 500)
		return
	}
	savedCard, _ := repo.GetCardByID(newCard.ID)
	json.NewEncoder(w).Encode(savedCard)
}

func ListAllCardsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	cards, _ := repo.GetAllCards(store.ID)
	json.NewEncoder(w).Encode(cards)
}

func SearchHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	query := r.URL.Query().Get("q")
	if len(query) < 2 {
		json.NewEncoder(w).Encode([]interface{}{})
		return
	}
	cards, _ := repo.SearchCards(store.ID, query)
	json.NewEncoder(w).Encode(cards)
}

func GetStatusHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	card, err := repo.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Not found", 404)
		return
	}
	json.NewEncoder(w).Encode(card)
}

func StampHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	updatedCard, _ := repo.AddStamp(cardID)
	json.NewEncoder(w).Encode(updatedCard)
}

func UseRewardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	if err := repo.UseReward(r.URL.Query().Get("id")); err != nil {
		http.Error(w, "Error", 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	cardID := r.URL.Query().Get("id")
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	targetURL := fmt.Sprintf("%s://%s/card?store=%s&id=%s", scheme, r.Host, store.Slug, cardID)
	png, _ := qrcode.Encode(targetURL, qrcode.Medium, 256)
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}

func PublicStatsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	stats, _ := repo.GetPublicStats(store.ID)
	json.NewEncoder(w).Encode(stats)
}

func GetSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	json.NewEncoder(w).Encode(store)
}

func UpdateSettingsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r) // Vai buscar a loja atual ao contexto

	var input models.Store
	if err := json.NewDecoder(r.Body).Decode(&input); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}

	// Atualizar os campos com o que vem do Frontend
	store.Name = input.Name
	store.LogoURL = input.LogoURL
	store.PrimaryColor = input.PrimaryColor // <--- Importante
	store.StampIcon = input.StampIcon       // <--- Importante
	store.ThemeMode = input.ThemeMode
	store.Bronze = input.Bronze
	store.Silver = input.Silver
	store.Gold = input.Gold

	// Gravar na BD
	if err := repo.UpdateSettings(*store); err != nil {
		log.Printf("Erro ao atualizar settings: %v", err)
		http.Error(w, "DB Error", 500)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func UpdateSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Design string `json:"design"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	repo.UpdateSkin(store.ID, req.Design)
	w.WriteHeader(http.StatusOK)
}

func UpdatePasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Old string `json:"old"`
		New string `json:"new"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Old != store.AdminPassword {
		http.Error(w, "Wrong old password", 401)
		return
	}
	repo.UpdatePassword(store.ID, req.New)
	w.WriteHeader(http.StatusOK)
}

func VerifyPasswordHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	store := getCurrentStore(r)
	var req struct {
		Password string `json:"password"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	if req.Password == store.AdminPassword {
		w.WriteHeader(http.StatusOK)
	} else {
		http.Error(w, "Unauthorized", 401)
	}
}

func UpdateCardHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.BrunchCard
	json.NewDecoder(r.Body).Decode(&req)
	repo.UpdateCard(req)
	w.WriteHeader(http.StatusOK)
}

func UpdateConsentHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		ID   string `json:"id"`
		Rgpd *bool  `json:"rgpd_accepted"`
		Mkt  *bool  `json:"marketing_accepted"`
	}
	json.NewDecoder(r.Body).Decode(&req)
	repo.UpdateConsent(req.ID, req.Rgpd, req.Mkt)
	w.WriteHeader(http.StatusOK)
}

func AdminResetHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	repo.ResetCard(r.URL.Query().Get("id"))
	w.WriteHeader(http.StatusOK)
}

// Master: Get All Skins
func GetMasterSkinsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	skins, err := repo.GetAllSkins()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(skins)
}

// Master: Save Skin
func SaveSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var s models.Skin
	if err := json.NewDecoder(r.Body).Decode(&s); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if s.ID == "" {
		s.ID = uuid.New().String()
	} // Gera ID se for novo

	if err := repo.SaveSkin(s); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// Master: Delete Skin
func DeleteSkinHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	id := r.URL.Query().Get("id")
	if err := repo.DeleteSkin(id); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

// System: Get Available Skins (Para Admin e Stores)
func GetAvailableSkinsHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	skins, err := repo.GetAvailableSkins()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	json.NewEncoder(w).Encode(skins)
}

func MasterToggleStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req struct {
		ID       string `json:"id"`
		IsActive bool   `json:"isActive"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", 400)
		return
	}
	if err := repo.ToggleStoreStatus(req.ID, req.IsActive); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func MasterUpdateStoreHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// MUDANÇA: Em vez de 'database.Store', usamos 'models.Store'
	var store models.Store

	// Descodificar o JSON
	if err := json.NewDecoder(r.Body).Decode(&store); err != nil {
		http.Error(w, "Invalid input format", http.StatusBadRequest)
		return
	}

	// Validar ID
	if store.ID == "" {
		http.Error(w, "Store ID is required", http.StatusBadRequest)
		return
	}

	// Executar update
	if err := repo.UpdateStore(store); err != nil {
		log.Printf("Error updating store: %v", err)
		http.Error(w, "Failed to update store", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "Store updated successfully"})
}

func PublicRegisterHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	var req models.RegisterStoreRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Dados inválidos", http.StatusBadRequest)
		return
	}

	// Validação básica (Sem Slug)
	if req.Name == "" || req.Email == "" || req.Password == "" {
		http.Error(w, "Nome, Email e Password são obrigatórios", http.StatusBadRequest)
		return
	}

	// --- GERAÇÃO AUTOMÁTICA DO SLUG ---
	baseSlug := generateSlug(req.Name)
	req.Slug = baseSlug

	// Tentar criar. Se der conflito (slug já existe), tentamos adicionar um número.
	// Esta é uma lógica simples de "retry".
	err := repo.RegisterStore(req)
	if err != nil {
		// Se o erro for de duplicado (provavelmente slug repetido), tentamos de novo com sufixo
		// Ex: "padaria" falhou -> tenta "padaria-2"
		req.Slug = fmt.Sprintf("%s-%d", baseSlug, rand.Intn(9999))
		err = repo.RegisterStore(req)

		if err != nil {
			log.Printf("Register Error: %v", err)
			http.Error(w, "Erro ao criar conta. O Email pode já estar em uso.", http.StatusConflict)
			return
		}
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]string{
		"message":  "Loja criada com sucesso!",
		"redirect": "/?store=" + req.Slug,
	})
}

func generateSlug(name string) string {
	// 1. Converter para minúsculas
	slug := strings.ToLower(name)

	// 2. Substituir tudo o que não for letra ou número por hifen
	// (Esta regex simples mantém a-z e 0-9, o resto vira -)
	reg := regexp.MustCompile("[^a-z0-9]+")
	slug = reg.ReplaceAllString(slug, "-")

	// 3. Remover hífens do início e do fim
	slug = strings.Trim(slug, "-")

	// 4. Se ficar vazio (ex: nome era "!!!"), gerar algo aleatório
	if slug == "" {
		slug = fmt.Sprintf("store-%d", rand.Intn(100000))
	}

	return slug
}
