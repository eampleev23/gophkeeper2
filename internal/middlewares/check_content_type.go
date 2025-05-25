package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
)

type errorMsg struct {
	ErrorMessage string `json:"error_message"`
}

func CheckContentType(next http.Handler) http.Handler {
	return http.HandlerFunc(func(responseWriter http.ResponseWriter, gotRequest *http.Request) {
		contentType := gotRequest.Header.Get("Content-Type")
		if !strings.HasPrefix(contentType, "application/json") {
			errorMsg := errorMsg{ErrorMessage: "Content-Type header is not application/json"}
			msg, _ := json.Marshal(errorMsg)
			responseWriter.WriteHeader(http.StatusUnsupportedMediaType)
			responseWriter.Write(msg)
			return
		}
	})
}
