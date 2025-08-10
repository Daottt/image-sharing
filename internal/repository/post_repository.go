package repository

import (
	"context"
	"database/sql"
	"errors"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"image-sharing/internal/db/gen"
)

var ErrNotFound = errors.New("not found")

type PostRepository interface {
	GetPostByID(ctx context.Context, id int) (db.GetPostRow, error)
	GetAllPosts(ctx context.Context, page int, limit int) ([]db.ListPostsRow, int64, error)
	CreatePost(ctx context.Context, post db.CreatePostParams, file multipart.File, format string) (db.Post, error)
	GetPostUserID(ctx context.Context, id int) (int32, error)
	DeletePost(ctx context.Context, id int) error
}

type postRepository struct {
	db        *sql.DB
	queries   *db.Queries
	uploadDir string
}

func NewPostRepository(db *sql.DB, queries *db.Queries, uploadDir string) PostRepository {
	return &postRepository{db: db, queries: queries, uploadDir: uploadDir}
}

func (r *postRepository) GetPostByID(ctx context.Context, id int) (db.GetPostRow, error) {
	post, err := r.queries.GetPost(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return db.GetPostRow{}, ErrNotFound
		}
		return db.GetPostRow{}, err
	}
	return post, nil
}

func (r *postRepository) GetAllPosts(ctx context.Context, page int, limit int) ([]db.ListPostsRow, int64, error) {
	offset := (page - 1) * limit

	posts, err := r.queries.ListPosts(ctx, db.ListPostsParams{Limit: int32(limit), Offset: int32(offset)})
	if err != nil {
		return nil, 0, err
	}

	count, err := r.queries.CountPosts(ctx)
	if err != nil {
		return nil, 0, err
	}

	return posts, count, nil
}

func (r *postRepository) CreatePost(ctx context.Context, post db.CreatePostParams, file multipart.File, format string) (db.Post, error) {

	newFilename := uuid.New().String() + format
	uploadPath := filepath.Join(r.uploadDir, newFilename)

	dst, err := os.Create(uploadPath)
	if err != nil {
		return db.Post{}, err
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		os.Remove(uploadPath)
		return db.Post{}, err
	}

	post.ImagePath = uploadPath
	createdPost, err := r.queries.CreatePost(ctx, post)
	if err != nil {
		os.Remove(uploadPath)
		return db.Post{}, err
	}

	return createdPost, nil
}

func (r *postRepository) GetPostUserID(ctx context.Context, id int) (int32, error) {
	post, err := r.queries.GetPostUserID(ctx, int32(id))
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, ErrNotFound
		}
		return 0, err
	}
	return post, nil
}
func (r *postRepository) DeletePost(ctx context.Context, id int) error {
	return r.queries.DeletPost(ctx, int32(id))
}
