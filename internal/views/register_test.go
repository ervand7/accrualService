package views

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_200Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/register"
	var body = []byte(`{
		"login": "hello",
		"password": "world"
	}`)
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)

	server := NewServer()
	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.Register)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	setCookie := response.Header.Get("Set-Cookie")
	assert.Contains(t, setCookie, "auth_token=")

	err := response.Body.Close()
	require.NoError(t, err)
}

func TestRegister_400BadRequest(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/register"
	var body = []byte("")
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)

	server := NewServer()
	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.Register)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusBadRequest, response.StatusCode)
	setCookie := response.Header.Get("Set-Cookie")
	assert.Empty(t, setCookie)

	err := response.Body.Close()
	require.NoError(t, err)
}

func TestRegister_409LoginAlreadyExists(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/register"
	var body = []byte(`{
		"login": "hello",
		"password": "world"
	}`)
	request1 := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)

	server := NewServer()
	router1 := chi.NewRouter()
	router1.HandleFunc(apiMethod, server.Register)
	writer1 := httptest.NewRecorder()
	router1.ServeHTTP(writer1, request1)

	response1 := writer1.Result()
	assert.Equal(t, response1.StatusCode, http.StatusOK)

	request2 := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)
	router2 := chi.NewRouter()
	router2.HandleFunc(apiMethod, server.Register)
	writer2 := httptest.NewRecorder()
	router2.ServeHTTP(writer2, request2)

	response2 := writer2.Result()
	assert.Equal(t, response2.StatusCode, http.StatusConflict)

	err := response1.Body.Close()
	require.NoError(t, err)
	err = response2.Body.Close()
	require.NoError(t, err)
}
