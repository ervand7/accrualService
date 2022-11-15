package views

import (
	"bytes"
	"context"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func loadOrder(
	server Server,
	request *http.Request,
	assertStatus int,
	t *testing.T,
) {
	router := chi.NewRouter()
	router.HandleFunc("/api/user/orders", server.LoadOrder)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, assertStatus, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}

func createUser(login, password, token string, server Server, t *testing.T) {
	err := server.Storage.CreateUser(
		context.TODO(), login, password, token,
	)
	assert.NoError(t, err)
}

func TestLoadOrder_Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/orders"
	var body = []byte("12345678903")
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)

	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	createUser("1", "2", token, server, t)
	loadOrder(server, request, http.StatusAccepted, t)

	request = httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)
	server.SetRequestCookie(token, request)
	loadOrder(server, request, http.StatusOK, t)
}

func TestLoadOrder_400BadRequest(t *testing.T) {
	defer database.Downgrade()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/user/orders",
		bytes.NewBuffer([]byte("")),
	)

	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	createUser("1", "2", token, server, t)
	loadOrder(server, request, http.StatusBadRequest, t)
}

func TestLoadOrder_401Unauthorized(t *testing.T) {
	defer database.Downgrade()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/user/orders",
		bytes.NewBuffer([]byte("12345678903")),
	)

	server := NewServer()
	loadOrder(server, request, http.StatusUnauthorized, t)
}

func TestLoadOrder_409Conflict(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/orders"
	var body = []byte("12345678903")
	server := NewServer()

	token := uuid.New().String()
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)
	server.SetRequestCookie(token, request)
	createUser("hello", "world", token, server, t)
	loadOrder(server, request, http.StatusAccepted, t)

	token = uuid.New().String()
	request = httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(body),
	)
	server.SetRequestCookie(token, request)
	createUser("world", "hello", token, server, t)
	loadOrder(server, request, http.StatusConflict, t)
}

func TestLoadOrder_422InvalidOrderNumber(t *testing.T) {
	defer database.Downgrade()
	request := httptest.NewRequest(
		http.MethodPost,
		"/api/user/orders",
		bytes.NewBuffer([]byte("123")),
	)

	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	createUser("1", "2", token, server, t)
	loadOrder(server, request, http.StatusUnprocessableEntity, t)
}
