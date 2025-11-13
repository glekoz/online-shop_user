package models

// то, что используется при входе в аккаунт и хранится в токене
type UserTokenWithPassword struct {
	ID             string
	Name           string
	HashedPassword string `json:"-"`
	IsModer        bool
	IsAdmin        bool
	IsCore         bool
}

type UserToken struct {
	ID      string
	Name    string
	IsModer bool
	IsAdmin bool
	IsCore  bool
}

// то, что видно пользователю на его странице профиля
type User struct {
	ID               string
	Name             string
	Email            string
	IsEmailConfirmed bool
	// день рождения
	// адрес
	// телефон
}

// то, что видно администратору в интерфейсе управления пользователями
type UserInfo struct {
	ID               string
	Name             string
	Email            string
	IsEmailConfirmed bool
	IsModer          bool
	IsAdmin          bool
	IsCore           bool
}

type Admin struct {
	ID     string
	IsCore bool
}
