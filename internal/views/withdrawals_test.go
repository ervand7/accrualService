package views

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/ervand7/go-musthave-diploma-tpl/internal/database"
	d "github.com/ervand7/go-musthave-diploma-tpl/internal/datamapping"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserWithdrawals_200Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/withdrawals"
	request := httptest.NewRequest(
		http.MethodGet,
		apiMethod,
		bytes.NewBuffer([]byte("")),
	)
	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)

	userID := database.UserIDFixture(server.Storage, "1", "1", token, t)
	ctx := context.TODO()
	orderID := 2200135834
	WithdrawalOrderID := 1169934492
	amount := 10.0
	err := server.Storage.CreateOrder(ctx, orderID, userID)
	assert.NoError(t, err)
	err = server.Storage.UpdateOrderAndAccrual(ctx, orderID, d.OrderStatus.NEW, amount)
	assert.NoError(t, err)
	err = server.Storage.CreateWithdraw(ctx, userID, WithdrawalOrderID, amount-1)
	assert.NoError(t, err)

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.GetUserWithdrawals)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	bodyRaw, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	var respBodyData []map[string]interface{}
	err = json.Unmarshal(bodyRaw, &respBodyData)
	assert.NoError(t, err)
	assert.Len(t, respBodyData, 1)
	assert.Equal(
		t,
		strconv.Itoa(WithdrawalOrderID),
		respBodyData[0]["order"],
	)

	err = response.Body.Close()
	require.NoError(t, err)
}

func TestGetUserWithdrawals_204Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/withdrawals"
	request := httptest.NewRequest(
		http.MethodGet,
		apiMethod,
		bytes.NewBuffer([]byte("")),
	)
	server := NewServer()
	token := uuid.New().String()
	server.SetRequestCookie(token, request)
	database.UserIDFixture(server.Storage, "1", "1", token, t)

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.GetUserWithdrawals)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}

func TestGetUserWithdrawals_401Unauthorized(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/withdrawals"
	request := httptest.NewRequest(
		http.MethodGet,
		apiMethod,
		bytes.NewBuffer([]byte("")),
	)

	server := NewServer()
	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.GetUserWithdrawals)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}
