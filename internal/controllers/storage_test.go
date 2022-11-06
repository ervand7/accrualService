package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func TestCreateUser_Success(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	login := "hello"
	password := "world"
	token := uuid.New().String()
	err := storage.CreateUser(context.TODO(), login, password, token)
	assert.NoError(t, err)

	rows, err := storage.DB.Conn.Query(`select * from "public"."user"`)
	assert.NoError(t, err)
	defer func() {
		storage.DB.CloseRows(rows)
	}()

	var user models.User
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
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	login := "hello"
	password := "world"
	token := uuid.New().String()
	ctx := context.TODO()
	err := storage.CreateUser(ctx, login, password, token)
	assert.NoError(t, err)

	err = storage.CreateUser(ctx, login, password, token)
	assert.Error(t, err)
	assert.IsType(t, err, &e.LoginAlreadyExistsError{})
}

func TestUpdateToken_Success(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	login := "hello"
	password := "world"
	oldToken := uuid.New().String()
	newToken := uuid.New().String()
	assert.NotEqual(t, oldToken, newToken)

	ctx := context.TODO()
	err := storage.CreateUser(ctx, login, password, oldToken)
	assert.NoError(t, err)

	err = storage.UpdateToken(ctx, login, password, newToken)
	assert.NoError(t, err)

	rows, err := storage.DB.Conn.Query(`
		select "token" from "public"."user" where "login" = $1`,
		login,
	)
	assert.NoError(t, err)
	defer func() {
		storage.DB.CloseRows(rows)
	}()

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
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	login := "hello"
	password := "world"
	token := uuid.New().String()
	err := storage.UpdateToken(context.TODO(), login, password, token)
	assert.Error(t, err)

	rows, err := storage.DB.Conn.Query(`
		select "token" from "public"."user" where "login" = $1`,
		login,
	)
	assert.NoError(t, err)
	defer func() {
		storage.DB.CloseRows(rows)
	}()

	var resultToken string
	for rows.Next() {
		err = rows.Scan(&resultToken)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)
	assert.Equal(t, resultToken, "")
}

func TestGetUserIDByToken_Success(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	ctx := context.TODO()
	login := "hello"
	password := "world"
	token := "foobar"
	rows, err := storage.DB.Conn.Query(`
		insert into "public"."user" ("login", "password", "token") values
		($1, $2, $3) returning "user"."id"
		`, login, password, token,
	)
	assert.NoError(t, err)
	defer func() {
		storage.DB.CloseRows(rows)
	}()
	var userID string
	for rows.Next() {
		err = rows.Scan(&userID)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)

	result, err := storage.GetUserIDByToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, userID, result)
}

func TestGetUserIDByToken_FailNotFound(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	ctx := context.TODO()
	token := "hello"
	result, err := storage.GetUserIDByToken(ctx, token)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCreateOrder_Success(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	ctx := context.TODO()
	userID := uuid.New().String()
	number := rand.Intn(1000000)
	err := storage.CreateOrder(ctx, userID, number)
	assert.NoError(t, err)

	rows, err := storage.DB.Conn.Query(`
		select "id", "user_id", "number", "status", "uploaded_at" 
		from "public"."order" 
		where "user_id" = $1`, userID,
	)
	assert.NoError(t, err)
	defer func() {
		storage.DB.CloseRows(rows)
	}()
	var order models.Order
	for rows.Next() {
		err = rows.Scan(
			&order.ID, &order.UserID, &order.Number, &order.Status, &order.UploadedAt,
		)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)

	assert.NotNil(t, order.ID)
	assert.Equal(t, userID, order.UserID)
	assert.Equal(t, number, order.Number)
	assert.Equal(t, models.OrderStatus.NEW, order.Status)
	assert.NotNil(t, order.UploadedAt)
}

func TestCreateOrder_FailAlreadyCreatedByCurrentUser(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	ctx := context.TODO()
	userID := uuid.New().String()
	number := rand.Intn(1000000)

	err := storage.CreateOrder(ctx, userID, number)
	assert.NoError(t, err)

	err = storage.CreateOrder(ctx, userID, number)
	errData, ok := err.(*e.OrderAlreadyExistsError)
	assert.True(t, ok)
	assert.True(t, errData.FromCurrentUser)
}

func TestCreateOrder_FailAlreadyCreatedByAnotherUser(t *testing.T) {
	storage := NewStorage()
	defer func() {
		storage.DB.Downgrade()
	}()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	err := storage.CreateOrder(ctx, uuid.New().String(), number)
	assert.NoError(t, err)

	err = storage.CreateOrder(ctx, uuid.New().String(), number)
	errData, ok := err.(*e.OrderAlreadyExistsError)
	assert.True(t, ok)
	assert.False(t, errData.FromCurrentUser)
}
