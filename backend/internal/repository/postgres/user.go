package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/yumikokawaii/sherry-archive/internal/apperror"
	"github.com/yumikokawaii/sherry-archive/internal/model"
)

type UserRepo struct{ db *sqlx.DB }

func NewUserRepo(db *sqlx.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *model.User) error {
	const q = `
		INSERT INTO users (id, username, email, password_hash, avatar_url, bio, created_at, updated_at)
		VALUES (:id, :username, :email, :password_hash, :avatar_url, :bio, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, u)
	return err
}

func (r *UserRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE email = $1`, email)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *UserRepo) GetByUsername(ctx context.Context, username string) (*model.User, error) {
	var u model.User
	err := r.db.GetContext(ctx, &u, `SELECT * FROM users WHERE username = $1`, username)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	return &u, err
}

func (r *UserRepo) Update(ctx context.Context, u *model.User) error {
	const q = `
		UPDATE users SET username=:username, email=:email, password_hash=:password_hash,
		avatar_url=:avatar_url, bio=:bio, updated_at=:updated_at
		WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, u)
	return err
}
