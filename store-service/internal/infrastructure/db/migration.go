package db

import (
	"fmt"

	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
	"gorm.io/gorm"
)

func Migrate(db *gorm.DB, resetDb bool) error {
	// Enable UUID extension
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	if resetDb {
		// Drop existing tables if they exist
		err = db.Migrator().DropTable(&entities.UserStoreRole{}, &entities.StoreInvitation{}, &entities.Store{})
		if err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	// Create tables with new schema
	err = db.AutoMigrate(
		&entities.UserStoreRole{},
		&entities.StoreInvitation{},
		&entities.Store{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate tables: %w", err)
	}

	// Migrate existing stores to have proper settings
	err = migrateStoreSettings(db)
	if err != nil {
		return fmt.Errorf("failed to migrate store settings: %w", err)
	}

	return nil
}

// migrateStoreSettings ensures all stores have proper JSONB settings
func migrateStoreSettings(db *gorm.DB) error {
	// Update stores with null or empty settings to have default settings
	defaultSettings := entities.GetDefaultStoreSettings()
	settingsJSON, err := defaultSettings.Value()
	if err != nil {
		return fmt.Errorf("failed to marshal default settings: %w", err)
	}

	// Update stores where settings is null or empty
	err = db.Exec("UPDATE stores SET settings = ? WHERE settings IS NULL OR settings = '{}'", settingsJSON).Error
	if err != nil {
		return fmt.Errorf("failed to update store settings: %w", err)
	}

	return nil
}
