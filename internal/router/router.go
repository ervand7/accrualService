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
		r.Post("/api/user/register", server.Register)
		r.Post("/api/user/login", server.Login)
		r.Post("/api/user/orders", server.LoadOrder)
		r.Get("/api/user/orders", server.GetUserOrders)
		r.Get("/api/user/balance", server.UserBalance)
	})

	return r
}
