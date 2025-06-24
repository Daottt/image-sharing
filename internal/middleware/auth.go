package middleware

import (
	"context"
	"errors"
	"net/http"
	"strings"

	"image-sharing/pkg/token"
)

type AuthKey struct{}

type AccessClaims struct {
	AccesToken string
	*token.UserClaims
}

func GetAuthMiddleware(tokenMaker *token.JWTMaker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			claims, err := verifyClaims(r, tokenMaker)
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			ctx := context.WithValue(r.Context(), AuthKey{}, claims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func verifyClaims(r *http.Request, tokenMaker *token.JWTMaker) (*AccessClaims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return nil, errors.New("authorization header is missing")
	}
	fields := strings.Fields(authHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return nil, errors.New("invalid authorization header")
	}
	token := fields[1]

	calims, err := tokenMaker.VerifyToken(token)
	if err != nil {
		return nil, err
	}
	return &AccessClaims{AccesToken: token, UserClaims: calims}, nil
}
