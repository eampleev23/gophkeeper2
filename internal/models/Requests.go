package models

// UserRegReq - модель запроса на регистрацию.
type UserRegReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}

// UserLoginReq - модель запроса на авторизацию.
type UserLoginReq struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
