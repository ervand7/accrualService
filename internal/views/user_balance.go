package views

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserBalance /api/user/balance
func (server *Server) UserBalance(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()
	userID := server.GetUserIDFromRequest(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	balance, err := server.Storage.GetUserBalance(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	body, err := json.Marshal(balance)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	server.Write(body, w)
}
