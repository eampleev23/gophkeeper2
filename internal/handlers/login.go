package handlers

import (
	"encoding/json"
	"github.com/eampleev23/gophkeeper2.git/internal/models"
	"github.com/eampleev23/gophkeeper2.git/internal/store"
	"net/http"
)

/*
На вход хэндлер ожидает json такого формата:
{
    "login": "<login>",
    "password": "<password>"
}
*/

func (handlers *Handlers) Login(responseWriter http.ResponseWriter, gotRequest *http.Request) {

	handlers.logger.ZL.Info("Handling /login")

	// Создаем модель для парсинга запроса.
	var userLoginReq models.UserLoginReq

	// Пробуем спарсить запрос в модель.
	decoder := json.NewDecoder(gotRequest.Body)
	if err := decoder.Decode(&userLoginReq); err != nil {
		sendResponse(
			true,
			"Not a valid user login request",
			http.StatusBadRequest,
			responseWriter)
		return
	}

	handlers.logger.ZL.Info("Request is correct")

	// Дополнительная проверка на пустые значения
	if userLoginReq.Login == "" || userLoginReq.Password == "" {
		sendResponse(
			true,
			"Login and password are required",
			http.StatusBadRequest,
			responseWriter)
		return
	}

	foundUser, err := handlers.store.GetUserByLogin(gotRequest.Context(), userLoginReq)
	if err != nil {
		sendResponse(
			true,
			"User with this login does not exist",
			http.StatusNotFound,
			responseWriter)
		return
	}

	isCorrectPassword, err := store.VerifyPassword(userLoginReq.Password, foundUser.PasswordHash)
	if err != nil {
		sendResponse(
			true,
			"Internal server error",
			http.StatusInternalServerError,
			responseWriter)
		return
	}

	if !isCorrectPassword {
		sendResponse(
			true,
			"The passwords don't match",
			http.StatusUnauthorized,
			responseWriter)
		return
	}

	sendResponse(
		false,
		"Successfully logged in",
		http.StatusOK,
		responseWriter)
	return

}
