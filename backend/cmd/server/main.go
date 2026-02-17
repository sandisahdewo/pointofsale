package main

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/go-chi/chi/v5"
	"github.com/pressly/goose/v3"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/pointofsale/backend/config"
	"github.com/pointofsale/backend/handlers"
	"github.com/pointofsale/backend/middleware"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/routes"
	"github.com/pointofsale/backend/seeds"
	"github.com/pointofsale/backend/services"
	"github.com/pointofsale/backend/storage"
	"github.com/pointofsale/backend/utils"
)

// userEmailAdapter adapts utils.EmailService to services.UserEmailService interface
type userEmailAdapter struct {
	svc *utils.EmailService
}

func (a *userEmailAdapter) SendUserCredentials(toEmail, userName, tempPassword string) error {
	return a.svc.SendCredentialsEmail(toEmail, userName, tempPassword)
}

func (a *userEmailAdapter) SendUserApproved(toEmail, userName string) error {
	return a.svc.SendAccountApprovedEmail(toEmail, userName)
}

func (a *userEmailAdapter) SendUserRejected(toEmail, userName string) error {
	return a.svc.SendRejectionEmail(toEmail, userName)
}

func main() {
	// Setup structured logging
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Load config
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	// Connect to PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.DSN()), &gorm.Config{})
	if err != nil {
		slog.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to PostgreSQL")

	// Run goose migrations
	sqlDB, err := db.DB()
	if err != nil {
		slog.Error("failed to get sql.DB", "error", err)
		os.Exit(1)
	}
	if err := runMigrations(sqlDB); err != nil {
		slog.Error("failed to run migrations", "error", err)
		os.Exit(1)
	}

	// Run seeds
	if err := seeds.Run(db); err != nil {
		slog.Error("failed to seed database", "error", err)
		os.Exit(1)
	}

	// Connect to Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", cfg.RedisHost, cfg.RedisPort),
		Password: cfg.RedisPassword,
		DB:       0,
	})
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		slog.Error("failed to connect to Redis", "error", err)
		os.Exit(1)
	}
	slog.Info("connected to Redis")

	// Initialize email service
	emailService := utils.NewEmailService(cfg.SMTPHost, cfg.SMTPPort, cfg.SMTPFrom)

	// Initialize repositories
	userRepo := repositories.NewUserRepository(db)
	roleRepo := repositories.NewRoleRepository(db)
	categoryRepo := repositories.NewCategoryRepository(db)
	supplierRepo := repositories.NewSupplierRepository(db)
	rackRepo := repositories.NewRackRepository(db)
	productRepo := repositories.NewProductRepository(db)
	poRepo := repositories.NewPORepository(db)
	stockMovementRepo := repositories.NewStockMovementRepository(db)
	salesRepo := repositories.NewSalesRepository(db)

	var imageStorage services.ImageStorage
	if cfg.MinIOEnabled {
		minioStorage, err := storage.NewMinIOImageStorage(storage.MinIOConfig{
			Endpoint:  cfg.MinIOEndpoint,
			AccessKey: cfg.MinIOAccessKey,
			SecretKey: cfg.MinIOSecretKey,
			Bucket:    cfg.MinIOBucket,
			UseSSL:    cfg.MinIOUseSSL,
			PublicURL: cfg.MinIOPublicURL,
		})
		if err != nil {
			slog.Error("failed to initialize MinIO storage", "error", err)
			os.Exit(1)
		}
		imageStorage = minioStorage
		slog.Info("MinIO storage initialized", "endpoint", cfg.MinIOEndpoint, "bucket", cfg.MinIOBucket)
	}

	// Initialize services
	authService := services.NewAuthService(userRepo, rdb, cfg, emailService)
	userEmailSvc := &userEmailAdapter{svc: emailService}
	userService := services.NewUserService(userRepo, rdb, cfg, userEmailSvc)
	roleService := services.NewRoleService(roleRepo)
	categoryService := services.NewCategoryService(categoryRepo)
	supplierService := services.NewSupplierService(supplierRepo)
	rackService := services.NewRackService(rackRepo)
	productService := services.NewProductService(productRepo, imageStorage)
	seqService := services.NewSequenceService(db)
	poService := services.NewPOService(db, poRepo, stockMovementRepo, seqService)
	salesService := services.NewSalesService(db, salesRepo, seqService)

	// Initialize middleware
	authMiddleware := middleware.NewAuthMiddleware(cfg.JWTAccessSecret, rdb, userRepo)
	permMiddleware := middleware.NewPermissionMiddleware(db, rdb)

	// Initialize handlers
	healthHandler := handlers.NewHealthHandler(db, rdb)
	authHandler := handlers.NewAuthHandler(authService)
	userHandler := handlers.NewUserHandler(userService)
	roleHandler := handlers.NewRoleHandler(roleService)
	permissionHandler := handlers.NewPermissionHandler(db, rdb)
	categoryHandler := handlers.NewCategoryHandler(categoryService)
	supplierHandler := handlers.NewSupplierHandler(supplierService)
	rackHandler := handlers.NewRackHandler(rackService)
	productHandler := handlers.NewProductHandler(productService)
	poHandler := handlers.NewPOHandler(poService)
	salesHandler := handlers.NewSalesHandler(salesService)

	// Setup router and routes
	r := chi.NewRouter()
	routes.Setup(r, healthHandler, authHandler, userHandler, roleHandler, permissionHandler, categoryHandler, supplierHandler, rackHandler, productHandler, poHandler, salesHandler, authMiddleware, permMiddleware, cfg)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	slog.Info("starting server", "address", addr, "env", cfg.AppEnv)
	if err := http.ListenAndServe(addr, r); err != nil {
		slog.Error("server failed", "error", err)
		os.Exit(1)
	}
}

func runMigrations(db *sql.DB) error {
	goose.SetDialect("postgres")
	if err := goose.Up(db, "migrations"); err != nil {
		return fmt.Errorf("goose up: %w", err)
	}
	slog.Info("migrations completed")
	return nil
}
