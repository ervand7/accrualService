package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/models"
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

func (storage *Storage) UpdateToken(
	ctx context.Context, login, password, newToken string,
) error {
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	query := `
		update "public"."user" set "token" = $3
		where "login" = $1 and "password" = $2
		returning "token";
	`
	var setToken string
	rows, err := storage.DB.Conn.QueryContext(ctx, query, login, hashedPassword, newToken)
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
		return e.NewUserNotFoundError(login, password)
	}

	return nil
}

func (storage *Storage) CreateOrder(
	ctx context.Context, userID string, orderNumber int,
) error {
	query := `
		with cte as (
			insert into "public"."order" ("user_id", "number", "status")
				values ($1, $2, $3)
				on conflict ("number") do nothing
				returning "user_id")
		select null as result
		where exists(select 1 from cte)
		union all
		select "user_id"
		from "public"."order"
		where "number" = $2
		  and not exists(select 1 from cte);
	`
	rows, err := storage.DB.Conn.QueryContext(
		ctx, query, userID, orderNumber, models.OrderStatus.NEW,
	)
	if err != nil {
		return err
	}
	defer storage.DB.CloseRows(rows)

	var UserIDFromException interface{}
	for rows.Next() {
		err = rows.Scan(&UserIDFromException)
		if err != nil {
			return err
		}
	}
	err = rows.Err()
	if err != nil {
		return err
	}
	if UserIDFromException != nil {
		switch {
		case UserIDFromException == userID:
			return e.NewOrderAlreadyExistsError(userID, true)
		default:
			return e.NewOrderAlreadyExistsError(userID, false)
		}
	}

	return nil
}

func (storage *Storage) GetUserIDByToken(
	ctx context.Context, token string,
) (userID string, err error) {
	query := `
		select "id" from "public"."user"
		where "token" = $1
	`
	rows, err := storage.DB.Conn.QueryContext(ctx, query, token)
	if err != nil {
		return "", err
	}
	defer storage.DB.CloseRows(rows)

	for rows.Next() {
		err = rows.Scan(&userID)
		if err != nil {
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}

	return userID, nil
}
