package database

import (
	"context"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestCreateWithdraw_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	orderID := rand.Intn(1000000)
	accrual := rand.Float64()
	err := storage.CreateOrder(ctx, orderID, userID)
	assert.NoError(t, err)
	err = storage.UpdateOrderAndAccrual(ctx, orderID, d.OrderStatus.PROCESSED, accrual)
	assert.NoError(t, err)
	balance, err := storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]float64{
			"current":   accrual,
			"withdrawn": 0,
		},
		balance,
	)

	err = storage.CreateWithdraw(ctx, userID, orderID, accrual)
	assert.NoError(t, err)
	balance, err = storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]float64{
			"current":   0,
			"withdrawn": accrual,
		},
		balance,
	)
}

func TestCreateWithdraw_FailNotEnoughMoney(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	orderID := rand.Intn(1000000)
	accrual := rand.Float64()
	err := storage.CreateOrder(ctx, orderID, userID)
	assert.NoError(t, err)
	err = storage.UpdateOrderAndAccrual(ctx, orderID, d.OrderStatus.PROCESSED, accrual)
	assert.NoError(t, err)
	balanceBefore, err := storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]float64{
			"current":   accrual,
			"withdrawn": 0,
		},
		balanceBefore,
	)

	err = storage.CreateWithdraw(ctx, userID, orderID, accrual+1)
	assert.Error(t, err)
	assert.IsType(t, err, &e.NotEnoughMoneyError{})
	balanceAfter, err := storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, balanceBefore, balanceAfter)
}

func TestGetUserWithdrawals_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	amount := 1.0
	for i := 0; i < 10; i++ {
		orderID := rand.Intn(1000000)
		query := `insert into "public"."withdrawn" ("order_id", "user_id", "amount") 
			values ($1, $2, $3);`
		storage.db.conn.QueryRow(query, orderID, userID, amount)
	}

	ctx := context.TODO()
	userWithdrawals, err := storage.GetUserWithdrawals(ctx, userID)
	assert.NoError(t, err)

	sum := 0.0
	for _, value := range userWithdrawals {
		sum += *value.Sum
		assert.NotNil(t, value.Order)
		assert.NotNil(t, value.ProcessedAt)
	}
	assert.Equal(t, amount*10.0, sum)
}

func TestGetUserWithdrawals_SuccessNoOrders(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	result, err := storage.GetUserWithdrawals(ctx, userID)
	assert.NoError(t, err)
	assert.Nil(t, result)
}
