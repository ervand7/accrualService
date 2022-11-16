package database

import (
	"context"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
)

func (storage *Storage) CreateWithdraw(
	ctx context.Context, userID string, orderID int, withdrawn float64,
) error {
	balance, err := storage.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}
	if balance["current"]-withdrawn < 0 {
		return e.NewNotEnoughMoneyError(balance["current"])
	}

	query := `insert into "public"."withdrawn" ("order_id", "user_id", "amount") 
	values ($1, $2, $3);`
	_, err = storage.db.Conn.ExecContext(ctx, query, orderID, userID, withdrawn)
	if err != nil {
		return err
	}

	return nil
}
