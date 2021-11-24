package model

type User struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Authentication struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type Token struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	TokenString string `json:"token"`
}
