package router

import (
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	"log"
	"net/http"
)

func Run() {
	router := newRouter()
	log.Fatal(
		http.ListenAndServe(
			config.GetConfig().RunAddress,
			router,
		),
	)
}
