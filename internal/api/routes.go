package api

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	// Маршруты для healthcheck
	mux.HandleFunc("/health/alive", healthcheck)
	mux.HandleFunc("/health/ready", healthcheck)
	mux.HandleFunc("/login", h.getHelloWorld)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Используем зарегистрированные маршруты
		mux.ServeHTTP(w, r)
	})
}

func healthcheck(w http.ResponseWriter, r *http.Request) {
	enableCORS(w, r)
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"success": true,
		"status":  http.StatusOK,
	}
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}
