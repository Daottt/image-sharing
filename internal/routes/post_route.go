package routes

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"strconv"

	db "image-sharing/internal/db/gen"
	"image-sharing/internal/repository"

	"github.com/go-chi/chi/v5"
)

const maxUploadSize = 500 << 20
const standartPostLimit = 10
const minLimit = 5
const maxLimit = 50

var allowedFileFormats = map[string]string{"image/png": ".png", "image/jpeg": ".jpeg", "image/gif": ".gif", "video/mp4": ".mp4", "video/webm": ".webm"}

type PostResponse struct {
	ID       int32  `json:"post_id"`
	UserID   int32  `json:"user_id"`
	UserName string `json:"user_name"`
}
type PaginatedPostResponse struct {
	TotalCount int            `json:"total_count"`
	Page       int            `json:"page"`
	Limit      int            `json:"limit"`
	TotalPages int            `json:"total_pages"`
	Posts      []PostResponse `json:"posts"`
}

type PostRoute struct {
	repo repository.PostRepository
}

func NewPostRoute(repo repository.PostRepository) *PostRoute {
	return &PostRoute{repo: repo}
}

func (p *PostRoute) GetPost(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	post, err := p.repo.GetPostByID(r.Context(), id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "user not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	http.ServeFile(w, r, post.ImagePath)
}

func (p *PostRoute) GetPosts(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	limitStr := r.URL.Query().Get("limit")

	page := 1
	if pageStr != "" {
		pa, err := strconv.Atoi(pageStr)
		if err != nil || pa < 1 {
			http.Error(w, "invalid page number", http.StatusBadRequest)
			return
		}
		page = pa
	}
	limit := standartPostLimit
	if limitStr != "" {
		l, err := strconv.Atoi(limitStr)
		if err != nil || l < minLimit || l > maxLimit {
			http.Error(w, "invalid limit value", http.StatusBadRequest)
			return
		}
		limit = l
	}

	posts, count, err := p.repo.GetAllPosts(r.Context(), page, limit)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	result := make([]PostResponse, len(posts))
	for i, j := range posts {
		result[i] = PostResponse{
			ID:       j.ID,
			UserID:   j.UserID,
			UserName: j.UserName,
		}
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(PaginatedPostResponse{
		TotalCount: int(count),
		Page:       page,
		Limit:      limit,
		TotalPages: (int(count) + limit - 1) / limit,
		Posts:      result,
	})
}

func (p *PostRoute) CreatePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	claims, err := CheckClaims(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	file, _, err := r.FormFile("post")
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer file.Close()

	format, err := isAllowedFileFormat(file)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	post, err := p.repo.CreatePost(ctx, db.CreatePostParams{UserID: claims.ID, ImagePath: ""}, file, format)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(fmt.Sprintf("post id:%v", post.ID)))
}

func (p *PostRoute) DeletePost(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id, err := strconv.Atoi(chi.URLParam(r, "id"))
	if err != nil {
		http.Error(w, "Invalid ID", http.StatusBadRequest)
		return
	}

	claims, err := CheckClaims(ctx)
	if err != nil {
		http.Error(w, err.Error(), http.StatusForbidden)
		return
	}

	userID, err := p.repo.GetPostUserID(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	if !claims.IsAdmin && claims.ID != userID {
		http.Error(w, "not an owner", http.StatusForbidden)
		return
	}

	err = p.repo.DeletePost(ctx, id)
	if err != nil {
		if err == repository.ErrNotFound {
			http.Error(w, "post not found", http.StatusNotFound)
			return
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("post deleted"))
}

func isAllowedFileFormat(file multipart.File) (string, error) {
	buffer := make([]byte, 512)
	if _, err := file.Read(buffer); err != nil {
		return "", err
	}

	_, err := file.Seek(0, io.SeekStart)
	if err != nil {
		return "", err
	}

	contentType := http.DetectContentType(buffer)
	format, ok := allowedFileFormats[contentType]
	if !ok {
		return "", errors.New("not allowed file format")
	}
	return format, nil
}
