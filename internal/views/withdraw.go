package views

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"

	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/theplant/luhn"
)

// Withdraw /api/user/balance/withdraw
func (s *Server) Withdraw(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), ctxSecond)
	defer cancel()
	userID := s.GetRequestUserID(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	defer s.CloseBody(r)
	bodyRaw, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(bodyRaw) == 0 {
		http.Error(w, "body is empty", http.StatusBadRequest)
		return
	}

	var body struct {
		Order string
		Sum   float64
	}
	if err = json.Unmarshal(bodyRaw, &body); err != nil {
		logger.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	orderID, err := strconv.Atoi(body.Order)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	var respMessage string
	if !luhn.Valid(orderID) {
		respMessage = fmt.Sprintf("%d wrong number format", orderID)
		http.Error(w, respMessage, http.StatusUnprocessableEntity)
		return
	}

	httpStatus := http.StatusOK
	err = s.Storage.CreateWithdraw(ctx, userID, orderID, body.Sum)
	if err != nil {
		respMessage = err.Error()
		_, ok := err.(*e.NotEnoughMoneyError)
		switch ok {
		case ok:
			httpStatus = http.StatusPaymentRequired
		default:
			httpStatus = http.StatusInternalServerError
		}
	}

	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	s.Write([]byte(respMessage), w)
}
