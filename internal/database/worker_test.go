package database

import (
	"context"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestFindOrdersToAccrual_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	for i := 0; i < 10; i++ {
		err := storage.CreateOrder(ctx, rand.Intn(1000000), userID)
		assert.NoError(t, err)
	}

	result, err := storage.FindOrdersToAccrual(nil)
	assert.NoError(t, err)
	assert.Equal(t, config.OrdersBatchSize, len(result))

	lastID := result[len(result)-1]
	result, err = storage.FindOrdersToAccrual(lastID)
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
		storage.db.Conn.QueryRow(query, rand.Intn(1000000), userID, status)
	}
	ctx := context.TODO()
	orders, err := storage.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, len(statuses), len(orders))

	result, err := storage.FindOrdersToAccrual(nil)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
