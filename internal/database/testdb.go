package database

import (
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	"github.com/pressly/goose/v3"
)

func Downgrade() {
	db := database{}
	db.Run()
	if err := goose.Run("down", db.Conn, getMigrationsDir()); err != nil {
		logger.Logger.Error(err.Error())
	}
}
