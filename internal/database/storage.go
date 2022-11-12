package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
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

func (storage *Storage) FindOrdersToAccrual(lastID interface{}) (orders []int, err error) {
	var lastIDCondition string
	if lastID != nil {
		lastIDCondition = fmt.Sprintf(` and "id" > %d `, lastID)
	} else {
		lastIDCondition = ""
	}
	query := fmt.Sprintf(`
			select "id" from "public"."order" where "status" in ('NEW', 'PROCESSING') 
			%s order by "id" limit %d
			`, lastIDCondition, config.OrdersBatchSize)

	rows, err := storage.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer storage.db.CloseRows(rows)

	for rows.Next() {
		var orderID int
		err = rows.Scan(&orderID)
		if err != nil {
			return nil, err
		}
		orders = append(orders, orderID)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}
	return orders, nil
}

func (storage *Storage) UpdateOrderAndAccrual(
	id int, status models.OrderStatusValue, accrual float64,
) error {
	transaction, err := storage.db.Conn.Begin()
	if err != nil {
		return err
	}
	defer transaction.Rollback()

	orderQuery := `
		update "public"."order" set "status" = $1 where "id" = $2 
		returning "user_id";
	`
	orderStmt, err := transaction.Prepare(orderQuery)
	if err != nil {
		return err
	}
	defer orderStmt.Close()

	result := orderStmt.QueryRow(status, id)
	var userID string
	err = result.Scan(&userID)
	if err != nil {
		return err
	}

	if accrual > 0 {
		accrualQuery := `
			insert into "public"."accrual" ("order_id", "user_id", "amount") 
			values ($1, $2, $3);
		`
		accrualStmt, err := transaction.Prepare(accrualQuery)
		if err != nil {
			return err
		}
		defer accrualStmt.Close()
		_, err = accrualStmt.Exec(id, userID, accrual)
		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}
