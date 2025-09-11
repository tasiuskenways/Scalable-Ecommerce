package entities

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type Product struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string         `json:"name" gorm:"not null"`
	Description string         `json:"description"`
	Price       float64        `json:"price" gorm:"not null"`
	Stock       int            `json:"stock" gorm:"default:0"`
	CategoryID  string         `json:"category_id" gorm:"type:uuid;not null"`
	Category    Category       `json:"category" gorm:"foreignKey:CategoryID"`
	SKU         string         `json:"sku" gorm:"unique;not null"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Product) TableName() string {
	return "products"
}

// BeforeCreate hook to set default values
func (p *Product) BeforeCreate(tx *gorm.DB) error {
	if p.ID == "" {
		p.ID = uuid.NewString()
	}
	if !p.IsActive {
		p.IsActive = true
	}
	return nil
}

type Category struct {
	ID          string         `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	Name        string         `json:"name" gorm:"unique;not null"`
	Description string         `json:"description"`
	IsActive    bool           `json:"is_active" gorm:"default:true"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
	Products    []Product      `json:"products,omitempty" gorm:"foreignKey:CategoryID"`
}

func (Category) TableName() string {
	return "categories"
}

// BeforeCreate hook to set default values
func (c *Category) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	if !c.IsActive {
		c.IsActive = true
	}
	return nil
}
