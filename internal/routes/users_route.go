package routes

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"

	"image-sharing/internal/db/gen"
	"image-sharing/internal/middleware"
	"image-sharing/internal/repository"
	"image-sharing/pkg/password"
)

type UserRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UserResponse struct {
	ID          int32
	Name        string
	Description string
}

type UserRoute struct {
	repo repository.UserRepository
}

func NewUserRoute(repo repository.UserRepository) *UserRoute {
	return &UserRoute{repo: repo}
}

func (u *UserRoute) GetUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	fmt.Println(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}

	user, err := u.repo.GetUserByID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userToResponse(user))
}

func (u *UserRoute) CreateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var user LoginRequest
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if user.Login == "" || user.Password == "" {
		http.Error(w, "login and password required", http.StatusBadRequest)
		return
	}

	hash, err := password.HashPassword(user.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	user.Password = hash

	createdUser, err := u.repo.CreateUser(ctx, db.CreateUserAuthParams{
		Login:        user.Login,
		PasswordHash: hash})
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userToResponse(createdUser))
}

func (u *UserRoute) UpdateUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = CheckOwnership(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	var user UserRequest
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	err = u.repo.UpdateUser(ctx, db.UpdateUserParams{
		ID:          int32(id),
		Name:        user.Name,
		Description: sql.NullString{String: user.Description, Valid: user.Description != ""},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user updated"))
}

func (u *UserRoute) DeleteUser(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	err = CheckOwnership(ctx, id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	err = u.repo.DeleteUser(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("user deleted"))
}

func CheckClaims(ctx context.Context) (*middleware.AccessClaims, error) {
	claims, ok := ctx.Value(middleware.AuthKey{}).(*middleware.AccessClaims)
	if !ok {
		return nil, errors.New("invalid claims")
	}
	return claims, nil
}

func CheckOwnership(ctx context.Context, id int) error {
	claims, ok := ctx.Value(middleware.AuthKey{}).(*middleware.AccessClaims)
	if !ok {
		return errors.New("invalid claims")
	}
	if !claims.IsAdmin && claims.ID != int32(id) {
		return errors.New("not an owner")
	}
	return nil
}

func userToResponse(user db.User) UserResponse {
	return UserResponse{
		ID:          user.ID,
		Name:        user.Name,
		Description: user.Description.String,
	}
}
