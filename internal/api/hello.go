package api

import (
	"encoding/json"
	"net/http"
)

type HelloResponse struct {
	Message string `json:"message" example:"¡Hola desde Fluxio API!"`
	Status  string `json:"status" example:"success"`
}

// HelloHandler godoc
// @Summary Endpoint de prueba
// @Description Endpoint público para probar la API
// @Tags public
// @Accept json
// @Produce json
// @Success 200 {object} HelloResponse
// @Router /hello [get]
func HelloHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	response := HelloResponse{
		Message: "¡Hola desde Fluxio API!",
		Status:  "success",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}