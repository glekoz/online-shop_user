package repository

import "errors"

const (
	ForeignKeyViolationCode = "23503"
	UniqueViolationCode     = "23505"
)

var (
	ErrAlreadyExists = errors.New("aslready exist")
	ErrNotFound      = errors.New("no result found")
)
