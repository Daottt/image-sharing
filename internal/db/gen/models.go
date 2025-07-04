// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0

package db

import (
	"database/sql"
)

type Post struct {
	ID        int32
	UserID    int32
	ImagePath string
}

type Session struct {
	ID           string
	UserLogin    string
	AccessToken  string
	RefreshToken string
	IsRevoked    bool
	CreatedAt    sql.NullTime
	ExpiresAt    sql.NullTime
}

type User struct {
	ID          int32
	Name        string
	Description sql.NullString
}

type UsersAuth struct {
	UserID       int32
	Login        string
	PasswordHash string
	IsAdmin      sql.NullBool
}
