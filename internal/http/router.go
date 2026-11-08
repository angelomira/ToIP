package router

import (
	"net/http"
	"example.com/pz10-auth/internal/core"
	"example.com/pz10-auth/internal/http/middleware"
	"example.com/pz10-auth/internal/platform/config"
	"example.com/pz10-auth/internal/platform/jwt"
	"example.com/pz10-auth/internal/repo"
	"github.com/go-chi/chi/v5"
)

func Build(cfg config.Config) http.Handler {
	r := chi.NewRouter()

	userRepo := repo.NewUserMem()
	jwtw := jwt.NewHS256(cfg.JWTSecret)
	svc := core.NewService(userRepo, jwtw, cfg)

	r.Post("/api/v1/login", svc.LoginHandler)
	r.Post("/api/v1/refresh", svc.RefreshHandler) // Эндпоинт обновления

	r.Group(func(priv chi.Router) {
		priv.Use(middleware.AuthN(jwtw))
		priv.Use(middleware.AuthZRoles("admin", "user"))
		priv.Get("/api/v1/me", svc.MeHandler)
		
		// Задание ABAC
		priv.With(middleware.AuthZABAC()).Get("/api/v1/users/{id}", svc.UserProfile)
	})

	r.Group(func(admin chi.Router) {
		admin.Use(middleware.AuthN(jwtw))
		admin.Use(middleware.AuthZRoles("admin"))
		admin.Get("/api/v1/admin/stats", svc.AdminStats)
	})

	return r
}