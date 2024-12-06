package dto

type User struct {
	Email    string `validate:"required,email"`
	Password string `validate:"required,min=6,max=32"`
}
