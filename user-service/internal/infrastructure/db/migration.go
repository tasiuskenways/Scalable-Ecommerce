package db

import (
	"fmt"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
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
		err = db.Migrator().DropTable(&entities.User{})
		if err != nil {
			return fmt.Errorf("failed to drop users table: %w", err)
		}
	}

	// Create tables with new schema
	return db.AutoMigrate(
		&entities.User{},
	)
}
