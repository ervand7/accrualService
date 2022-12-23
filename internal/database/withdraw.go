package database

import (
	"context"
	"time"

	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
)

type withdrawal struct {
	Order       string    `json:"order"`
	Sum         *float64  `json:"sum"`
	ProcessedAt time.Time `json:"processed_at"`
}

func (s *Storage) CreateWithdraw(
	ctx context.Context, userID string, orderID int, withdrawn float64,
) error {
	balance, err := s.GetUserBalance(ctx, userID)
	if err != nil {
		return err
	}
	if balance["current"]-withdrawn < 0 {
		return e.NewNotEnoughMoneyError(balance["current"])
	}

	query := `insert into "public"."withdrawn" ("order_id", "user_id", "amount") 
	values ($1, $2, $3);`
	_, err = s.db.conn.ExecContext(ctx, query, orderID, userID, withdrawn)
	if err != nil {
		return err
	}

	return nil
}

func (s *Storage) GetUserWithdrawals(
	ctx context.Context, userID string,
) (data []withdrawal, err error) {
	query := `
		select "order_id", "amount", "processed_at"::timestamptz from withdrawn 
		where "user_id" = $1 
		order by "processed_at"; 
	`
	rows, err := s.db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer s.db.closeRows(rows)

	var w withdrawal
	for rows.Next() {
		err = rows.Scan(&w.Order, &w.Sum, &w.ProcessedAt)
		if err != nil {
			return nil, err
		}
		data = append(data, w)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}
