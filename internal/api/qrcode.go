package api

import (
	"net/http"

	"brunch-card-digital/internal/database"

	"github.com/skip2/go-qrcode"
)

// GetQRCodeHandler generates a PNG QR code for a specific card ID
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Extract the ID from query parameters (?id=UUID)
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// 2. Security: Validate if the card exists in the database before generating QR
	_, err := repo.GetCardByID(cardID)
	if err != nil {
		// Return 404 if the ID is not found to prevent generating QR codes for non-existent users
		http.Error(w, "Card ID not found in database. Create a new one first.", http.StatusNotFound)
		return
	}

	// 3. Generate QR Code bytes using the skip2/go-qrcode library
	// Medium recovery level and 256px size
	png, err := qrcode.Encode(cardID, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR Code", http.StatusInternalServerError)
		return
	}

	// 4. Set the correct Content-Type so the browser renders it as an image
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600") // Cache for 1 hour for better performance
	w.Write(png)
}
