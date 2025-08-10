package routes

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"image-sharing/internal/configs"
	"image-sharing/internal/db/gen"
	"image-sharing/internal/metrics"
	midle "image-sharing/internal/middleware"
	"image-sharing/internal/repository"
	ssov1 "image-sharing/protos/gen"
)

func SetupRouter(dbConnetcion *sql.DB, config configs.Config) *chi.Mux {
	router := chi.NewRouter()
	metrics := metrics.New()
	router.Use(middleware.Logger)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.Recoverer)
	router.Use(metrics.Middleware())

	querys := db.New(dbConnetcion)

	con, err := grpc.NewClient(config.SSOAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		panic(err)
	}
	authClient := ssov1.NewAuthClient(con)
	authMiddleware := midle.GetAuthMiddleware(authClient)

	postRepository := repository.NewPostRepository(dbConnetcion, querys, config.ImagesDirectory)
	postRoute := NewPostRoute(postRepository)

	router.Get("/metrics", metrics.Handler().ServeHTTP)

	router.Route("/post", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Get("/{id}", postRoute.GetPost)
			r.Get("/", postRoute.GetPosts)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Post("/", postRoute.CreatePost)
			r.Delete("/{id}", postRoute.DeletePost)
		})
	})

	return router
}
