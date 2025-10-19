package myerrors

import "errors"

var (
	ErrNotFound           = errors.New("no result found")
	ErrInternal           = errors.New("something goes wrong")
	ErrAlreadyExists      = errors.New("already exists")
	ErrInvalidCredentials = errors.New("invalid credentials") // неверный логин или пароль
	ErrForbidden          = errors.New("forbidden")           // нет прав
	ErrInvalidToken       = errors.New("invalid token")
)
