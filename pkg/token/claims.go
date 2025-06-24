package token

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

type UserClaims struct {
	ID      int32  `json:"id"`
	Login   string `json:"login"`
	IsAdmin bool   `json:"is_admin"`
	jwt.RegisteredClaims
}

func NewUserClaims(id int32, login string, isAdmin bool, duration time.Duration) (*UserClaims, error) {
	tokenID, err := uuid.NewRandom()
	if err != nil {
		return nil, err
	}

	userClaims := &UserClaims{
		ID:      id,
		Login:   login,
		IsAdmin: isAdmin,
		RegisteredClaims: jwt.RegisteredClaims{
			ID:        tokenID.String(),
			Subject:   login,
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(duration)),
		},
	}
	return userClaims, nil
}
