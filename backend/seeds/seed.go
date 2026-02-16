package seeds

import (
	"log/slog"

	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	var count int64
	db.Model(&models.User{}).Count(&count)
	if count > 0 {
		slog.Info("Seed data already exists, skipping")
		return nil
	}

	slog.Info("Seeding database...")

	hashedPassword, err := utils.HashPassword("password123")
	if err != nil {
		return err
	}

	users := []models.User{
		{Name: "Admin", Email: "admin@pointofsale.local", Password: hashedPassword, Role: "admin", IsActive: true},
	}

	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			return err
		}
	}

	slog.Info("Database seeded successfully")
	return nil
}
