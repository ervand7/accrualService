package views

import (
	"context"
	"encoding/hex"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/theplant/luhn"
	"io"
	"net/http"
	"strconv"
)

// UserLoadOrders /api/user/orders
func (server *Server) UserLoadOrders(w http.ResponseWriter, r *http.Request) {
	encodedToken, err := server.GetTokenFromCookie(r)
	if err != nil {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
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

	orderNum, _ := strconv.Atoi(string(body))
	var respMessage string
	if !luhn.Valid(orderNum) {
		respMessage = fmt.Sprintf(
			"%s did not pass the luhn algorithm check", string(body))
		http.Error(w, respMessage, http.StatusUnprocessableEntity)
		return
	}

	decodedToken, err := hex.DecodeString(encodedToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	httpStatus := http.StatusAccepted
	ctx, cancel := context.WithTimeout(r.Context(), 222*CtxSecond)
	defer cancel()
	err = server.Storage.CreateOrder(ctx, string(decodedToken), orderNum)
	if err != nil {
		if errData, ok := err.(*e.OrderAlreadyCreatedByCurrentUserError); ok {
			respMessage = errData.Error()
			httpStatus = http.StatusOK
		}
		if errData, ok := err.(*e.OrderAlreadyCreatedByAnotherUserError); ok {
			respMessage = errData.Error()
			httpStatus = http.StatusConflict
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	}

	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	server.Write([]byte(respMessage), w)
}
