package repository

import (
	"context"
	"database/sql"

	"image-sharing/internal/db/gen"
)

type AuthRepository interface {
	GetUserAuth(ctx context.Context, login string) (db.UsersAuth, error)
	GetSessionByID(ctx context.Context, id string) (db.Session, error)
	CreateSession(ctx context.Context, session db.CreateSessionParams) (db.Session, error)
	RevokeSession(ctx context.Context, accessToken string) error
	RevokeAllSessions(ctx context.Context, login string) error
	RenewSession(ctx context.Context, id string, accessToken string) error
}

type authRepository struct {
	db      *sql.DB
	queries *db.Queries
}

func NewAuthRepository(db *sql.DB, queries *db.Queries) AuthRepository {
	return &authRepository{db: db, queries: queries}
}

func (r *authRepository) GetUserAuth(ctx context.Context, login string) (db.UsersAuth, error) {
	user, err := r.queries.GetUserAuth(ctx, login)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.UsersAuth{}, ErrNotFound
		}
		return db.UsersAuth{}, err
	}
	return user, nil
}

func (r *authRepository) GetSessionByID(ctx context.Context, id string) (db.Session, error) {
	session, err := r.queries.GetSession(ctx, id)
	if err != nil {
		if err == sql.ErrNoRows {
			return db.Session{}, ErrNotFound
		}
		return db.Session{}, err
	}
	return session, nil
}

func (r *authRepository) CreateSession(ctx context.Context, session db.CreateSessionParams) (db.Session, error) {
	createdSession, err := r.queries.CreateSession(ctx, session)
	if err != nil {
		return db.Session{}, err
	}
	return createdSession, nil
}

func (r *authRepository) RenewSession(ctx context.Context, id string, accessToken string) error {
	return r.queries.RenewSession(ctx, db.RenewSessionParams{ID: id, AccessToken: accessToken})
}

func (r *authRepository) RevokeSession(ctx context.Context, accessToken string) error {
	return r.queries.RevokeSession(ctx, accessToken)
}

func (r *authRepository) RevokeAllSessions(ctx context.Context, login string) error {
	return r.queries.RevokeSessionsByLogin(ctx, login)
}
