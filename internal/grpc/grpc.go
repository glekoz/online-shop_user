package grpc

// ЭТО КАК ХЕНДЛЕР В ХТТП
// СЮДА НАДО ДАТЬ БИЗНЕС ЛОГИКУ КАК ЗАВИСИМОСТЬ
// ПРЕОБРАЗОВАТЬ ПРОТОБАФФ В МЕСТНЫЕ ТИПЫ И ВЫЗВАТЬ БизЛогику
// А обычную проверку входных данных типа регулярные выражения для почты
// оставить для апи шлюза (хотя и здесь можно)

import (
	"context"

	"github.com/Gleb988/online-shop_proto/protouser"
	"github.com/Gleb988/online-shop_user/internal/models"
)

func (a *grpcService) Create(ctx context.Context, pu *protouser.CreateUserRequest) (*protouser.CreateUserResponse, error) {
	user := models.UserDTO{
		Name:     pu.Name,
		Email:    pu.Email,
		Password: pu.Password,
	}
	id, err := a.api.SaveUser(ctx, user)
	if err != nil {
		return nil, err
	}
	return &protouser.CreateUserResponse{Id: int32(id)}, nil
}

func (a *grpcService) Get(ctx context.Context, pu *protouser.GetUserRequest) (*protouser.GetUserResponse, error) {
	id := int(pu.Id)
	user, err := a.api.GetUser(ctx, id)
	if err != nil {
		return nil, err
	}
	return &protouser.GetUserResponse{
		Id:    int32(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

/*
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
