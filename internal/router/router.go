package router

import (
	"github.com/ervand7/go-musthave-diploma-tpl/internal/views"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newRouter() chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(GzipMiddleware)

	server := views.NewServer()
	r.Route("/", func(r chi.Router) {
		r.Post("/api/user/register", server.UserRegister)
		r.Post("/api/user/login", server.UserLogin)
		r.Post("/api/user/orders", server.UserLoadOrders)
	})

	return r
}
