package views

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetUserWithdrawals /api/user/withdrawals
func (server *Server) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()
	userID := server.GetRequestUserID(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := server.Storage.GetUserWithdrawals(ctx, userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var (
		body       []byte
		httpStatus int
	)
	if withdrawals == nil {
		httpStatus = http.StatusNoContent
	} else {
		body, err = json.Marshal(withdrawals)
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
