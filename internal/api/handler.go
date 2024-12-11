package api

import (
	"encoding/json"
	"net/http"
	"project/internal/service"
)

type Handler struct {
	services *service.Service
}

func NewHandler(services *service.Service) *Handler {
	return &Handler{services: services}
}

func (h *Handler) getHelloWorld(w http.ResponseWriter, r *http.Request) {
	str, err := h.services.GetString()
	if err != nil {
		http.Error(w, "Failed to retrieve string", http.StatusInternalServerError)
		return
	}

	res := map[string]string{"token": str}

	w.Header().Set("Content-Type", "application/json")
	err = json.NewEncoder(w).Encode(res)
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
