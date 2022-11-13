package views

import (
	"context"
	"encoding/hex"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/utils/accrual"
	"net/http"
	"time"
)

const CtxSecond = 2 * time.Second

type Server struct {
	Storage *database.Storage
}

func NewServer() Server {
	s := Server{
		Storage: database.NewStorage(),
	}
	accrual.StartWorker(s.Storage)
	return s
}

func (server Server) SetCookieToResponse(token string, w http.ResponseWriter) {
	encoded := hex.EncodeToString([]byte(token))
	cookie := &http.Cookie{Name: "auth_token", Value: encoded, HttpOnly: true}
	http.SetCookie(w, cookie)
}

func (server Server) SetCookieToRequest(token string, r *http.Request) {
	encoded := hex.EncodeToString([]byte(token))
	cookie := &http.Cookie{Name: "auth_token", Value: encoded, HttpOnly: true}
	r.AddCookie(cookie)
}

func (server Server) GetUserIDFromRequest(
	ctx context.Context, r *http.Request,
) (userID string) {
	data, err := r.Cookie("auth_token")
	if err != nil {
		logger.Logger.Error(err.Error())
		return ""
	}

	encodedToken := data.Value
	decodedToken, err := hex.DecodeString(encodedToken)
	if err != nil {
		logger.Logger.Error(err.Error())
		return ""
	}

	userID, err = server.Storage.GetUserIDByToken(ctx, string(decodedToken))
	if err != nil {
		logger.Logger.Error(err.Error())
		return ""
	}

	return userID
}

func (server Server) Write(msg []byte, w http.ResponseWriter) {
	_, err := w.Write(msg)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
}

func (server Server) CloseBody(r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
}
