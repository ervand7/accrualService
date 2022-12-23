package views

import (
	"context"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/utils"
)

const ctxSecond = 2 * time.Second

type Server struct {
	Storage *database.Storage
}

func NewServer() Server {
	s := Server{
		Storage: database.NewStorage(),
	}
	utils.StartWorker(s.Storage)
	return s
}

func (s Server) SetResponseCookie(token string, w http.ResponseWriter) {
	encoded := hex.EncodeToString([]byte(token))
	cookie := &http.Cookie{Name: "auth_token", Value: encoded, HttpOnly: true}
	http.SetCookie(w, cookie)
}

func (s Server) SetRequestCookie(token string, r *http.Request) {
	encoded := hex.EncodeToString([]byte(token))
	cookie := &http.Cookie{Name: "auth_token", Value: encoded, HttpOnly: true}
	r.AddCookie(cookie)
}

func (s Server) GetRequestUserID(
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

	userID, err = s.Storage.GetUserByToken(ctx, string(decodedToken))
	if err != nil {
		logger.Logger.Error(err.Error())
		return ""
	}

	return userID
}

func (s Server) Write(msg []byte, w http.ResponseWriter) {
	_, err := w.Write(msg)
	if err != nil {
		logger.Logger.Error(err.Error())
	}
}

func (s Server) CloseBody(r *http.Request) {
	err := r.Body.Close()
	if err != nil {
		logger.Logger.Warn(err.Error())
	}
}
