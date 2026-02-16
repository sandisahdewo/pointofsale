package routes

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/handlers"
	"github.com/pointofsale/backend/middleware"
)

func Setup(
	r chi.Router,
	healthHandler *handlers.HealthHandler,
	authHandler *handlers.AuthHandler,
	userHandler *handlers.UserHandler,
	roleHandler *handlers.RoleHandler,
	permissionHandler *handlers.PermissionHandler,
	authMiddleware *middleware.AuthMiddleware,
	permMiddleware *middleware.PermissionMiddleware,
	cfg *config.Config,
) {
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

	// Serve uploaded files
	fileServer := http.FileServer(http.Dir("uploads"))
	r.Handle("/uploads/*", http.StripPrefix("/uploads/", fileServer))

	// API v1 routes
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (public)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", authHandler.Register)
			r.Post("/login", authHandler.Login)
			r.Post("/refresh", authHandler.Refresh)
			r.Post("/forgot-password", authHandler.ForgotPassword)
			r.Post("/reset-password", authHandler.ResetPassword)

			// Protected auth routes
			r.Group(func(r chi.Router) {
				r.Use(authMiddleware.Authenticate)
				r.Post("/logout", authHandler.Logout)
				r.Get("/me", authHandler.GetMe)
			})
		})

		// Protected routes (require auth)
		r.Group(func(r chi.Router) {
			r.Use(authMiddleware.Authenticate)

			// User management
			r.Route("/users", func(r chi.Router) {
				r.With(permMiddleware.RequirePermission("Settings", "Users", "read")).Get("/", userHandler.ListUsers)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "read")).Get("/{id}", userHandler.GetUser)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "create")).Post("/", userHandler.CreateUser)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "update")).Put("/{id}", userHandler.UpdateUser)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "delete")).Delete("/{id}", userHandler.DeleteUser)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "update")).Patch("/{id}/approve", userHandler.ApproveUser)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "delete")).Delete("/{id}/reject", userHandler.RejectUser)
				r.With(permMiddleware.RequirePermission("Settings", "Users", "update")).Post("/{id}/profile-picture", userHandler.UploadProfilePicture)
			})

			// Role management
			r.Route("/roles", func(r chi.Router) {
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "read")).Get("/", roleHandler.ListRoles)
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "read")).Get("/{id}", roleHandler.GetRole)
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "create")).Post("/", roleHandler.CreateRole)
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "update")).Put("/{id}", roleHandler.UpdateRole)
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "delete")).Delete("/{id}", roleHandler.DeleteRole)

				// Role permissions
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "read")).Get("/{id}/permissions", permissionHandler.GetRolePermissions)
				r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "update")).Put("/{id}/permissions", permissionHandler.UpdateRolePermissions)
			})

			// Permissions
			r.With(permMiddleware.RequirePermission("Settings", "Roles & Permissions", "read")).Get("/permissions", permissionHandler.ListPermissions)
		})
	})
}
