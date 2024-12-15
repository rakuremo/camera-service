package model

type User struct {
	ID string `json:"id"`
	Name string `json:"name"`
	Username string `json:"username"`
	Password string `json:"password"`
	Role string `json:"role"`
	Status string `json:"status"`
}

