package api

import (
	"net/http"

	"brunch-card-digital/internal/database" // Use your actual module name from go.mod

	"github.com/skip2/go-qrcode"
)

// GetQRCodeHandler generates a PNG QR code for a specific card ID
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// 1. IMPORTANTE: Validar se o cartão existe antes de gerar o QR
	_, err := repo.GetCardByID(cardID)
	if err != nil {
		// Se não encontrar na DB, damos um erro claro em vez de 404 genérico
		http.Error(w, "Card ID not found in database. Create a new one first.", http.StatusNotFound)
		return
	}

	// 2. Se existe, gera o QR
	png, err := qrcode.Encode(cardID, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}
