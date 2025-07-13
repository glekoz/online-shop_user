package main

import (
	"context"

	"github.com/Gleb988/online-shop_user/internal/models"
	"github.com/Gleb988/online-shop_user/internal/ports"
	"golang.org/x/crypto/bcrypt"
)

type Application struct {
	db ports.DBAPI
	// логгеры и тому подобное, см. "Let's Go" by Alex Edwards
	// а логгеры могут быть в файл или даже обращаться к отдельному серверу
}

func NewApplication(dbAPI ports.DBAPI) *Application {
	return &Application{db: dbAPI}
}

// добавить тут кастомные ошибки, чтобы в хендлерах понятные ошибки отправлять
func (db *Application) GetUser(ctx context.Context, id int) (models.User, error) {
	user, err := db.db.Get(ctx, id)
	if err != nil {
		return models.User{}, err
	}
	return user, nil
}

// сделать DTO, а не использовать models.User напрямую и не передавать пароль в Save
// или использовать models.User только для чтения, а для записи использовать другой тип
func (db *Application) SaveUser(ctx context.Context, user models.UserDTO) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, models.ErrHashingPassword
	}
	user.Password = string(hashedPassword)
	//  а также рассылка email с подтверждением регистрации, если нужно
	//  и т.д.
	id, err := db.db.Save(ctx, user)
	if err != nil {
		return 0, err
	}
	return id, nil
}

/*
func MustOpenDB(ctx context.Context, dsn string) *DB {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		panic(err)
	}
	repo := repository.NewUserRepository(pool)
	return NewDB(repo)
}

type DBAPI interface {
	Get(ctx context.Context, id int) (models.User, error)
	Save(ctx context.Context, user models.UserDTO) (int, error)
}

type DB struct {
	db DBAPI
}

func NewDB(dbAPI DBAPI) *DB {
	return &DB{db: dbAPI}
}

func (db *DB) getUser(ctx context.Context, id int) models.User {
	user, _ := db.db.Get(ctx, id)
	return user
}

// сделать DTO, а не использовать models.User напрямую и не передавать пароль в Save
// или использовать models.User только для чтения, а для записи использовать другой тип
func (db *DB) saveUser(ctx context.Context, user models.UserDTO) (int, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), 12)
	if err != nil {
		return 0, models.ErrHashingPassword
	}
	user.Password = string(hashedPassword)
	//  а также рассылка email с подтверждением регистрации, если нужно
	//  и т.д.
	id, err := db.db.Save(ctx, user)
	if err != nil {
		return 0, err
	}
	return id, nil
}
*/
