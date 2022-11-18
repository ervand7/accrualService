package utils

import (
	"context"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"strconv"
	"testing"
	"time"
)

func TestStartWorker_Success(t *testing.T) {
	defer database.Downgrade()
	storage := database.NewStorage()

	ctx := context.TODO()
	userID := database.UserIDFixture(storage, "1", "1", "1", t)
	orderID := rand.Intn(1000000)
	accrual := rand.Float64()
	err := storage.CreateOrder(ctx, orderID, userID)
	assert.NoError(t, err)

	requestAccrualServer = func(orderID int) respBody {
		return respBody{
			Order:   strconv.Itoa(orderID),
			Status:  "PROCESSED",
			Accrual: accrual,
		}
	}
	StartWorker(storage)
	time.Sleep(2 * waitDuration)
	balance, err := storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, accrual, balance["current"])
}
