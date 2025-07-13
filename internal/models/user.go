package models

type User struct {
	ID int // подумать о замене на int32 или тип в БД поменять
	// Боднер говорит, что такой инт - хорошее начало
	Name  string
	Email string
	//HashedPassword []byte
	//Created        time.Time
}

type UserDTO struct {
	Name     string
	Email    string
	Password string
}

// models - просто общая модель для сервиса
// main - потребитель, он определяет интерфейс
// db  - адаптер, так сказать, конкретное исполнение, надо ещё sqlc добавить
