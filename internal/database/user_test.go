package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestCreateUser_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	login := "1"
	password := "1"
	token := uuid.New().String()
	err := storage.CreateUser(context.TODO(), login, password, token)
	assert.NoError(t, err)

	rows, err := storage.db.Conn.Query(`select * from "public"."user"`)
	assert.NoError(t, err)
	defer storage.db.CloseRows(rows)

	var user struct {
		ID       string
		Login    string
		Password string
		Token    string
	}
	for rows.Next() {
		err = rows.Scan(&user.ID, &user.Login, &user.Password, &user.Token)
		assert.NoError(t, err)
	}

	err = rows.Err()
	assert.NoError(t, err)
	assert.NotNil(t, user.ID)
	assert.Equal(t, user.Login, login)
	assert.Equal(
		t, user.Password,
		fmt.Sprintf("%x", sha256.Sum256([]byte(password))),
	)
	assert.Equal(t, user.Token, token)
}

func TestCreateUser_FailLoginAlreadyExists(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	login := "1"
	password := "1"
	token := uuid.New().String()
	ctx := context.TODO()
	err := storage.CreateUser(ctx, login, password, token)
	assert.NoError(t, err)

	err = storage.CreateUser(ctx, login, password, token)
	assert.Error(t, err)
	assert.IsType(t, err, &e.LoginAlreadyExistsError{})
}

func TestGetUserOrders_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	accrual := 11.1
	for i := 0; i < 10; i++ {
		orderID := rand.Intn(1000000)
		err := storage.CreateOrder(ctx, orderID, userID)
		assert.NoError(t, err)
		if i%2 == 0 {
			query := `insert into "public"."accrual" ("order_id", "user_id", "amount") 
			values ($1, $2, $3);`
			storage.db.Conn.QueryRow(query, orderID, userID, accrual)
		}
	}

	userOrders, err := storage.GetUserOrders(ctx, userID)
	assert.NoError(t, err)
	for index, order := range userOrders {
		if index%2 == 0 {
			assert.Equal(t, accrual, *order.Accrual)
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

func TestGetUserByToken_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	token := "1"
	userID := UserIDFixture(storage, "1", "1", token, t)

	result, err := storage.GetUserByToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, userID, result)
}

func TestGetUserByToken_FailNotFound(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	token := "hello"
	result, err := storage.GetUserByToken(ctx, token)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestGetUserBalance_Success(t *testing.T) {
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
			storage.db.Conn.QueryRow(query, orderID, userID, amount)
		} else {
			query := `insert into "public"."withdrawn" ("order_id", "user_id", "amount") 
			values ($1, $2, $3);`
			storage.db.Conn.QueryRow(query, orderID, userID, amount)
		}
	}

	result, err := storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]float64{
			"current":   0,
			"withdrawn": 55.5,
		},
		result,
	)
}

func TestGetUserBalance_SuccessNoOrders(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := UserIDFixture(storage, "1", "1", "1", t)
	ctx := context.TODO()
	result, err := storage.GetUserBalance(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(
		t,
		map[string]float64{
			"current":   0,
			"withdrawn": 0,
		},
		result,
	)
}

func TestUpdateToken_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	login := "1"
	password := "1"
	oldToken := uuid.New().String()
	newToken := uuid.New().String()
	assert.NotEqual(t, oldToken, newToken)

	ctx := context.TODO()
	err := storage.CreateUser(ctx, login, password, oldToken)
	assert.NoError(t, err)

	err = storage.UpdateToken(ctx, login, password, newToken)
	assert.NoError(t, err)

	rows, err := storage.db.Conn.Query(`
		select "token" from "public"."user" where "login" = $1`,
		login,
	)
	assert.NoError(t, err)
	defer storage.db.CloseRows(rows)

	var resultToken string
	for rows.Next() {
		err = rows.Scan(&resultToken)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)
	assert.Equal(t, resultToken, newToken)
}

func TestUpdateToken_FailUserNotFound(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	login := "1"
	password := "1"
	token := uuid.New().String()
	err := storage.UpdateToken(context.TODO(), login, password, token)
	assert.Error(t, err)

	rows, err := storage.db.Conn.Query(`
		select "token" from "public"."user" where "login" = $1`,
		login,
	)
	assert.NoError(t, err)
	defer storage.db.CloseRows(rows)

	var resultToken string
	for rows.Next() {
		err = rows.Scan(&resultToken)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)
	assert.Equal(t, resultToken, "")
}
