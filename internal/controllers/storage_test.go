package controllers

import (
	"context"
	"crypto/sha256"
	"fmt"
	e "github.com/ervand7/go-musthave-diploma-tpl/internal/errors"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/models"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
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

	ctx := context.Background()
	err := storage.CreateUser(ctx, login, password, token)
	assert.NoError(t, err)

	rows, err := storage.DB.Conn.Query(`select * from "public"."user"`)
	assert.NoError(t, err)
	defer func() {
		err := rows.Close()
		assert.NoError(t, err)
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
	assert.Equal(t, user.Password, fmt.Sprintf("%x", sha256.Sum256([]byte(password))))
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

	ctx := context.Background()
	err := storage.CreateUser(ctx, login, password, token)
	assert.NoError(t, err)

	err = storage.CreateUser(ctx, login, password, token)
	assert.Error(t, err)
	assert.IsType(t, err, &e.LoginAlreadyExistsError{})
}
