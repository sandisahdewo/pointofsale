package seeds

import (
	"log/slog"
	"strings"
	"time"

	"github.com/lib/pq"
	"github.com/pointofsale/backend/models"
	"github.com/pointofsale/backend/repositories"
	"github.com/pointofsale/backend/services"
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

	// 9. Seed Products
	if err := seedProducts(db); err != nil {
		return err
	}

	// 10. Seed Purchase Orders
	if err := seedPurchaseOrders(db); err != nil {
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

func seedProducts(db *gorm.DB) error {
	if !db.Migrator().HasTable("products") {
		return nil
	}

	// Lookup IDs
	var categories []models.Category
	if err := db.Find(&categories).Error; err != nil {
		return err
	}
	categoryByName := make(map[string]uint, len(categories))
	for _, category := range categories {
		categoryByName[category.Name] = category.ID
	}

	var suppliers []models.Supplier
	if err := db.Find(&suppliers).Error; err != nil {
		return err
	}
	supplierByName := make(map[string]uint, len(suppliers))
	for _, supplier := range suppliers {
		supplierByName[supplier.Name] = supplier.ID
	}

	var racks []models.Rack
	if err := db.Find(&racks).Error; err != nil {
		return err
	}
	rackByName := make(map[string]uint, len(racks))
	for _, rack := range racks {
		rackByName[rack.Name] = rack.ID
	}

	productService := services.NewProductService(repositories.NewProductRepository(db))
	markupPercentage := "percentage"

	inputs := []services.CreateProductInput{
		{
			Name:         "T-Shirt",
			Description:  "Premium cotton t-shirt available in multiple colors and sizes",
			CategoryID:   categoryByName["Clothing"],
			PriceSetting: "fixed",
			HasVariants:  true,
			Status:       "active",
			SupplierIDs: []uint{
				supplierByName["PT Sumber Makmur"],
				supplierByName["CV Jaya Abadi"],
			},
			Units: []services.CreateProductUnitInput{
				{Name: "Pcs", IsBase: true},
				{Name: "Dozen", ConversionFactor: 12, ConvertsToName: "Pcs"},
				{Name: "Box", ConversionFactor: 12, ConvertsToName: "Dozen"},
				{Name: "Bag", ConversionFactor: 50, ConvertsToName: "Pcs"},
			},
			Variants: []services.CreateProductVariantInput{
				{
					SKU:     "TS-R-S",
					Barcode: "8901234567890",
					Attributes: []services.CreateVariantAttributeInput{
						{AttributeName: "Color", AttributeValue: "Red"},
						{AttributeName: "Size", AttributeValue: "S"},
					},
					PricingTiers: []services.CreateVariantPricingTierInput{
						{MinQty: 1, Value: 75000},
						{MinQty: 12, Value: 70000},
					},
					RackIDs: []uint{rackByName["Main Display"]},
				},
				{
					SKU:     "TS-B-M",
					Barcode: "8901234567891",
					Attributes: []services.CreateVariantAttributeInput{
						{AttributeName: "Color", AttributeValue: "Blue"},
						{AttributeName: "Size", AttributeValue: "M"},
					},
					PricingTiers: []services.CreateVariantPricingTierInput{
						{MinQty: 1, Value: 75000},
						{MinQty: 12, Value: 70000},
					},
					RackIDs: []uint{rackByName["Main Display"]},
				},
			},
		},
		{
			Name:         "Rice",
			Description:  "Premium quality white rice",
			CategoryID:   categoryByName["Food & Beverages"],
			PriceSetting: "fixed",
			HasVariants:  false,
			Status:       "active",
			SupplierIDs:  []uint{supplierByName["UD Berkah Sentosa"]},
			Units: []services.CreateProductUnitInput{
				{Name: "Kg", IsBase: true},
				{Name: "Karung", ConversionFactor: 50, ConvertsToName: "Kg"},
				{Name: "Bag", ConversionFactor: 25, ConvertsToName: "Kg"},
			},
			Variants: []services.CreateProductVariantInput{
				{
					SKU:     "RC-001",
					Barcode: "8901234567800",
					PricingTiers: []services.CreateVariantPricingTierInput{
						{MinQty: 1, Value: 15000},
						{MinQty: 50, Value: 14000},
					},
					RackIDs: []uint{rackByName["Bulk Storage"]},
				},
			},
		},
		{
			Name:         "Notebook",
			Description:  "Lined notebook, A5 size",
			CategoryID:   categoryByName["Stationery"],
			PriceSetting: "markup",
			MarkupType:   &markupPercentage,
			HasVariants:  false,
			Status:       "active",
			SupplierIDs:  []uint{supplierByName["CV Jaya Abadi"]},
			Units: []services.CreateProductUnitInput{
				{Name: "Pcs", IsBase: true},
				{Name: "Carton", ConversionFactor: 48, ConvertsToName: "Pcs"},
			},
			Variants: []services.CreateProductVariantInput{
				{
					SKU:     "NB-001",
					Barcode: "8901234567700",
					PricingTiers: []services.CreateVariantPricingTierInput{
						{MinQty: 1, Value: 25},
					},
					RackIDs: []uint{rackByName["Main Display"]},
				},
			},
		},
		{
			Name:         "Cooking Oil",
			Description:  "Premium vegetable cooking oil",
			CategoryID:   categoryByName["Household"],
			PriceSetting: "fixed",
			HasVariants:  false,
			Status:       "active",
			Units: []services.CreateProductUnitInput{
				{Name: "Liter", IsBase: true},
			},
			Variants: []services.CreateProductVariantInput{
				{
					SKU:     "CO-001",
					Barcode: "8901234567600",
					PricingTiers: []services.CreateVariantPricingTierInput{
						{MinQty: 1, Value: 28000},
					},
					RackIDs: []uint{rackByName["Cold Storage"]},
				},
			},
		},
	}

	// Stock values reflect what will be set after PO seeding (received POs add stock).
	// These are the "initial" stock before PO receive.
	stockBySKU := map[string]int{
		"TS-R-S": 50,
		"TS-B-M": 25,
		"RC-001": 200,
		"NB-001": 150,
		"CO-001": 5,
	}

	for _, input := range inputs {
		var existing models.Product
		err := db.Where("name = ?", input.Name).First(&existing).Error
		if err == nil {
			continue
		}
		if err != gorm.ErrRecordNotFound {
			return err
		}

		created, serviceErr := productService.CreateProduct(input)
		if serviceErr != nil {
			return serviceErr
		}

		for sku, stock := range stockBySKU {
			if err := db.Model(&models.ProductVariant{}).
				Where("product_id = ? AND sku = ?", created.ID, sku).
				Update("current_stock", stock).Error; err != nil {
				return err
			}
		}

		slog.Info("created product", "name", input.Name)
	}

	return nil
}

func seedPurchaseOrders(db *gorm.DB) error {
	if !db.Migrator().HasTable("purchase_orders") {
		return nil
	}

	// Check if POs already exist
	var count int64
	if err := db.Model(&models.PurchaseOrder{}).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		slog.Info("purchase orders already exist, skipping seed")
		return nil
	}

	// Look up suppliers
	supplierByName := make(map[string]*models.Supplier)
	var allSuppliers []models.Supplier
	if err := db.Preload("BankAccounts").Find(&allSuppliers).Error; err != nil {
		return err
	}
	for i := range allSuppliers {
		supplierByName[allSuppliers[i].Name] = &allSuppliers[i]
	}

	// Look up products with variants and units
	productByName := make(map[string]*models.Product)
	var allProducts []models.Product
	if err := db.Preload("Units").Preload("Variants").Preload("Variants.Attributes").Find(&allProducts).Error; err != nil {
		return err
	}
	for i := range allProducts {
		productByName[allProducts[i].Name] = &allProducts[i]
	}

	findVariant := func(product *models.Product, sku string) *models.ProductVariant {
		for i := range product.Variants {
			if product.Variants[i].SKU == sku {
				return &product.Variants[i]
			}
		}
		return nil
	}

	findBaseUnit := func(product *models.Product) *models.ProductUnit {
		for i := range product.Units {
			if product.Units[i].IsBase {
				return &product.Units[i]
			}
		}
		return nil
	}

	buildLabel := func(v *models.ProductVariant) string {
		if len(v.Attributes) == 0 {
			return "Default"
		}
		labels := make([]string, len(v.Attributes))
		for i, attr := range v.Attributes {
			labels[i] = attr.AttributeValue
		}
		return strings.Join(labels, " / ")
	}

	tshirt := productByName["T-Shirt"]
	notebook := productByName["Notebook"]
	rice := productByName["Rice"]

	tshirtBaseUnit := findBaseUnit(tshirt)
	notebookBaseUnit := findBaseUnit(notebook)
	riceBaseUnit := findBaseUnit(rice)

	tsRS := findVariant(tshirt, "TS-R-S")
	tsBM := findVariant(tshirt, "TS-B-M")
	nbVariant := findVariant(notebook, "NB-001")
	rcVariant := findVariant(rice, "RC-001")

	smSupplier := supplierByName["PT Sumber Makmur"]
	jaSupplier := supplierByName["CV Jaya Abadi"]
	bsSupplier := supplierByName["UD Berkah Sentosa"]

	bankTransfer := "bank_transfer"
	cash := "cash"

	po1Subtotal := float64(50*45000 + 50*45000 + 100*20000)
	po1TotalItems := 200
	po1ReceivedDate := time.Date(2026, 2, 6, 14, 30, 0, 0, time.UTC)
	smBankID := ""
	if len(smSupplier.BankAccounts) > 0 {
		smBankID = smSupplier.BankAccounts[0].ID
	}

	po1 := models.PurchaseOrder{
		PONumber: "PO-2026-0001", SupplierID: smSupplier.ID, Date: "2026-02-05",
		Status: "completed", Notes: "First restocking order",
		ReceivedDate: &po1ReceivedDate, PaymentMethod: &bankTransfer,
		SupplierBankAccountID: &smBankID, Subtotal: &po1Subtotal, TotalItems: &po1TotalItems,
		Items: []models.PurchaseOrderItem{
			{ProductID: tshirt.ID, VariantID: tsRS.ID, UnitID: tshirtBaseUnit.ID, UnitName: tshirtBaseUnit.Name, ProductName: tshirt.Name, VariantLabel: buildLabel(tsRS), SKU: tsRS.SKU, CurrentStock: 0, OrderedQty: 50, ReceivedQty: intPtr(50), ReceivedPrice: floatPtr(45000), IsVerified: true},
			{ProductID: tshirt.ID, VariantID: tsBM.ID, UnitID: tshirtBaseUnit.ID, UnitName: tshirtBaseUnit.Name, ProductName: tshirt.Name, VariantLabel: buildLabel(tsBM), SKU: tsBM.SKU, CurrentStock: 0, OrderedQty: 50, ReceivedQty: intPtr(50), ReceivedPrice: floatPtr(45000), IsVerified: true},
			{ProductID: notebook.ID, VariantID: nbVariant.ID, UnitID: notebookBaseUnit.ID, UnitName: notebookBaseUnit.Name, ProductName: notebook.Name, VariantLabel: buildLabel(nbVariant), SKU: nbVariant.SKU, CurrentStock: 0, OrderedQty: 100, ReceivedQty: intPtr(100), ReceivedPrice: floatPtr(20000), IsVerified: true},
		},
	}

	po2Subtotal := float64(50 * 20000)
	po2TotalItems := 50
	po2ReceivedDate := time.Date(2026, 2, 9, 10, 0, 0, 0, time.UTC)

	po2 := models.PurchaseOrder{
		PONumber: "PO-2026-0002", SupplierID: jaSupplier.ID, Date: "2026-02-08",
		Status: "received", Notes: "Urgent notebook restock",
		ReceivedDate: &po2ReceivedDate, PaymentMethod: &cash,
		Subtotal: &po2Subtotal, TotalItems: &po2TotalItems,
		Items: []models.PurchaseOrderItem{
			{ProductID: notebook.ID, VariantID: nbVariant.ID, UnitID: notebookBaseUnit.ID, UnitName: notebookBaseUnit.Name, ProductName: notebook.Name, VariantLabel: buildLabel(nbVariant), SKU: nbVariant.SKU, CurrentStock: 100, OrderedQty: 50, ReceivedQty: intPtr(50), ReceivedPrice: floatPtr(20000), IsVerified: true},
		},
	}

	po3 := models.PurchaseOrder{
		PONumber: "PO-2026-0003", SupplierID: bsSupplier.ID, Date: "2026-02-10",
		Status: "sent", Notes: "Monthly rice order",
		Items: []models.PurchaseOrderItem{
			{ProductID: rice.ID, VariantID: rcVariant.ID, UnitID: riceBaseUnit.ID, UnitName: riceBaseUnit.Name, ProductName: rice.Name, VariantLabel: buildLabel(rcVariant), SKU: rcVariant.SKU, CurrentStock: 200, OrderedQty: 100},
		},
	}

	po4 := models.PurchaseOrder{
		PONumber: "PO-2026-0004", SupplierID: smSupplier.ID, Date: "2026-02-12",
		Status: "draft", Notes: "Pending review",
		Items: []models.PurchaseOrderItem{
			{ProductID: tshirt.ID, VariantID: tsRS.ID, UnitID: tshirtBaseUnit.ID, UnitName: tshirtBaseUnit.Name, ProductName: tshirt.Name, VariantLabel: buildLabel(tsRS), SKU: tsRS.SKU, CurrentStock: 50, OrderedQty: 25},
			{ProductID: tshirt.ID, VariantID: tsBM.ID, UnitID: tshirtBaseUnit.ID, UnitName: tshirtBaseUnit.Name, ProductName: tshirt.Name, VariantLabel: buildLabel(tsBM), SKU: tsBM.SKU, CurrentStock: 25, OrderedQty: 25},
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, po := range []*models.PurchaseOrder{&po1, &po2, &po3, &po4} {
			if err := tx.Create(po).Error; err != nil {
				return err
			}
			slog.Info("created purchase order", "poNumber", po.PONumber, "status", po.Status)
		}

		movements := []models.StockMovement{
			{VariantID: tsRS.ID, MovementType: "purchase_receive", Quantity: 50, ReferenceType: "purchase_order", ReferenceID: &po1.ID, Notes: "PO-2026-0001 receive"},
			{VariantID: tsBM.ID, MovementType: "purchase_receive", Quantity: 50, ReferenceType: "purchase_order", ReferenceID: &po1.ID, Notes: "PO-2026-0001 receive"},
			{VariantID: nbVariant.ID, MovementType: "purchase_receive", Quantity: 100, ReferenceType: "purchase_order", ReferenceID: &po1.ID, Notes: "PO-2026-0001 receive"},
			{VariantID: nbVariant.ID, MovementType: "purchase_receive", Quantity: 50, ReferenceType: "purchase_order", ReferenceID: &po2.ID, Notes: "PO-2026-0002 receive"},
		}

		for _, m := range movements {
			if err := tx.Create(&m).Error; err != nil {
				return err
			}
		}
		slog.Info("created stock movements for received POs")

		return nil
	})
}

func intPtr(v int) *int          { return &v }
func floatPtr(v float64) *float64 { return &v }
