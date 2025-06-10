package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
)

func CheckAndSetContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, gotRequest *http.Request) {
		if gotRequest.URL.Scheme != "https" && gotRequest.TLS == nil {
			resultMsg := resultMsg{IsError: true, ResultMessage: "HTTPS required"}
			msg, _ := json.Marshal(resultMsg)
			responseWriter.WriteHeader(http.StatusForbidden)
			responseWriter.Write(msg)
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
