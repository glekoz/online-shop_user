package ports

import (
	"context"

	"github.com/Gleb988/online-shop_user/internal/models"
)

type DBAPI interface {
	Get(ctx context.Context, id int) (models.User, error)
	Save(ctx context.Context, user models.UserDTO) (int, error)
}

type AppAPI interface {
	GetUser(ctx context.Context, id int) (models.User, error)
	SaveUser(ctx context.Context, user models.UserDTO) (int, error)
}

/*
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
