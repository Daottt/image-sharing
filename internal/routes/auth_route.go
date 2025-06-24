package routes

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	db "image-sharing/internal/db/gen"
	"image-sharing/internal/middleware"
	"image-sharing/internal/repository"
	"image-sharing/pkg/password"
	"image-sharing/pkg/token"
)

const AccessTokenDuration = 15 * time.Minute
const RefreshTokenDuration = 24 * time.Hour

type LoginRequest struct {
	Login    string `json:"login"`
	Password string `json:"password"`
}
type LoginResponse struct {
	SessionID             string    `json:"session_id"`
	AccessToken           string    `json:"access_token"`
	RefreshToken          string    `json:"refresh_token"`
	AccessTokenExpiresAt  time.Time `json:"access_token_expires_at"`
	RefreshTokenExpiresAt time.Time `json:"refrsh_token_expires_at"`
	Login                 string    `json:"login"`
}

type RenewAccessTokenRequest struct {
	RefreshToken string `json:"refresh_token"`
}
type RenewAccessTokenResponse struct {
	AccessToken          string    `json:"access_token"`
	AccessTokenExpiresAt time.Time `json:"access_token_expires_at"`
}

type AuthRoute struct {
	repo       repository.AuthRepository
	tokenMaker *token.JWTMaker
}

func NewAuthRoute(repo repository.AuthRepository, tokenMaker *token.JWTMaker) *AuthRoute {
	return &AuthRoute{repo: repo, tokenMaker: tokenMaker}
}

func (a *AuthRoute) LoginUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&user); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Login == "" || user.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	userAuth, err := a.repo.GetUserAuth(ctx, user.Login)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	err = password.CheckPasswrod(user.Password, userAuth.PasswordHash)
	if err != nil {
		http.Error(w, "wrong password", http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := a.tokenMaker.CreateToken(userAuth.UserID, userAuth.Login, userAuth.IsAdmin.Bool, AccessTokenDuration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	refreshToken, refreshClaims, err := a.tokenMaker.CreateToken(userAuth.UserID, userAuth.Login, userAuth.IsAdmin.Bool, RefreshTokenDuration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	session, err := a.repo.CreateSession(ctx, db.CreateSessionParams{
		ID:           refreshClaims.RegisteredClaims.ID,
		UserLogin:    userAuth.Login,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		IsRevoked:    false,
		ExpiresAt:    sql.NullTime{Time: refreshClaims.RegisteredClaims.ExpiresAt.Time, Valid: true},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(LoginResponse{
		SessionID:             session.ID,
		AccessToken:           accessToken,
		RefreshToken:          refreshToken,
		AccessTokenExpiresAt:  accessClaims.RegisteredClaims.ExpiresAt.Time,
		RefreshTokenExpiresAt: refreshClaims.RegisteredClaims.ExpiresAt.Time,
		Login:                 userAuth.Login})
}

func (a *AuthRoute) RenewAccessToken(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var req RenewAccessTokenRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	if req.RefreshToken == "" {
		http.Error(w, "access_token required", http.StatusBadRequest)
		return
	}

	refreshClaims, err := a.tokenMaker.VerifyToken(req.RefreshToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	session, err := a.repo.GetSessionByID(ctx, refreshClaims.RegisteredClaims.ID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if session.IsRevoked {
		http.Error(w, "session revoked", http.StatusUnauthorized)
		return
	}

	if session.UserLogin != refreshClaims.Login {
		http.Error(w, "invaild session", http.StatusUnauthorized)
		return
	}

	accessToken, accessClaims, err := a.tokenMaker.CreateToken(refreshClaims.ID, refreshClaims.Login, refreshClaims.IsAdmin, AccessTokenDuration)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	err = a.repo.RenewSession(ctx, session.ID, accessToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(RenewAccessTokenResponse{AccessToken: accessToken, AccessTokenExpiresAt: accessClaims.RegisteredClaims.ExpiresAt.Time})
}

func (a *AuthRoute) LogoutUser(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.AuthKey{}).(*middleware.AccessClaims)

	err := a.repo.RevokeSession(r.Context(), claims.AccesToken)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (a *AuthRoute) RevokeSessions(w http.ResponseWriter, r *http.Request) {
	claims := r.Context().Value(middleware.AuthKey{}).(*middleware.AccessClaims)
	fmt.Println(claims.Login)
	err := a.repo.RevokeAllSessions(r.Context(), claims.Login)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
