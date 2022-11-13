package views

import (
	"context"
	"encoding/json"
	"net/http"
)

// UserGetOrders /api/user/orders
func (server *Server) UserGetOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()
	userID := server.GetUserIDFromRequest(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	ordersInfo, _ := server.Storage.GetUserOrders(ctx, userID)
	var body []byte
	var httpStatus int
	if ordersInfo == nil {
		httpStatus = http.StatusNoContent
	} else {
		marshaled, _ := json.Marshal(ordersInfo)
		body = marshaled
		httpStatus = http.StatusOK
	}

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(httpStatus)
	server.Write(body, w)
}
