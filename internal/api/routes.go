package api

import (
	"encoding/json"
	"net/http"
)

func (h *Handler) InitRoutes() http.Handler {
	mux := http.NewServeMux()

	// Wrap handlers with the logging middleware
	mux.Handle("/health/alive", logRequests(http.HandlerFunc(healthcheck)))
	mux.Handle("/health/ready", logRequests(http.HandlerFunc(healthcheck)))
	mux.Handle("/login", logRequests(handleRequest(http.MethodPost, h.doLogin)))
	mux.Handle("/get-conversations", logRequests(parseUserTokenMiddleware(handleRequest(http.MethodPost, h.getConversations))))
	mux.Handle("/get-users", logRequests(handleRequest(http.MethodGet, h.getUsers)))
	mux.Handle("/get-messages", logRequests(parseUserTokenMiddleware(handleRequest(http.MethodPost, h.getMessages))))
	mux.Handle("/send-message", logRequests(parseUserTokenMiddleware(handleRequest(http.MethodPost, h.sendMessage))))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		// Use the registered routes
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

func handleRequest(method string, handlerFunc http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method == method {
			handlerFunc(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	}
}
