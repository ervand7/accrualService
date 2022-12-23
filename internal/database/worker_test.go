package database

import (
	"context"
	"math/rand"
	"testing"
	"time"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	"github.com/stretchr/testify/assert"
)

func TestFindOrdersToAccrual_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	for i := 0; i < 10; i++ {
		err := storage.CreateOrder(ctx, rand.Intn(1000000), userID)
		assert.NoError(t, err)
	}

	result, err := storage.FindOrdersToAccrual(ctx)
	assert.NoError(t, err)
	assert.Equal(t, config.OrdersBatchSize, len(result))

	result, err = storage.FindOrdersToAccrual(ctx)
	assert.NoError(t, err)
	assert.Equal(t, config.OrdersBatchSize, len(result))
}

func TestFindOrdersToAccrual_FailNoOrders(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	statuses := []d.OrderStatusValue{
		d.OrderStatus.PROCESSED, d.OrderStatus.INVALID,
	}
	for _, status := range statuses {
		query := `
			insert into "public"."order" ("id", "user_id", "status") values ($1, $2, $3);
		`
		storage.db.conn.QueryRow(query, rand.Intn(1000000), userID, status)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	orders, err := storage.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, len(statuses), len(orders))

	result, err := storage.FindOrdersToAccrual(ctx)
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestUpdateOrderAndAccrual_SuccessAccrual(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	userID := UserIDFixture(storage, "1", "1", "1", t)
	orderID := rand.Intn(1000000)
	err := storage.CreateOrder(ctx, orderID, userID)
	assert.NoError(t, err)

	amount := rand.Float64()
	err = storage.UpdateOrderAndAccrual(ctx, orderID, d.OrderStatus.PROCESSED, amount)
	assert.NoError(t, err)

	row := storage.db.conn.QueryRow(`
		select "status" from "public"."order" 
		where "id" = $1`, orderID,
	)
	var status d.OrderStatusValue
	err = row.Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, d.OrderStatus.PROCESSED, status)

	row = storage.db.conn.QueryRow(`
		select "amount" from "public"."accrual" 
		where "order_id" = $1`, orderID,
	)
	var accrual float64
	err = row.Scan(&accrual)
	assert.NoError(t, err)
	assert.Equal(t, amount, accrual)
}

func TestUpdateOrderAndAccrual_SuccessNotAccrual(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	userID := UserIDFixture(storage, "1", "1", "1", t)
	orderID := rand.Intn(1000000)
	err := storage.CreateOrder(ctx, orderID, userID)
	assert.NoError(t, err)

	amount := 0.0
	err = storage.UpdateOrderAndAccrual(ctx, orderID, d.OrderStatus.INVALID, amount)
	assert.NoError(t, err)

	row := storage.db.conn.QueryRow(`
		select "status" from "public"."order" 
		where "id" = $1`, orderID,
	)
	var status d.OrderStatusValue
	err = row.Scan(&status)
	assert.NoError(t, err)
	assert.Equal(t, d.OrderStatus.INVALID, status)

	row = storage.db.conn.QueryRow(`select exists(select * from "public"."accrual")`)
	var exists bool
	err = row.Scan(&exists)
	assert.NoError(t, err)
	assert.False(t, exists)
}
