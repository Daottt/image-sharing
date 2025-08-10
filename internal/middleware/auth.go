package middleware

import (
	"context"
	"net/http"
	"strings"

	ssov1 "image-sharing/protos/gen"
)

type AuthKey struct{}

type AccessClaims struct {
	ID      int32
	Login   string
	IsAdmin bool
}

func GetAuthMiddleware(client ssov1.AuthClient) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "authorization header is missing", http.StatusUnauthorized)
				return
			}
			fields := strings.Fields(authHeader)
			if len(fields) != 2 || fields[0] != "Bearer" {
				http.Error(w, "invalid authorization header", http.StatusUnauthorized)
				return
			}
			token := fields[1]

			ctx := r.Context()

			claims, err := client.GetAuthorization(ctx, &ssov1.TokenRequest{AccessToken: token})
			if err != nil {
				http.Error(w, err.Error(), http.StatusUnauthorized)
				return
			}
			claimsCtx := context.WithValue(ctx, AuthKey{}, AccessClaims{ID: claims.Id, Login: claims.Login, IsAdmin: claims.IsAdmin})
			next.ServeHTTP(w, r.WithContext(claimsCtx))
		})
	}
}
