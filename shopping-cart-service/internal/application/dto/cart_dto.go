package dto

import (
	"time"

	"github.com/shopspring/decimal"
)

type AddItemRequest struct {
	ProductID string `json:"product_id" validate:"required,uuid"`
	Quantity  int    `json:"quantity" validate:"required,min=1"`
}

type UpdateItemRequest struct {
	Quantity int `json:"quantity" validate:"required,min=1"`
}

type CartItemResponse struct {
	ID          string          `json:"id"`
	ProductID   string          `json:"product_id"`
	Quantity    int             `json:"quantity"`
	PriceAtTime decimal.Decimal `json:"price_at_time"`
	Subtotal    decimal.Decimal `json:"subtotal"`
	Product     *ProductInfo    `json:"product"`
	Available   bool            `json:"available"`
	StockStatus string          `json:"stock_status"`
}

type ProductInfo struct {
	Name     string          `json:"name"`
	Price    decimal.Decimal `json:"price"`
	Category string          `json:"category"`
	SKU      string          `json:"sku"`
	IsActive bool            `json:"is_active"`
	Stock    int             `json:"stock"`
}

type CartResponse struct {
	ID         string             `json:"id"`
	UserID     string             `json:"user_id"`
	Items      []CartItemResponse `json:"items"`
	TotalItems int                `json:"total_items"`
	TotalPrice decimal.Decimal    `json:"total_price"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

type CartValidationResponse struct {
	Valid         bool                  `json:"valid"`
	TotalItems    int                   `json:"total_items"`
	TotalPrice    decimal.Decimal       `json:"total_price"`
	InvalidItems  []InvalidItemResponse `json:"invalid_items,omitempty"`
	UpdatedPrices []PriceUpdateResponse `json:"updated_prices,omitempty"`
}

type InvalidItemResponse struct {
	ProductID string `json:"product_id"`
	Reason    string `json:"reason"`
}

type PriceUpdateResponse struct {
	ProductID string          `json:"product_id"`
	OldPrice  decimal.Decimal `json:"old_price"`
	NewPrice  decimal.Decimal `json:"new_price"`
}
