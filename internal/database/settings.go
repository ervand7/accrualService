package database

import (
	"database/sql"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"syscall"
	"time"
)

type database struct {
	Conn *sql.DB
}

func (db *database) Run() {
	err := db.ConnStart()
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	db.setConnPool()
	err = db.migrate()
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}

	ch := make(chan os.Signal, 3)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	go func() {
		<-ch
		signal.Stop(ch)
		err = db.ConnClose()
		if err != nil {
			logger.Logger.Error(err.Error())
			os.Exit(1)
		}
		logger.Logger.Info("DB Connection was closed successfully")
		os.Exit(0)
	}()
}

func (db *database) ConnStart() (err error) {
	conn, err := goose.OpenDBWithDriver("pgx", config.GetConfig().DatabaseURI)
	if err != nil {
		return err
	}
	db.Conn = conn
	return nil
}

func (db *database) ConnClose() (err error) {
	err = db.Conn.Close()
	if err != nil {
		return err
	}
	return nil
}

func (db *database) Ping() (err error) {
	err = db.Conn.Ping()
	if err != nil {
		return err
	}
	return nil
}

func (db *database) CloseRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logger.Logger.Error(err.Error())
	}
}

func (db *database) setConnPool() {
	db.Conn.SetMaxOpenConns(20)
	db.Conn.SetMaxIdleConns(20)
	db.Conn.SetConnMaxIdleTime(time.Second * 30)
	db.Conn.SetConnMaxLifetime(time.Minute * 2)
}

func (db *database) migrate() (err error) {
	if err = goose.Run("up", db.Conn, getMigrationsDir()); err != nil {
		return err
	}
	return nil
}

func getMigrationsDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	migrationsDir := filepath.Join(currentDir, "/../../migrations")
	return migrationsDir
}
