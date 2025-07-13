package main

import (
	"fmt"
)

func main() {
	/*
		ctx := context.Background() // добавить ограничения для конекста
		dsn := "postgresql://postgres:postgres@db:5432/order?sslmode=disable"
		repo := repository.NewUserRepository(ctx, dsn)
		app := NewApplication(repo)
		grpcService := grpc.NewGrpcService(app)
		grpcService.Run()
	*/
	fmt.Println("Получилось")
}

/*
func main() {
	// some logic
	//asd := NewDB(db.NewUserDB(context.Background(), "потом"))
	// а потом при необходимости то, что внутри NewDB() можно поменять на MySQL, например
	// вызов openDB() должен быть в init() или в main()

	ctx := context.Background()                                          // добавить параметры
	dsn := "postgresql://postgres:postgres@db:5432/user?sslmode=disable" // получать строку из окружения
	db := MustOpenDB(ctx, dsn)


}

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
