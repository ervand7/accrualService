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
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetUserOrders_200Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/orders"
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
	ordersNumbers := []int{2200135834, 1169934492}
	for _, orderNumber := range ordersNumbers {
		err := server.Storage.CreateOrder(ctx, orderNumber, userID)
		assert.NoError(t, err)
	}

	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.GetUserOrders)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusOK, response.StatusCode)

	bodyRaw, err := io.ReadAll(response.Body)
	assert.NoError(t, err)
	var respBodyData []map[string]interface{}
	err = json.Unmarshal(bodyRaw, &respBodyData)
	assert.NoError(t, err)
	assert.Len(t, respBodyData, len(ordersNumbers))

	for index, value := range respBodyData {
		assert.Equal(
			t,
			strconv.Itoa(ordersNumbers[index]),
			value["number"],
		)
	}

	err = response.Body.Close()
	require.NoError(t, err)
}

func TestGetUserOrders_204Success(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/orders"
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
	router.HandleFunc(apiMethod, server.GetUserOrders)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusNoContent, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}

func TestGetUserOrders_401Unauthorized(t *testing.T) {
	defer database.Downgrade()
	apiMethod := "/api/user/orders"
	request := httptest.NewRequest(
		http.MethodGet,
		apiMethod,
		bytes.NewBuffer([]byte("")),
	)

	server := NewServer()
	router := chi.NewRouter()
	router.HandleFunc(apiMethod, server.GetUserOrders)
	writer := httptest.NewRecorder()
	router.ServeHTTP(writer, request)

	response := writer.Result()
	assert.Equal(t, http.StatusUnauthorized, response.StatusCode)
	err := response.Body.Close()
	require.NoError(t, err)
}
