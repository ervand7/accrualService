package database

import (
	"database/sql"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/logger"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/pressly/goose/v3"
	"path/filepath"
	"runtime"
	"time"
)

const (
	maxOpenConnections    = 20
	maxIdleConnections    = 20
	connMaxIdleTimeSecond = 30
	connMaxLifetimeSecond = 2
)

type Storage struct {
	db database
}

func NewStorage() *Storage {
	db := database{}
	db.run()
	return &Storage{
		db: db,
	}
}

type database struct {
	conn *sql.DB
}

func (db *database) run() {
	err := db.connStart()
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
	db.setConnPool()
	err = db.migrate()
	if err != nil {
		logger.Logger.Fatal(err.Error())
	}
}

func (db *database) connStart() (err error) {
	conn, err := goose.OpenDBWithDriver("pgx", config.GetConfig().DatabaseURI)
	if err != nil {
		return err
	}
	db.conn = conn
	return nil
}

func (db *database) setConnPool() {
	db.conn.SetMaxOpenConns(maxOpenConnections)
	db.conn.SetMaxIdleConns(maxIdleConnections)
	db.conn.SetConnMaxIdleTime(time.Second * connMaxIdleTimeSecond)
	db.conn.SetConnMaxLifetime(time.Minute * connMaxLifetimeSecond)
}

func (db *database) migrate() (err error) {
	if err = goose.Run("up", db.conn, getMigrationsDir()); err != nil {
		return err
	}
	return nil
}

func (db *database) closeRows(rows *sql.Rows) {
	if err := rows.Close(); err != nil {
		logger.Logger.Error(err.Error())
	}
}

func getMigrationsDir() string {
	_, currentFile, _, _ := runtime.Caller(0)
	currentDir := filepath.Dir(currentFile)
	migrationsDir := filepath.Join(currentDir, "/../../migrations")
	return migrationsDir
}
