package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/pointofsale/backend/utils"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db  *gorm.DB
	rdb *redis.Client
}

func NewHealthHandler(db *gorm.DB, rdb *redis.Client) *HealthHandler {
	return &HealthHandler{db: db, rdb: rdb}
}

func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
	defer cancel()

	// Check database
	sqlDB, err := h.db.DB()
	if err != nil {
		utils.Error(w, http.StatusServiceUnavailable, "database connection error")
		return
	}
	if err := sqlDB.PingContext(ctx); err != nil {
		utils.Error(w, http.StatusServiceUnavailable, "database ping failed")
		return
	}

	// Check Redis
	if err := h.rdb.Ping(ctx).Err(); err != nil {
		utils.Error(w, http.StatusServiceUnavailable, "redis ping failed")
		return
	}

	utils.Success(w, http.StatusOK, "healthy", map[string]string{
		"status":   "ok",
		"database": "connected",
		"redis":    "connected",
	})
}
