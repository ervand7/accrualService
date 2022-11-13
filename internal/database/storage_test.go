package database

import (
	"context"
	"crypto/sha256"
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/config"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"math/rand"
	"testing"
)

func userIDFixture(
	storage *Storage, login, password, token string, t *testing.T,
) (userID string) {
	rows, err := storage.db.Conn.Query(`
		insert into "public"."user" ("login", "password", "token") values
		($1, $2, $3) returning "user"."id"
		`, login, password, token,
	)
	assert.NoError(t, err)
	defer storage.db.CloseRows(rows)

	for rows.Next() {
		err = rows.Scan(&userID)
		assert.NoError(t, err)
	}
	err = rows.Err()
	assert.NoError(t, err)

	return userID
}

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

func TestGetUserIDByToken_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	token := "1"
	userID := userIDFixture(storage, "1", "1", token, t)

	result, err := storage.GetUserIDByToken(ctx, token)
	assert.NoError(t, err)
	assert.Equal(t, userID, result)
}

func TestGetUserIDByToken_FailNotFound(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	token := "hello"
	result, err := storage.GetUserIDByToken(ctx, token)
	assert.NoError(t, err)
	assert.Empty(t, result)
}

func TestCreateOrder_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	userID := userIDFixture(storage, "1", "1", "1", t)
	err := storage.CreateOrder(ctx, number, userID)
	assert.NoError(t, err)

	rows, err := storage.db.Conn.Query(`
		select "id", "user_id", "status", "uploaded_at" 
		from "public"."order" 
		where "user_id" = $1`, userID,
	)
	assert.NoError(t, err)
	defer storage.db.CloseRows(rows)

	var order models.Order
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
	assert.Equal(t, models.OrderStatus.NEW, order.Status)
	assert.NotNil(t, order.UploadedAt)
}

func TestCreateOrder_FailAlreadyCreatedByCurrentUser(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	ctx := context.TODO()
	number := rand.Intn(1000000)
	userID := userIDFixture(storage, "1", "1", "1", t)

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
	userID := userIDFixture(storage, "1", "1", "1", t)
	err := storage.CreateOrder(ctx, number, userID)
	assert.NoError(t, err)

	userID = userIDFixture(storage, "2", "2", "2", t)
	err = storage.CreateOrder(ctx, number, userID)
	errData, ok := err.(*e.OrderAlreadyExistsError)
	assert.True(t, ok)
	assert.False(t, errData.FromCurrentUser)
}

func TestFindOrdersToAccrual_Success(t *testing.T) {
	defer Downgrade()
	storage := NewStorage()

	userID := userIDFixture(storage, "1", "1", "1", t)
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
