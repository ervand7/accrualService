package database

import (
	"context"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/enum"
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
	err = storage.UpdateOrderAndAccrual(orderID, enum.OrderStatus.PROCESSED, accrual)
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
	err = storage.UpdateOrderAndAccrual(orderID, enum.OrderStatus.PROCESSED, accrual)
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
