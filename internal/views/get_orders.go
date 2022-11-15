package views

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetUserOrders /api/user/orders
func (server *Server) GetUserOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()
	userID := server.GetUserIDFromRequest(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	userOrders, err := server.Storage.GetUserOrders(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var (
		body       []byte
		httpStatus int
	)
	if userOrders == nil {
		httpStatus = http.StatusNoContent
	} else {
		body, err = json.Marshal(userOrders)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		httpStatus = http.StatusOK
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(httpStatus)
	server.Write(body, w)
}
