package router

import (
	"log"
	"net/http"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
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
