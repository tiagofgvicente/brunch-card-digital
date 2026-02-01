package api

import (
	"fmt"
	"net/http"

	"brunch-card-digital/internal/database"

	"github.com/skip2/go-qrcode"
)

// GetQRCodeHandler gera um QR Code que aponta para o URL público do cartão
func GetQRCodeHandler(w http.ResponseWriter, r *http.Request, repo *database.CardRepository) {
	// 1. Extrair o ID dos parâmetros da query (?id=UUID)
	cardID := r.URL.Query().Get("id")
	if cardID == "" {
		http.Error(w, "ID parameter is required", http.StatusBadRequest)
		return
	}

	// 2. Validar se o cartão existe na base de dados
	_, err := repo.GetCardByID(cardID)
	if err != nil {
		http.Error(w, "Card ID not found", http.StatusNotFound)
		return
	}

	// 3. Construir o URL que o cliente vai usar
	// detetamos o protocolo (http/https) e o host (localhost:8080 ou domínio real)
	scheme := "http"
	if r.TLS != nil {
		scheme = "https"
	}
	// Usamos o r.Host para que o QR Code funcione tanto em localhost como no Kubernetes
	targetURL := fmt.Sprintf("%s://%s/card?id=%s", scheme, r.Host, cardID)

	// 4. Gerar os bytes do QR Code com o URL completo
	// Nível de recuperação Médio e tamanho de 256px
	png, err := qrcode.Encode(targetURL, qrcode.Medium, 256)
	if err != nil {
		http.Error(w, "Error generating QR Code", http.StatusInternalServerError)
		return
	}

	// 5. Enviar a imagem com os headers corretos
	w.Header().Set("Content-Type", "image/png")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write(png)
}
