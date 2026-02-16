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

	// 5. Seed Test Users
	if err := seedTestUsers(db); err != nil {
		return err
	}

	// 6. Seed Categories
	if err := seedCategories(db); err != nil {
		return err
	}

	// 7. Seed Suppliers
	if err := seedSuppliers(db); err != nil {
		return err
	}

	// 8. Seed Racks
	if err := seedRacks(db); err != nil {
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
		{Module: "Settings", Feature: "Users", Actions: pq.StringArray{"create", "read", "update", "delete"}},
		{Module: "Settings", Feature: "Roles & Permissions", Actions: pq.StringArray{"create", "read", "update", "delete"}},
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
			{module: "Settings", feature: "Users", actions: []string{"create", "read", "update"}},
			{module: "Settings", feature: "Roles & Permissions", actions: []string{"read"}},
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

func seedTestUsers(db *gorm.DB) error {
	// Hash the common test password
	hashedPassword, err := utils.HashPassword("Password@123")
	if err != nil {
		return err
	}

	// Get roles
	var managerRole, cashierRole, warehouseRole, accountantRole models.Role
	db.Where("name = ?", "Manager").First(&managerRole)
	db.Where("name = ?", "Cashier").First(&cashierRole)
	db.Where("name = ?", "Warehouse").First(&warehouseRole)
	db.Where("name = ?", "Accountant").First(&accountantRole)

	testUsers := []struct {
		user  models.User
		roles []models.Role
	}{
		{
			user: models.User{
				Name:         "Budi Santoso",
				Email:        "budi@pointofsale.com",
				PasswordHash: hashedPassword,
				Phone:        "+62-812-0000-0002",
				Status:       "active",
				IsSuperAdmin: false,
			},
			roles: []models.Role{managerRole},
		},
		{
			user: models.User{
				Name:         "Siti Rahayu",
				Email:        "siti@pointofsale.com",
				PasswordHash: hashedPassword,
				Phone:        "+62-812-0000-0003",
				Status:       "active",
				IsSuperAdmin: false,
			},
			roles: []models.Role{cashierRole},
		},
		{
			user: models.User{
				Name:         "Ahmad Wijaya",
				Email:        "ahmad@pointofsale.com",
				PasswordHash: hashedPassword,
				Phone:        "+62-812-0000-0004",
				Status:       "active",
				IsSuperAdmin: false,
			},
			roles: []models.Role{warehouseRole, accountantRole},
		},
		{
			user: models.User{
				Name:         "Dewi Lestari",
				Email:        "dewi@pointofsale.com",
				PasswordHash: hashedPassword,
				Phone:        "+62-812-0000-0005",
				Status:       "inactive",
				IsSuperAdmin: false,
			},
			roles: []models.Role{cashierRole},
		},
		{
			user: models.User{
				Name:         "Rizky Pratama",
				Email:        "rizky@pointofsale.com",
				PasswordHash: hashedPassword,
				Phone:        "+62-812-0000-0006",
				Status:       "pending",
				IsSuperAdmin: false,
			},
			roles: []models.Role{},
		},
	}

	for _, tu := range testUsers {
		var existingUser models.User
		if err := db.Where("email = ?", tu.user.Email).First(&existingUser).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&tu.user).Error; err != nil {
					return err
				}

				// Assign roles
				if len(tu.roles) > 0 {
					if err := db.Model(&tu.user).Association("Roles").Append(tu.roles); err != nil {
						return err
					}
				}

				slog.Info("created test user", "email", tu.user.Email)
			} else {
				return err
			}
		}
	}

	return nil
}

func seedCategories(db *gorm.DB) error {
	categories := []models.Category{
		{Name: "Clothing", Description: "Apparel and garments"},
		{Name: "Food & Beverages", Description: "Food items and drinks"},
		{Name: "Stationery", Description: "Office and school supplies"},
		{Name: "Household", Description: "Home and kitchen essentials"},
	}

	for _, category := range categories {
		var existing models.Category
		if err := db.Where("name = ?", category.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Create(&category).Error; err != nil {
					return err
				}
				slog.Info("created category", "name", category.Name)
			} else {
				return err
			}
		}
	}

	return nil
}

func seedSuppliers(db *gorm.DB) error {
	suppliers := []struct {
		supplier     models.Supplier
		bankAccounts []models.SupplierBankAccount
	}{
		{
			supplier: models.Supplier{
				Name:    "PT Sumber Makmur",
				Address: "Jl. Industri No. 45, Jakarta",
				Phone:   "+62-21-5550001",
				Email:   "order@sumbermakmur.co.id",
				Website: "sumbermakmur.co.id",
				Active:  true,
			},
			bankAccounts: []models.SupplierBankAccount{
				{AccountName: "BCA - Main Account", AccountNumber: "1234567890"},
				{AccountName: "Mandiri - Operations", AccountNumber: "0987654321"},
			},
		},
		{
			supplier: models.Supplier{
				Name:    "CV Jaya Abadi",
				Address: "Jl. Perdagangan No. 12, Surabaya",
				Phone:   "+62-31-5550002",
				Email:   "sales@jayaabadi.com",
				Active:  true,
			},
			bankAccounts: []models.SupplierBankAccount{
				{AccountName: "BCA - Main Account", AccountNumber: "1122334455"},
			},
		},
		{
			supplier: models.Supplier{
				Name:    "UD Berkah Sentosa",
				Address: "Jl. Pasar Baru No. 8, Bandung",
				Active:  true,
			},
			bankAccounts: []models.SupplierBankAccount{},
		},
		{
			supplier: models.Supplier{
				Name:    "PT Global Supplies",
				Address: "Jl. Raya Serpong No. 100, Tangerang",
				Phone:   "+62-21-5550004",
				Email:   "info@globalsupplies.co.id",
				Website: "globalsupplies.co.id",
				Active:  false,
			},
			bankAccounts: []models.SupplierBankAccount{
				{AccountName: "BNI - Main Account", AccountNumber: "5566778899"},
				{AccountName: "BRI - Operations", AccountNumber: "9988776655"},
			},
		},
	}

	for _, s := range suppliers {
		var existing models.Supplier
		if err := db.Where("name = ?", s.supplier.Name).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create supplier in transaction
				if err := db.Transaction(func(tx *gorm.DB) error {
					// Create supplier
					if err := tx.Select("*").Create(&s.supplier).Error; err != nil {
						return err
					}

					// Create bank accounts
					for _, ba := range s.bankAccounts {
						ba.SupplierID = s.supplier.ID
						if err := tx.Create(&ba).Error; err != nil {
							return err
						}
					}

					return nil
				}); err != nil {
					return err
				}
				slog.Info("created supplier", "name", s.supplier.Name, "bank_accounts", len(s.bankAccounts))
			} else {
				return err
			}
		}
	}

	return nil
}

func seedRacks(db *gorm.DB) error {
	racks := []models.Rack{
		{Name: "Main Display", Code: "R-001", Location: "Store Front", Capacity: 100, Description: "Primary display shelf near entrance", Active: true},
		{Name: "Electronics Shelf", Code: "R-002", Location: "Store Front", Capacity: 50, Description: "Dedicated electronics display", Active: true},
		{Name: "Cold Storage", Code: "R-003", Location: "Warehouse Zone A", Capacity: 200, Description: "Refrigerated storage area", Active: true},
		{Name: "Bulk Storage", Code: "R-004", Location: "Warehouse Zone B", Capacity: 500, Description: "Large item storage", Active: true},
		{Name: "Clearance Rack", Code: "R-005", Location: "Store Back", Capacity: 30, Description: "Discounted items", Active: false},
	}

	for _, rack := range racks {
		var existing models.Rack
		if err := db.Where("code = ?", rack.Code).First(&existing).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				if err := db.Select("*").Create(&rack).Error; err != nil {
					return err
				}
				slog.Info("created rack", "code", rack.Code, "name", rack.Name)
			} else {
				return err
			}
		}
	}

	return nil
}
