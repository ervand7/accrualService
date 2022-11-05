package views

import (
	"bytes"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserRegister(t *testing.T) {
	apiMethod := "/api/user/register"
	type want struct {
		statusCode int
	}
	tests := []struct {
		name string
		body string
		want want
	}{
		{
			name: "success 200",
			body: `{
				"login": "hello",
				"password": "world"
			}`,
			want: want{
				statusCode: http.StatusOK,
			},
		},
		{
			name: "fail 400",
			body: "",
			want: want{
				statusCode: http.StatusBadRequest,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			var body = []byte(tt.body)
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
			router.HandleFunc(apiMethod, server.UserRegister)
			writer := httptest.NewRecorder()
			router.ServeHTTP(writer, request)

			response := writer.Result()
			assert.Equal(t, tt.want.statusCode, response.StatusCode)

			err := response.Body.Close()
			require.NoError(t, err)
		})
	}
}

func TestUserRegister_LoginAlreadyExists(t *testing.T) {
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
	defer func() {
		server.Storage.DB.Downgrade()
	}()

	router1 := chi.NewRouter()
	router1.HandleFunc(apiMethod, server.UserRegister)
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
	router2.HandleFunc(apiMethod, server.UserRegister)
	writer2 := httptest.NewRecorder()
	router2.ServeHTTP(writer2, request2)

	response2 := writer2.Result()
	assert.Equal(t, response2.StatusCode, http.StatusConflict)

	err := response1.Body.Close()
	require.NoError(t, err)
	err = response2.Body.Close()
	require.NoError(t, err)
}
