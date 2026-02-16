package seeds

import (
	"log/slog"

	"github.com/lib/pq"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/utils"
	"gorm.io/gorm"
)

func Run(db *gorm.DB) error {
	slog.Info("Seeding database...")

	// 1. Seed Permissions
	if err := seedPermissions(db); err != nil {
		return err
	}

	// 2. Seed Roles
	if err := seedRoles(db); err != nil {
		return err
	}

	// 3. Seed Role Permissions
	if err := seedRolePermissions(db); err != nil {
		return err
	}

	// 4. Seed Super Admin User
	if err := seedSuperAdminUser(db); err != nil {
		return err
	}

	slog.Info("Database seeded successfully")
	return nil
}

func seedPermissions(db *gorm.DB) error {
	permissions := []models.Permission{
		{Module: "Master Data", Feature: "Category", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Master Data", Feature: "Supplier", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Master Data", Feature: "Rack", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Master Data", Feature: "Product", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Transaction", Feature: "Purchase Order", Actions: pq.StringArray{"create", "read", "update", "delete", "send", "receive"}},
		{Module: "Transaction", Feature: "Sale", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Transaction", Feature: "Stock Adjustment", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Settings", Feature: "User", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Settings", Feature: "Role", Actions: pq.StringArray{"create", "read", "update", "delete"}},
	}

	for _, perm := range permissions {
		var existing models.Permission
		if err := db.Where("module = ? AND feature = ?", perm.Module, perm.Feature).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&perm).Error; err != nil {
					return err
				}
				slog.Info("created permission", "module", perm.Module, "feature", perm.Feature)
			} else {
				return err
			}
		}
	}

	return nil
}

func seedRoles(db *gorm.DB) error {
	roles := []models.Role{
		{Name: "Super Admin", Description: "Full system access", IsSystem: true},
		{Name: "Manager", Description: "Store manager with full operational access", IsSystem: true},
		{Name: "Cashier", Description: "Can process sales and view products", IsSystem: true},
		{Name: "Accountant", Description: "Can view financial reports and transactions", IsSystem: true},
		{Name: "Warehouse", Description: "Can manage inventory and purchase orders", IsSystem: true},
	}

	for _, role := range roles {
		var existing models.Role
		if err := db.Where("name = ?", role.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&role).Error; err != nil {
					return err
				}
				slog.Info("created role", "name", role.Name)
			} else {
				return err
			}
		}
	}

	return nil
}

func seedRolePermissions(db *gorm.DB) error {
	// Define role permissions (excluding Super Admin - handled by is_super_admin flag)
	rolePerms := map[string][]struct {
		module  string
		feature string
		actions []string
	}{
		"Manager": {
			{module: "Master Data", feature: "Category", actions: []string{"create", "read", "update", "delete"}},
			{module: "Master Data", feature: "Supplier", actions: []string{"create", "read", "update", "delete"}},
			{module: "Master Data", feature: "Rack", actions: []string{"create", "read", "update", "delete"}},
			{module: "Master Data", feature: "Product", actions: []string{"create", "read", "update", "delete"}},
			{module: "Transaction", feature: "Purchase Order", actions: []string{"create", "read", "update", "delete", "send", "receive"}},
			{module: "Transaction", feature: "Sale", actions: []string{"create", "read", "update", "delete"}},
			{module: "Transaction", feature: "Stock Adjustment", actions: []string{"create", "read", "update", "delete"}},
			{module: "Settings", feature: "User", actions: []string{"create", "read", "update"}},
			{module: "Settings", feature: "Role", actions: []string{"read"}},
		},
		"Cashier": {
			{module: "Master Data", feature: "Product", actions: []string{"read"}},
			{module: "Transaction", feature: "Sale", actions: []string{"create", "read"}},
		},
		"Accountant": {
			{module: "Transaction", feature: "Purchase Order", actions: []string{"read"}},
			{module: "Transaction", feature: "Sale", actions: []string{"read"}},
			{module: "Transaction", feature: "Stock Adjustment", actions: []string{"read"}},
		},
		"Warehouse": {
			{module: "Master Data", feature: "Product", actions: []string{"read", "update"}},
			{module: "Transaction", feature: "Purchase Order", actions: []string{"read", "receive"}},
			{module: "Transaction", feature: "Stock Adjustment", actions: []string{"create", "read"}},
		},
	}

	for roleName, perms := range rolePerms {
		var role models.Role
		if err := db.Where("name = ?", roleName).First(&role).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				slog.Warn("role not found, skipping permissions", "role", roleName)
				continue
			}
			return err
		}

		for _, p := range perms {
			var permission models.Permission
			if err := db.Where("module = ? AND feature = ?", p.module, p.feature).First(&permission).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					slog.Warn("permission not found", "module", p.module, "feature", p.feature)
					continue
				}
				return err
			}

			var existing models.RolePermission
			if err := db.Where("role_id = ? AND permission_id = ?", role.ID, permission.ID).First(&existing).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					rolePerm := models.RolePermission{
						RoleID:       role.ID,
						PermissionID: permission.ID,
						Actions:      pq.StringArray(p.actions),
					}
					if err := db.Create(&rolePerm).Error; err != nil {
						return err
					}
					slog.Info("created role permission", "role", roleName, "permission", p.feature)
				} else {
					return err
				}
			}
		}
	}

	return nil
}

func seedSuperAdminUser(db *gorm.DB) error {
	var existingUser models.User
	if err := db.Where("email = ?", "admin@pointofsale.com").First(&existingUser).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			hashedPassword, err := utils.HashPassword("Admin@12345")
			if err != nil {
				return err
			}

			var superAdminRole models.Role
			if err := db.Where("name = ?", "Super Admin").First(&superAdminRole).Error; err != nil {
				return err
			}

			user := models.User{
				Name:         "Super Admin",
				Email:        "admin@pointofsale.com",
				PasswordHash: hashedPassword,
				Phone:        "+62-812-0000-0001",
				Status:       "active",
				IsSuperAdmin: true,
			}

			if err := db.Create(&user).Error; err != nil {
				return err
			}

			// Assign Super Admin role
			if err := db.Model(&user).Association("Roles").Append(&superAdminRole); err != nil {
				return err
			}

			slog.Info("created super admin user", "email", user.Email)
		} else {
			return err
		}
	} else {
		slog.Info("super admin user already exists, skipping")
	}

	return nil
}
