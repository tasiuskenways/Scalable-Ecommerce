package repositories

import (
	"context"

	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/entities"
)

type CartRepository interface {
	Create(ctx context.Context, cart *entities.Cart) error
	GetByID(ctx context.Context, id string) (*entities.Cart, error)
	GetByUserID(ctx context.Context, userID string) (*entities.Cart, error)
	Update(ctx context.Context, cart *entities.Cart) error
	Delete(ctx context.Context, id string) error
	DeleteByUserID(ctx context.Context, userID string) error
}

type CartItemRepository interface {
	Create(ctx context.Context, item *entities.CartItem) error
	GetByID(ctx context.Context, id string) (*entities.CartItem, error)
	GetByCartID(ctx context.Context, cartID string) ([]*entities.CartItem, error)
	GetByCartAndProduct(ctx context.Context, cartID, productID string) (*entities.CartItem, error)
	Update(ctx context.Context, item *entities.CartItem) error
	Delete(ctx context.Context, id string) error
	DeleteByCartID(ctx context.Context, cartID string) error
	DeleteByCartAndProduct(ctx context.Context, cartID, productID string) error
}
