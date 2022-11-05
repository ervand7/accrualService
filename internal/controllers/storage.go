package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
)

type Storage struct {
	DB database.Database
}

func NewStorage() *Storage {
	db := database.Database{}
	db.Run()
	return &Storage{
		DB: db,
	}
}

func (storage *Storage) CreateUser(
	ctx context.Context, login, password, token string,
) error {
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	query := `
		insert into "public"."user" ("login", "password", "token")
		values ($1, $2, $3)
		on conflict ("login") do nothing
		returning "token";
`
	var setToken string
	rows, err := storage.DB.Conn.QueryContext(ctx, query, login, hashedPassword, token)
	if err != nil {
		return err
	}
	defer storage.DB.CloseRows(rows)

	for rows.Next() {
		err = rows.Scan(&setToken)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	if setToken == "" {
		return e.NewLoginAlreadyExistsError(login)
	}

	return nil
}
