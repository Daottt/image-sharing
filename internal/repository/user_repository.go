package repository

import (
	"context"
	"database/sql"
	"errors"

	"image-sharing/internal/db/gen"
)

var ErrNotFound = errors.New("not found")

type UserRepository interface {
	GetUserByID(ctx context.Context, id int) (db.User, error)
	CreateUser(ctx context.Context, user db.CreateUserAuthParams) (db.User, error)
	UpdateUser(ctx context.Context, user db.UpdateUserParams) error
	DeleteUser(ctx context.Context, id int) error
}

type userRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewUserRepository(db *sql.DB, queries *db.Queries) UserRepository {
	return &userRepository{db: db, queries: queries}
}

func (r *userRepository) GetUserByID(ctx context.Context, id int) (db.User, error) {
	user, err := r.queries.GetUser(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return db.User{}, ErrNotFound
		}
		return db.User{}, err
	}
	return user, nil
}

func (r *userRepository) CreateUser(ctx context.Context, user db.CreateUserAuthParams) (db.User, error) {
	exists, err := r.queries.CheckLoginExists(ctx, user.Login)
	if err != nil {
		return db.User{}, err
	}
	if exists {
		return db.User{}, errors.New("login is taken")
	}

	tx, err := r.db.Begin()
	if err != nil {
		return db.User{}, err
	}
	defer tx.Rollback()
	qtx := r.queries.WithTx(tx)

	createdUser, err := qtx.CreateUser(ctx, db.CreateUserParams{Name: user.Login})
	if err != nil {
		return db.User{}, err
	}

	user.UserID = createdUser.ID
	err = qtx.CreateUserAuth(ctx, user)
	if err != nil {
		return db.User{}, err
	}

	return createdUser, tx.Commit()
}

func (r *userRepository) UpdateUser(ctx context.Context, user db.UpdateUserParams) error {
	return r.queries.UpdateUser(ctx, user)
}

func (r *userRepository) DeleteUser(ctx context.Context, id int) error {
	return r.queries.DeletUser(ctx, int32(id))
}
