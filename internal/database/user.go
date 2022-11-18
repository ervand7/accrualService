package database

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
)

func getValueFromRows(
	s *Storage, rows *sql.Rows,
) (result string, err error) {
	defer s.db.CloseRows(rows)

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

func (s *Storage) CreateUser(
	ctx context.Context, login, password, token string,
) error {
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	query := `
		insert into "public"."user" ("login", "password", "token")
		values ($1, $2, $3)
		on conflict ("login") do nothing
		returning "token";
	`
	rows, err := s.db.Conn.QueryContext(ctx, query, login, hashedPassword, token)
	if err != nil {
		return err
	}
	setToken, err := getValueFromRows(s, rows)
	if err != nil {
		return err
	}
	if setToken == "" {
		return e.NewLoginAlreadyExistsError(login)
	}

	return nil
}

func (s *Storage) GetUserByToken(
	ctx context.Context, token string,
) (userID string, err error) {
	query := `
		select "id" from "public"."user"
		where "token" = $1
	`
	rows, err := s.db.Conn.QueryContext(ctx, query, token)
	if err != nil {
		return "", err
	}
	userID, err = getValueFromRows(s, rows)
	if err != nil {
		return "", err
	}

	return userID, nil
}

func (s *Storage) GetUserBalance(
	ctx context.Context, userID string,
) (balance map[string]float64, err error) {
	var accrualSum, withdrawnSum float64
	balance = map[string]float64{
		"accrual":   accrualSum,
		"withdrawn": withdrawnSum,
	}

	for tableName, result := range balance {
		query := s.db.Conn.QueryRowContext(
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

func (s *Storage) UpdateToken(
	ctx context.Context, login, password, newToken string,
) error {
	hashedPassword := fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
	query := `
		update "public"."user" set "token" = $3
		where "login" = $1 and "password" = $2
		returning "token";
	`
	rows, err := s.db.Conn.QueryContext(
		ctx, query, login, hashedPassword, newToken,
	)
	if err != nil {
		return err
	}
	setToken, err := getValueFromRows(s, rows)
	if err != nil {
		return err
	}
	if setToken == "" {
		return e.NewUserNotFoundError(login, password)
	}

	return nil
}
