package database

import (
	"context"
	"math/rand"
	"testing"
	"time"

	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/stretchr/testify/assert"
)

func TestCreateOrder_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	userID := UserIDFixture(storage, "1", "1", "1", t)
	err := storage.CreateOrder(ctx, number, userID)
	assert.NoError(t, err)

	rows, err := storage.db.conn.Query(`
		select "id", "user_id", "status", "uploaded_at" 
		from "public"."order" 
		where "user_id" = $1`, userID,
	)
	assert.NoError(t, err)
	defer storage.db.closeRows(rows)

	var order struct {
		ID         int
		UserID     string
		Status     d.OrderStatusValue
		UploadedAt time.Time
	}
	for rows.Next() {
		err = rows.Scan(
			&order.ID, &order.UserID, &order.Status, &order.UploadedAt,
		)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)

	assert.Equal(t, number, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, d.OrderStatus.NEW, order.Status)
	assert.NotNil(t, order.UploadedAt)
}

func TestCreateOrder_FailAlreadyCreatedByCurrentUser(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	userID := UserIDFixture(storage, "1", "1", "1", t)

	err := storage.CreateOrder(ctx, number, userID)
	assert.NoError(t, err)

	err = storage.CreateOrder(ctx, number, userID)
	errData, ok := err.(*e.OrderAlreadyExistsError)
	assert.True(t, ok)
	assert.True(t, errData.FromCurrentUser)
}

func TestCreateOrder_FailAlreadyCreatedByAnotherUser(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	userID := UserIDFixture(storage, "1", "1", "1", t)
	err := storage.CreateOrder(ctx, number, userID)
	assert.NoError(t, err)

	userID = UserIDFixture(storage, "2", "2", "2", t)
	err = storage.CreateOrder(ctx, number, userID)
	errData, ok := err.(*e.OrderAlreadyExistsError)
	assert.True(t, ok)
	assert.False(t, errData.FromCurrentUser)
}

func TestGetUserOrders_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	amount := 11.1
	for i := 0; i < 10; i++ {
		orderID := rand.Intn(1000000)
		err := storage.CreateOrder(ctx, orderID, userID)
		assert.NoError(t, err)
		if i%2 == 0 {
			query := `insert into "public"."accrual" ("order_id", "user_id", "amount") 
			values ($1, $2, $3);`
			storage.db.conn.QueryRow(query, orderID, userID, amount)
		}
	}

	userOrders, err := storage.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	for index, order := range userOrders {
		if index%2 == 0 {
			assert.Equal(t, amount, *order.Accrual)
		} else {
			assert.Nil(t, order.Accrual)
		}
	}
}

func TestGetUserOrders_SuccessNoOrders(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	result, err := storage.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
