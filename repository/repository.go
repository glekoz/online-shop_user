package repository

import (
	"context"
	"errors"

	"github.com/glekoz/online-shop_user/repository/db"
	"github.com/glekoz/online-shop_user/shared/models"
	"github.com/glekoz/online-shop_user/shared/myerrors"
	"github.com/jackc/pgx/v5"
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

func (r *Repository) PromoteAdmin(ctx context.Context, id string) error {
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

func (r *Repository) GetUserByID(ctx context.Context, id string) (models.User, error) {
	u, err := r.q.GetUserByID(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.User{}, myerrors.ErrNotFound
		}
		return models.User{}, err
	}
	return models.User{
		ID:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		IsEmailConfirmed: u.EmailConfirmed,
	}, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (models.UserTokenWithPassword, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.UserTokenWithPassword{}, myerrors.ErrNotFound
		}
		return models.UserTokenWithPassword{}, err
	}
	return models.UserTokenWithPassword{
		ID:             u.ID,
		Name:           u.Name,
		HashedPassword: u.Password,
		IsModer:        u.IsModer,
		IsAdmin:        u.IsAdmin,
		IsCore:         u.IsCore,
	}, nil
}

func (r *Repository) GetUsersByEmail(ctx context.Context, email string) ([]models.UserInfo, error) {
	us, err := r.q.GetUsersByEmail(ctx, email+"%")
	if err != nil {
		return nil, err
	}
	if len(us) < 1 {
		return nil, myerrors.ErrNotFound
	}
	users := make([]models.UserInfo, 0, len(us))
	for _, u := range us {
		users = append(users, models.UserInfo{
			ID:               u.ID,
			Name:             u.Name,
			Email:            u.Email,
			IsEmailConfirmed: u.EmailConfirmed,
			IsModer:          u.IsModer,
			IsAdmin:          u.IsAdmin,
			IsCore:           u.IsCore,
		})
	}
	return users, nil
}

func (r *Repository) GetModer(ctx context.Context, id string) (string, error) {
	m, err := r.q.GetModer(ctx, id)
	if err != nil { // if m == ""
		if errors.Is(err, pgx.ErrNoRows) {
			return "", myerrors.ErrNotFound
		}
		return "", err
	}
	return m, nil
}

func (r *Repository) GetAdmin(ctx context.Context, id string) (models.Admin, error) {
	a, err := r.q.GetAdmin(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return models.Admin{}, myerrors.ErrNotFound
		}
		return models.Admin{}, err
	}
	return models.Admin{
		ID:     a.ID,
		IsCore: a.IsCore,
	}, nil
}

func (r *Repository) ConfirmEmail(ctx context.Context, id string) error {
	n, err := r.q.ConfirmEmail(ctx, id)
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound
	}
	return nil
}

func (r *Repository) ChangeName(ctx context.Context, id, newName string) error {
	n, err := r.q.ChangeName(ctx, db.ChangeNameParams{
		ID:   id,
		Name: newName,
	})
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound
	}
	return nil
}

func (r *Repository) ChangePassword(ctx context.Context, id, newHashedPassword string) error {
	n, err := r.q.ChangePassword(ctx, db.ChangePasswordParams{
		ID:       id,
		Password: newHashedPassword,
	})
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound // хотя это не должно произойти
	}
	return nil
}

func (r *Repository) ChangeEmail(ctx context.Context, id, newEmail string) error {
	n, err := r.q.ChangeEmail(ctx, db.ChangeEmailParams{
		ID:    id,
		Email: newEmail,
	})
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound // хотя это не должно произойти
	}
	return nil
}

func (r *Repository) DeleteUser(ctx context.Context, id string) error {
	n, err := r.q.DeleteUser(ctx, id)
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteModer(ctx context.Context, id string) error {
	n, err := r.q.DeleteModer(ctx, id)
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound
	}
	return nil
}

func (r *Repository) DeleteAdmin(ctx context.Context, id string) error {
	n, err := r.q.DeleteAdmin(ctx, id)
	if err != nil {
		return err
	}
	if n != 1 {
		return myerrors.ErrNotFound
	}
	return nil
}
