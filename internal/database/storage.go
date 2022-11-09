package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/models"
)

type Storage struct {
	db database
}

func NewStorage() *Storage {
	db := database{}
	db.Run()
	return &Storage{
		db: db,
	}
}

func getValueFromRows(
	storage *Storage, rows *sql.Rows,
) (result string, err error) {
	defer storage.db.CloseRows(rows)

	for rows.Next() {
		err = rows.Scan(&result)
		if err != nil {
			return "", err
		}
	}
	err = rows.Err()
	if err != nil {
		return "", err
	}

	return result, nil
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
	rows, err := storage.db.Conn.QueryContext(ctx, query, login, hashedPassword, token)
	if err != nil {
		return err
	}
	setToken, err := getValueFromRows(storage, rows)
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
	rows, err := storage.db.Conn.QueryContext(
		ctx, query, login, hashedPassword, newToken,
	)
	if err != nil {
		return err
	}
	setToken, err := getValueFromRows(storage, rows)
	if err != nil {
		return err
	}
	if setToken == "" {
		return e.NewUserNotFoundError(login, password)
	}

	return nil
}

func (storage *Storage) CreateOrder(
	ctx context.Context, orderNumber int, userID string,
) error {
	query := `
		with cte as (
			insert into "public"."order" ("id", "user_id", "status")
				values ($1, $2, $3)
				on conflict ("id") do nothing
				returning "user_id")
		select null as result
		where exists(select 1 from cte)
		union all
		select "user_id"
		from "public"."order"
		where "id" = $1
		  and not exists(select 1 from cte);
	`
	rows, err := storage.db.Conn.QueryContext(
		ctx, query, orderNumber, userID, models.OrderStatus.NEW,
	)
	if err != nil {
		return err
	}
	defer storage.db.CloseRows(rows)

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
	rows, err := storage.db.Conn.QueryContext(ctx, query, token)
	if err != nil {
		return "", err
	}
	userID, err = getValueFromRows(storage, rows)
	if err != nil {
		return "", err
	}

	return userID, nil
}
