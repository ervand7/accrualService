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

// UserLoadOrders /api/user/orders
func (server *Server) UserLoadOrders(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()
	userID := server.GetUserIDFromRequest(ctx, r)
	if userID == "" {
		http.Error(w, "not authorized", http.StatusUnauthorized)
		return
	}

	defer server.CloseBody(r)
	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if len(body) == 0 {
		http.Error(w, "body is empty", http.StatusBadRequest)
		return
	}
	orderNum, err := strconv.Atoi(string(body))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var respMessage string
	if !luhn.Valid(orderNum) {
		respMessage = fmt.Sprintf("%s wrong number format", string(body))
		http.Error(w, respMessage, http.StatusUnprocessableEntity)
		return
	}

	httpStatus := http.StatusAccepted
	err = server.Storage.CreateOrder(ctx, userID, orderNum)
	if err != nil {
		respMessage = err.Error()
		if errData, ok := err.(*e.OrderAlreadyExistsError); ok {
			if errData.FromCurrentUser {
				httpStatus = http.StatusOK
			} else {
				httpStatus = http.StatusConflict
			}
		} else {
			http.Error(w, respMessage, http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	server.Write([]byte(respMessage), w)
}
