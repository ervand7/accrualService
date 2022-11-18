package views

import (
	"bytes"
	"context"
	"fmt"
	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestWithdraw_200Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/balance/withdraw"
	amount := rand.Float64()
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer(
			[]byte(fmt.Sprintf(`{"order": "2377225624", "sum": %f}`, amount)),
		),
	)

	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	userID := database.UserIDFixture(server.Storage, "1", "1", token, t)
	orderID := 2200135834
	err := server.Storage.CreateOrder(context.TODO(), orderID, userID)
	assert.NoError(t, err)
	err = server.Storage.UpdateOrderAndAccrual(orderID, d.OrderStatus.NEW, amount)
	assert.NoError(t, err)

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.Withdraw)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)
	err = response.Body.Close()
	require.NoError(t, err)
}

func TestWithdraw_401Unauthorized(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/balance/withdraw"
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer([]byte(`{"order": "2377225624", "sum": 1.0}`)),
	)
	server := NewServer()

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.Withdraw)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}

func TestWithdraw_402PaymentRequired(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/balance/withdraw"
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer([]byte(`{"order": "2377225624", "sum": 1.0}`)),
	)
	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	database.UserIDFixture(server.Storage, "1", "1", token, t)

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.Withdraw)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusPaymentRequired, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}

func TestWithdraw_422InvalidOrderNumber(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/balance/withdraw"
	request := httptest.NewRequest(
		http.MethodPost,
		apiMethod,
		bytes.NewBuffer([]byte(`{"order": "1", "sum": 1.0}`)),
	)

	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	database.UserIDFixture(server.Storage, "1", "1", token, t)

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.Withdraw)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusUnprocessableEntity, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}
