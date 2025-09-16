package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type Cart struct {
	ID        string         `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	UserID    string         `json:"user_id" gorm:"type:uuid;not null;index"`
	Items     []CartItem     `json:"items,omitempty" gorm:"foreignKey:CartID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}

func (Cart) TableName() string {
	return "carts"
}

func (c *Cart) BeforeCreate(tx *gorm.DB) error {
	if c.ID == "" {
		c.ID = uuid.NewString()
	}
	return nil
}

// CalculateTotal calculates the total price of all items in the cart
func (c *Cart) CalculateTotal() decimal.Decimal {
	total := decimal.NewFromFloat(0)
	for _, item := range c.Items {
		itemTotal := item.PriceAtTime.Mul(decimal.NewFromInt(int64(item.Quantity)))
		total = total.Add(itemTotal)
	}
	return total
}

// GetTotalItems returns the total number of items in the cart
func (c *Cart) GetTotalItems() int {
	total := 0
	for _, item := range c.Items {
		total += item.Quantity
	}
	return total
}
