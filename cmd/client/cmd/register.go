package cmd

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/AlecAivazis/survey/v2"
)

type UserRegReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

func registerUser() {

	var (
		login    string
		password string
	)

	// Ввод логина
	survey.AskOne(&survey.Input{
		Message: "Введите логин:",
	}, &login)

	// Ввод пароля (с маскировкой)
	survey.AskOne(&survey.Password{
		Message: "Введите пароль:",
	}, &password)

	// Подтверждение пароля
	var confirmPassword string
	survey.AskOne(&survey.Password{Message: "Повторите пароль:"}, &confirmPassword)

	if password != confirmPassword {
		fmt.Println("Ошибка: пароли не совпадают.")
		return
	}

	// Отправка данных на сервер
	userReqReq := UserRegReq{Login: login, Password: password}
	jsonUserRegReq, err := json.Marshal(userReqReq)
	if err != nil {
		fmt.Println(err)
		return
	}
	response, err := http.Post("http://127.0.0.1:8080/api/user/registration/", "application/json", bytes.NewBuffer(jsonUserRegReq))
	if err != nil {
		fmt.Println("Ошибка при отправке запроса", err)
		return
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusOK {
		fmt.Println("Регистрация успешна")
	} else {
		fmt.Println("Ошибка регистрации:", response.StatusCode)
	}
}
