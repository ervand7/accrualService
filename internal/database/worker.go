package database

import (
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
)

func (s *Storage) FindOrdersToAccrual(
	lastID interface{},
) (orders []int, err error) {
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

	rows, err := s.db.Conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer s.db.CloseRows(rows)

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
	id int, status OrderStatusValue, accrual float64,
) error {
	transaction, err := s.db.Conn.Begin()
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
