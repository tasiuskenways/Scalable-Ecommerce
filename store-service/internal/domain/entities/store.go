package entities

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"
)

type Store struct {
	ID          string         `json:"id" gorm:"type:uuid;default:uuid_generate_v4();primaryKey"`
	Name        string         `json:"name" gorm:"not null;size:100"`
	Slug        string         `json:"slug" gorm:"not null;uniqueIndex;size:100"`
	Description string         `json:"description" gorm:"type:text"`
	Logo        string         `json:"logo,omitempty"`
	Banner      string         `json:"banner,omitempty"`
	Website     string         `json:"website,omitempty"`
	Phone       string         `json:"phone,omitempty"`
	Email       string         `json:"email,omitempty"`
	Address     string         `json:"address,omitempty"`
	City        string         `json:"city,omitempty"`
	State       string         `json:"state,omitempty"`
	Country     string         `json:"country,omitempty"`
	PostalCode  string         `json:"postal_code,omitempty"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	Settings    StoreSettings  `json:"settings" gorm:"type:jsonb"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`

	// Relationships
	Members []UserStoreRole `json:"members,omitempty" gorm:"foreignKey:StoreID"`
}

type StoreSettings struct {
	Currency           string `json:"currency"`
	Timezone           string `json:"timezone"`
	AllowPublicListing bool   `json:"allow_public_listing"`
	RequireApproval    bool   `json:"require_approval"`
	MaxProducts        int    `json:"max_products"`
}

// Value implements driver.Valuer interface for database storage
func (s StoreSettings) Value() (driver.Value, error) {
	return json.Marshal(s)
}

// Scan implements sql.Scanner interface for database retrieval
func (s *StoreSettings) Scan(value interface{}) error {
	if value == nil {
		*s = GetDefaultStoreSettings()
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal StoreSettings value:", value))
	}

	if len(bytes) == 0 {
		*s = GetDefaultStoreSettings()
		return nil
	}

	return json.Unmarshal(bytes, s)
}

// GetDefaultStoreSettings returns default settings for a new store
func GetDefaultStoreSettings() StoreSettings {
	return StoreSettings{
		Currency:           "USD",
		Timezone:           "UTC",
		AllowPublicListing: true,
		RequireApproval:    false,
		MaxProducts:        1000,
	}
}

func (Store) TableName() string {
	return "stores"
}