package views

import (
	"context"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/theplant/luhn"
	"io"
	"net/http"
	"strconv"
)

// LoadOrder /api/user/orders
func (s *Server) LoadOrder(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ctxSecond)
	defer cancel()
	userID := s.GetRequestUserID(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	defer s.CloseBody(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "body is empty", http.StatusBadRequest)
		return
	}
	orderNumber, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var respMessage string
	if !luhn.Valid(orderNumber) {
		respMessage = fmt.Sprintf("%s wrong number format", string(body))
		http.Error(w, respMessage, http.StatusUnprocessableEntity)
		return
	}

	httpStatus := http.StatusAccepted
	err = s.Storage.CreateOrder(ctx, orderNumber, userID)
	if err != nil {
		respMessage = err.Error()
		errData, ok := err.(*e.OrderAlreadyExistsError)
		switch ok {
		case errData.FromCurrentUser:
			httpStatus = http.StatusOK
		case !errData.FromCurrentUser:
			httpStatus = http.StatusConflict
		case !ok:
			http.Error(w, respMessage, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	s.Write([]byte(respMessage), w)
}
