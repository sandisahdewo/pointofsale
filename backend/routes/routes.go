package routes

import (
	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/handlers"
	"github.com/pointofsale/backend/middleware"
)

func Setup(r chi.Router, healthHandler *handlers.HealthHandler, authHandler *handlers.AuthHandler, cfg *config.Config) {
	// Global middleware
	r.Use(chiMiddleware.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{cfg.FrontendURL},
		AllowedMethods:   []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-Request-ID"},
		ExposedHeaders:   []string{"X-Request-ID"},
		AllowCredentials: true,
		MaxAge:           300,
	}))

	// Health check (no auth required)
	r.Get("/health", healthHandler.Health)

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/logout", authHandler.Logout)
		})

		// Protected routes (require auth)
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(cfg.JWTAccessSecret))
			// Future protected endpoints go here
		})
	})
}
