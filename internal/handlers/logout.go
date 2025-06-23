package handlers

import "net/http"

func (handlers *Handlers) Logout(responseWriter http.ResponseWriter, gotRequest *http.Request) {

	handlers.logger.ZL.Debug("Logout handler started successfully")
	handlers.auth.Logout(responseWriter)
	sendResponse(
		false,
		"Logged out successfully",
		http.StatusOK,
		responseWriter)
	return
}
