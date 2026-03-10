package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
)

func CheckAndSetContenType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, gotRequest *http.Request) {
		// GET/HEAD и /health — без проверки Content-Type (нет тела или health-check).
		if gotRequest.Method == http.MethodGet || gotRequest.Method == http.MethodHead ||
			strings.TrimSuffix(gotRequest.URL.Path, "/") == "/health" {
			next.ServeHTTP(responseWriter, gotRequest.WithContext(gotRequest.Context()))
			return
		}
		contentType := gotRequest.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			resultMsg := resultMsg{IsError: true, ResultMessage: "Content-Type header is not application/json"}
			msg, _ := json.Marshal(resultMsg)
			responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
			responseWriter.Write(msg)
			return
		}
		responseWriter.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(responseWriter, gotRequest.WithContext(gotRequest.Context()))
	})
}
