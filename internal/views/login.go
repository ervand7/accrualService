package views

import (
	"context"
	"encoding/json"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/google/uuid"
	"io"
	"net/http"
)

// Login /api/user/login
func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
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

	var credentials struct {
		Login    string `json:"login"`
		Password string `json:"password"`
	}
	if err = json.Unmarshal(body, &credentials); err != nil {
		logger.Logger.Error(err.Error())
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), CtxSecond)
	defer cancel()

	var respMessage string
	httpStatus := http.StatusOK
	token := uuid.New().String()
	if err = s.Storage.UpdateToken(
		ctx, credentials.Login, credentials.Password, token,
	); err != nil {
		if errData, ok := err.(*e.UserNotFoundError); ok {
			respMessage = errData.Error()
			httpStatus = http.StatusUnauthorized
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
	} else {
		s.SetResponseCookie(token, w)
	}

	w.Header().Add("Content-type", "text/plain; charset=utf-8")
	w.WriteHeader(httpStatus)
	s.Write([]byte(respMessage), w)
}
