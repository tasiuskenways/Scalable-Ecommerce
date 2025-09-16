package entities

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type CartItem struct {
	ID          string          `json:"id" gorm:"type:uuid;primaryKey;default:uuid_generate_v4()"`
	CartID      string          `json:"cart_id" gorm:"type:uuid;not null;index"`
	Cart        Cart            `json:"-" gorm:"foreignKey:CartID"`
	ProductID   string          `json:"product_id" gorm:"type:uuid;not null;index"`
	Quantity    int             `json:"quantity" gorm:"not null;check:quantity > 0"`
	PriceAtTime decimal.Decimal `json:"price_at_time" gorm:"type:decimal(10,2);not null"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	DeletedAt   gorm.DeletedAt  `json:"-" gorm:"index"`
}

func (CartItem) TableName() string {
	return "cart_items"
}

func (ci *CartItem) BeforeCreate(tx *gorm.DB) error {
	if ci.ID == "" {
		ci.ID = uuid.NewString()
	}
	return nil
}

// GetSubtotal calculates the subtotal for this cart item
func (ci *CartItem) GetSubtotal() decimal.Decimal {
	return ci.PriceAtTime.Mul(decimal.NewFromInt(int64(ci.Quantity)))
}
