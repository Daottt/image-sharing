package routes

import (
	"database/sql"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"image-sharing/pkg/token"

	"image-sharing/internal/configs"
	"image-sharing/internal/db/gen"
	midle "image-sharing/internal/middleware"
	"image-sharing/internal/repository"
)

func SetupRouter(dbConnetcion *sql.DB, config configs.Config) *chi.Mux {
	router := chi.NewRouter()
	router.Use(middleware.Logger)
	router.Use(middleware.StripSlashes)
	router.Use(middleware.Recoverer)

	querys := db.New(dbConnetcion)

	tokenMaker := token.NewJWTMaker(config.SecretKey)

	authMiddleware := midle.GetAuthMiddleware(tokenMaker)

	authRepository := repository.NewAuthRepository(dbConnetcion, querys)
	authRoute := NewAuthRoute(authRepository, tokenMaker)

	uerRepository := repository.NewUserRepository(dbConnetcion, querys)
	userRoute := NewUserRoute(uerRepository)

	postRepository := repository.NewPostRepository(dbConnetcion, querys, config.ImagesDirectory)
	postRoute := NewPostRoute(postRepository)

	router.Route("/user", func(r chi.Router) {
		r.Group(func(r chi.Router) {
			r.Get("/{id}", userRoute.GetUser)
			r.Post("/", userRoute.CreateUser)
			r.Post("/login", authRoute.LoginUser)
		})
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware)
			r.Put("/{id}", userRoute.UpdateUser)
			r.Delete("/{id}", userRoute.DeleteUser)
			r.Post("/logout", authRoute.LogoutUser)
		})
	})

	router.Route("/token", func(r chi.Router) {
		r.Post("/renew", authRoute.RenewAccessToken)
		r.With(authMiddleware).Post("/revoke", authRoute.RevokeSessions)
	})

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
