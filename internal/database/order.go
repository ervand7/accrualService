package database

import (
	"context"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/enum"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
)

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
		ctx, query, orderNumber, userID, enum.OrderStatus.NEW,
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
