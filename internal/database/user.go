package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"time"
)

type orderInfo struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
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

func (storage *Storage) GetUserOrders(
	ctx context.Context, userID string,
) (data []orderInfo, err error) {
	query := `
		select "order"."id", 
			   "order"."status", 
			   "accrual"."amount", 
			   "order"."uploaded_at"::timestamptz 
		from "order" 
				 left outer join "accrual" on "order"."id" = "accrual"."order_id" 
		where "order"."user_id" = $1 
		order by "uploaded_at"; 
	`
	rows, err := storage.db.Conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer storage.db.CloseRows(rows)

	var info orderInfo
	for rows.Next() {
		err = rows.Scan(
			&info.Number,
			&info.Status,
			&info.Accrual,
			&info.UploadedAt,
		)
		if err != nil {
			return nil, err
		}
		data = append(data, info)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}

func (storage *Storage) GetUserByToken(
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

func (storage *Storage) GetUserBalance(
	ctx context.Context, userID string,
) (balance map[string]float64, err error) {
	var accrualSum, withdrawnSum float64
	balance = map[string]float64{
		"accrual":   accrualSum,
		"withdrawn": withdrawnSum,
	}

	for tableName, result := range balance {
		query := storage.db.Conn.QueryRowContext(
			ctx,
			fmt.Sprintf(`select sum("amount") from %s where "user_id" = $1 
			   group by "user_id";`, tableName,
			),
			userID,
		)
		err = query.Scan(&result)
		if err != nil {
			if err == sql.ErrNoRows {
				accrualSum = 0
			} else {
				return nil, err
			}
		}
		balance[tableName] = result
	}

	balance["current"] = balance["accrual"] - balance["withdrawn"]
	delete(balance, "accrual")
	return balance, nil
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
