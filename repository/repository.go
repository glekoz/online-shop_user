package repository

import (
	"context"
	"errors"

	"github.com/glekoz/online-shop_user/repository/db"
	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	q *db.Queries
	p *pgxpool.Pool
}

func New(dsn string) (*Repository, error) {
	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, err
	}
	queries := db.New(pool)
	return &Repository{
		q: queries,
		p: pool,
	}, nil
}

func (r *Repository) CreateUser(ctx context.Context, id, name, email, hashedPassword string) error {
	err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:       id,
		Name:     name,
		Email:    email,
		Password: hashedPassword,
	})
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			if errp.Code == UniqueViolationCode {
				return myerrors.ErrAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (r *Repository) PromoteModer(ctx context.Context, id string) error {
	err := r.q.PromoteModer(ctx, id)
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			switch errp.Code {
			case ForeignKeyViolationCode:
				return myerrors.ErrNotFound
			case UniqueViolationCode:
				return myerrors.ErrAlreadyExists
			}
		}
		return err
	}
	return nil
}

func (r *Repository) PromoteAdmin(ctx context.Context, id string, isCore bool) error {
	tx, err := r.p.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	qtx := r.q.WithTx(tx)

	err = qtx.PromoteAdmin(ctx, id)
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			switch errp.Code {
			case ForeignKeyViolationCode:
				return myerrors.ErrNotFound
			case UniqueViolationCode:
				return myerrors.ErrAlreadyExists
			}
		}
		return err
	}

	err = qtx.PromoteModer(ctx, id)
	if err != nil {
		var errp *pgconn.PgError
		if errors.As(err, &errp) {
			if errp.Code == UniqueViolationCode {
				err = nil
			}
		}
		if err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func (r *Repository) PromoteCoreAdmin(ctx context.Context, id string) error {
	num, err := r.q.PromoteCoreAdmin(ctx, id)
	if err != nil {
		return err
	}
	if num != 1 {
		return myerrors.ErrNotFound
	}
	return nil
}
