package database

import (
	"context"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/enum"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
	"time"
)

func TestCreateOrder_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	userID := UserIDFixture(storage, "1", "1", "1", t)
	err := storage.CreateOrder(ctx, number, userID)
	assert.NoError(t, err)

	rows, err := storage.db.Conn.Query(`
		select "id", "user_id", "status", "uploaded_at" 
		from "public"."order" 
		where "user_id" = $1`, userID,
	)
	assert.NoError(t, err)
	defer storage.db.CloseRows(rows)

	var order struct {
		ID         int
		UserID     string
		Status     enum.OrderStatusValue
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
	assert.Equal(t, enum.OrderStatus.NEW, order.Status)
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
