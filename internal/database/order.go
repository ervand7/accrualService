package database

import (
	"context"

	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
)

func (s *Storage) CreateOrder(
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
	rows, err := s.db.conn.QueryContext(
		ctx, query, orderNumber, userID, d.OrderStatus.NEW,
	)
	if err != nil {
		return err
	}
	defer s.db.closeRows(rows)

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

func (s *Storage) GetUserOrders(
	ctx context.Context, userID string,
) (data []d.Order, err error) {
	query := `
		select "order"."id", "order"."status", "accrual"."amount", 
			   "order"."uploaded_at"::timestamptz 
		from "order" 
				 left outer join "accrual" on "order"."id" = "accrual"."order_id" 
		where "order"."user_id" = $1 
		order by "uploaded_at"; 
	`
	rows, err := s.db.conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer s.db.closeRows(rows)

	var o d.Order
	for rows.Next() {
		err = rows.Scan(&o.Number, &o.Status, &o.Accrual, &o.UploadedAt)
		if err != nil {
			return nil, err
		}
		data = append(data, o)
	}
	err = rows.Err()
	if err != nil {
		return nil, err
	}

	return data, nil
}
