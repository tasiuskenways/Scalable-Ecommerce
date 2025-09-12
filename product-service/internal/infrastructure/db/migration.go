package db

import (
	"fmt"

	"gorm.io/gorm"
	"tasius.my.id/SE/product-service/internal/domain/entities"
)

func Migrate(db *gorm.DB, resetDb bool) error {
	// Enable UUID extension
	err := db.Exec("CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\"").Error
	if err != nil {
		return fmt.Errorf("failed to create uuid extension: %w", err)
	}

	if resetDb {
		// Drop existing tables if they exist
		err = db.Migrator().DropTable(&entities.Product{}, &entities.Category{})
		if err != nil {
			return fmt.Errorf("failed to drop tables: %w", err)
		}
	}

	// Create tables with new schema
	return db.AutoMigrate(
		&entities.Category{},
		&entities.Product{},
	)
}
