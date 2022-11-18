package views

import (
	"context"
	"encoding/json"
	"net/http"
)

// GetUserWithdrawals /api/user/withdrawals
func (s *Server) GetUserWithdrawals(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()
	userID := s.GetRequestUserID(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	withdrawals, err := s.Storage.GetUserWithdrawals(ctx, userID)
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
	s.Write(body, w)
}
