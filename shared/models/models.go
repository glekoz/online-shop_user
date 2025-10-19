package models

type User struct {
	ID    string
	Name  string
	Email string
}

type Admin struct {
	ID     string
	IsCore bool
}
