package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestHandler - вспомогательный обработчик для тестов
type TestHandler struct{}

func (h *TestHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
}

func TestCheckAndSetContentType(t *testing.T) {
	tests := []struct {
		name           string
		contentType    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Valid content type",
			contentType:    "application/json",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Valid content type with charset",
			contentType:    "application/json; charset=UTF-8",
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid content type",
			contentType:    "text/plain",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectError:    true,
		},
		{
			name:           "empty content type",
			contentType:    "",
			expectedStatus: http.StatusUnsupportedMediaType,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Content-Type", tt.contentType)

			// Создаем новый респонз рекордер.
			responseRecorder := httptest.NewRecorder()

			handler := &TestHandler{}
			middleware := CheckAndSetContenType(handler)
			middleware.ServeHTTP(responseRecorder, req)

			// Проверяем StatusCode.
			if status := responseRecorder.Code; status != tt.expectedStatus {
				t.Errorf("handler returned wrong status code: got %v want %v", responseRecorder.Code, tt.expectedStatus)
			}

			// Для случая с ошибкой проверяем тело ответа
			if tt.expectError {
				var resp resultMsg
				if err := json.Unmarshal(responseRecorder.Body.Bytes(), &resp); err != nil {
					t.Errorf("failed to unmarshal response: %v", err)
				}
				if !resp.IsError {
					t.Errorf("expected error response: got success")
				}
			} else {
				// Проверяем, что Content-Type установлен правильно
				if contentType := responseRecorder.Header().Get("Content-Type"); contentType != "application/json" {
					t.Errorf("handler returned wrong content type: got %s", contentType)
				}
			}
		})
	}
}
