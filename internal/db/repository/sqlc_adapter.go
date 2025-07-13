package repository

import (
	"context"
	"log"

	"github.com/Gleb988/online-shop_user/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	q    *Queries
	pool *pgxpool.Pool
}

func NewUserRepository(ctx context.Context, dsn string) *UserRepository {
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatal("DB conn is not created")
	}
	return &UserRepository{
		q:    New(pool),
		pool: pool,
	}
}

func (r *UserRepository) Get(ctx context.Context, id int) (models.User, error) {
	//if user, err := r.q.Get(ctx, id); err != nil {
	//	return models.User{}, err
	//}
	user, err := r.q.Get(ctx, int32(id))
	if err != nil {
		return models.User{}, err
	}
	return models.User{
		ID:    int(user.ID),
		Name:  user.Name,
		Email: user.Email,
	}, nil
}

func (r *UserRepository) Save(ctx context.Context, user models.UserDTO) (int, error) { // заменить на name, email string вместо user models.User?
	params := SaveParams{
		Name:           user.Name,
		Email:          user.Email,
		HashedPassword: user.Password,
	}
	id, err := r.q.Save(ctx, params)
	if err != nil {
		return 0, err
	}
	// добавить проверку ошибки на уникальность email, если нужно
	return int(id), nil
}
