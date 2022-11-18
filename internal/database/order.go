package database

import (
	"context"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"time"
)

type OrderStatusValue string

var OrderStatus = struct {
	NEW        OrderStatusValue
	PROCESSING OrderStatusValue
	INVALID    OrderStatusValue
	PROCESSED  OrderStatusValue
}{
	NEW:        "NEW",
	PROCESSING: "PROCESSING",
	INVALID:    "INVALID",
	PROCESSED:  "PROCESSED",
}

func (o OrderStatusValue) FromEnum() OrderStatusValue {
	if o != OrderStatus.NEW &&
		o != OrderStatus.PROCESSING &&
		o != OrderStatus.PROCESSED &&
		o != OrderStatus.INVALID {
		return ""
	}
	return o
}

type orderInfo struct {
	Number     string    `json:"number"`
	Status     string    `json:"status"`
	Accrual    *float64  `json:"accrual,omitempty"`
	UploadedAt time.Time `json:"uploaded_at"`
}

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
	rows, err := s.db.Conn.QueryContext(
		ctx, query, orderNumber, userID, OrderStatus.NEW,
	)
	if err != nil {
		return err
	}
	defer s.db.CloseRows(rows)

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
) (data []orderInfo, err error) {
	query := `
		select "order"."id", "order"."status", "accrual"."amount", 
			   "order"."uploaded_at"::timestamptz 
		from "order" 
				 left outer join "accrual" on "order"."id" = "accrual"."order_id" 
		where "order"."user_id" = $1 
		order by "uploaded_at"; 
	`
	rows, err := s.db.Conn.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer s.db.CloseRows(rows)

	var o orderInfo
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
