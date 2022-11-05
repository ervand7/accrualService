package main

import (
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/router"
)

func main() {
	logger.Logger.Info("server started ====================")
	router.Run()
}
