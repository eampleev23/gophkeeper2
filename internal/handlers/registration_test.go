package handlers

import (
	"bytes"
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/auth"
	"github.com/eampleev23/gophkeeper2.git/internal/logger"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/eampleev23/gophkeeper2.git/internal/server_config"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandlers_Registration(t *testing.T) {

	type want struct {
		//contentType  string
		statusCode   int
		jsonResponse resultMsg
	}

	tests := []struct {
		name        string
		requestUrl  string
		requestBody models.UserRegReq
		tableUsers  interface{}
		want        want
	}{
		{
			name:       "test status 200 OK with auto authorization",
			requestUrl: "/api/user/registration/",
			requestBody: models.UserRegReq{
				Login:    "Петя",
				Password: "петр лучший",
			},
			tableUsers: map[int]models.User{
				1: {Login: "Александр"},
				2: {Login: "Андрей"},
				3: {Login: "Валерка"},
			},
			want: want{
				//contentType: "application/json",
				statusCode: 200,
				jsonResponse: resultMsg{
					IsError:       false,
					ResultMessage: "Success authenticate user after successful registration",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			jsonBody, err := json.Marshal(tt.requestBody)
			if err != nil {
				t.Fatalf("Failed to marshal request body: %v", err)
			}

			request := httptest.NewRequest(http.MethodPost, tt.requestUrl, bytes.NewBuffer(jsonBody))
			request.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			// Создаем логгер и прочее для получения в конечном итоге хэндлеров.
			l, err := logger.NewZapLogger("info")
			if err != nil {
				t.Log(err)
			}
			c := server_config.NewServerConfig()
			au, err := auth.Initialize(c, l)
			if err != nil {
				t.Log(err)
			}
			s, err := store.NewDBStore(c, l)
			if err != nil {
				t.Log(err)
			}
			handlers, _ := NewHandlers(s, c, l, au)
			h := http.HandlerFunc(handlers.Registration)
			h(w, request)

			result := w.Result()

			assert.Equal(t, tt.want.statusCode, result.StatusCode)
			//assert.Equal(t, tt.want.contentType, result.Header.Get("Content-Type"))

			bytesJsonResponse, err := io.ReadAll(result.Body)
			require.NoError(t, err)
			err = result.Body.Close()
			require.NoError(t, err)

			var jsonResponse resultMsg
			err = json.Unmarshal(bytesJsonResponse, &jsonResponse)
			require.NoError(t, err)

			assert.Equal(t, tt.want.jsonResponse, jsonResponse)

		})
	}
}
