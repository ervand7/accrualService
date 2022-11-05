package views

import (
	"encoding/hex"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/controllers"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"net/http"
	"time"
)

const CtxSecond = 2 * time.Second

type Server struct {
	Storage *controllers.Storage
}

func NewServer() Server {
	s := Server{
		Storage: controllers.NewStorage(),
	}
	return s
}

func (server Server) SetCookie(token string, w http.ResponseWriter) {
	encoded := hex.EncodeToString([]byte(token))
	cookie := &http.Cookie{Name: "auth_token", Value: encoded, HttpOnly: true}
	http.SetCookie(w, cookie)
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
