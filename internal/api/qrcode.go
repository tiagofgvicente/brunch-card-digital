package api

import (
	"net/http"

	"brunch-card-digital/internal/database" // Use your actual module name from go.mod

	"github.com/skip2/go-qrcode"
)

// GetQRCodeHandler generates a PNG QR code for a specific card ID
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Get ID from URL query: /api/v1/qrcode?id=UUID
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "Missing card ID parameter", http.StatusBadRequest)
		return
	}

	// 2. (Optional but recommended) Validate if card exists in DB
	// For now, we will just encode the ID to test the visual part

	// 3. Generate QR Code
	// We encode the ID. When scanned, it leads to our validation endpoint
	var png []byte
	png, err := qrcode.Encode(cardID, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Failed to generate QR Code", http.StatusInternalServerError)
		return
	}

	// 4. Return the image
	w.Header().Set("Content-Type", "image/png")
	w.Write(png)
}
