package database

import (
	"context"
	"fmt"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
)

func (s *Storage) FindOrdersToAccrual(ctx context.Context) (orders []int, err error) {
	query := fmt.Sprintf(`
			select "id" from "public"."order" where "status" in ('NEW', 'PROCESSING') limit %d
			`, config.OrdersBatchSize,
	)

	rows, err := s.db.conn.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer s.db.closeRows(rows)

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

func (s *Storage) UpdateOrderAndAccrual(
	ctx context.Context, id int, status d.OrderStatusValue, accrual float64,
) error {
	transaction, err := s.db.conn.Begin()
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

	result := orderStmt.QueryRowContext(ctx, status, id)
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
		_, err = accrualStmt.ExecContext(ctx, id, userID, accrual)
		if err != nil {
			return err
		}
	}

	return transaction.Commit()
}
