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

type UserInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
}
