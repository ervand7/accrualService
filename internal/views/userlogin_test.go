package views

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserLogin_200Success(t *testing.T) {
	apiMethod := "/api/user/login"
	login := "hello"
	password := "world"
	var body = []byte(fmt.Sprintf(`{
		"login": "%s",
		"password": "%s"
	}`, login, password))
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)
	oldToken := uuid.New().String()
	encoded := hex.EncodeToString([]byte(oldToken))
	cookie := &http.Cookie{Name: "auth_token", Value: encoded, HttpOnly: true}
	request.AddCookie(cookie)

	server := NewServer()
	defer func() {
		server.Storage.DB.Downgrade()
	}()
	err := server.Storage.CreateUser(
		context.TODO(), login, password, oldToken,
	)
	assert.NoError(t, err)

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.UserLogin)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	oldCookie := request.Header.Get("Cookie")
	newCookie := response.Header.Get("Set-Cookie")
	assert.NotEqual(t, oldCookie, newCookie)

	err = response.Body.Close()
	require.NoError(t, err)
}

func TestUserLogin_400BadRequest(t *testing.T) {
	apiMethod := "/api/user/login"
	var body = []byte("")
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)

	server := NewServer()
	defer func() {
		server.Storage.DB.Downgrade()
	}()

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.UserLogin)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	setCookie := response.Header.Get("Set-Cookie")
	assert.Empty(t, setCookie)

	err := response.Body.Close()
	require.NoError(t, err)
}

func TestUserLogin_401WrongLoginOrPassword(t *testing.T) {
	apiMethod := "/api/user/login"
	login := "hello"
	password := "world"
	var body = []byte(fmt.Sprintf(`{
		"login": "%s",
		"password": "%s"
	}`, login, password))
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)

	server := NewServer()
	defer func() {
		server.Storage.DB.Downgrade()
	}()

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.UserLogin)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	body, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.Equal(
		t,
		fmt.Sprintf("user not found with this credentials: %s %s", login, password),
		string(body),
	)
	setCookie := response.Header.Get("Set-Cookie")
	assert.Empty(t, setCookie)

	err = response.Body.Close()
	require.NoError(t, err)
}