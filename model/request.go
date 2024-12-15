package model

type LoginPostBody struct {
	Username string `json:"username"`;
	Password string `json:"password"`
}