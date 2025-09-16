package repositories

import (
	"context"
	"errors"

	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type cartItemRepository struct {
	db *gorm.DB
}

func NewCartItemRepository(db *gorm.DB) repositories.CartItemRepository {
	return &cartItemRepository{db: db}
}

func (r *cartItemRepository) Create(ctx context.Context, item *entities.CartItem) error {
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *cartItemRepository) GetByID(ctx context.Context, id string) (*entities.CartItem, error) {
	var item entities.CartItem
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *cartItemRepository) GetByCartID(ctx context.Context, cartID string) ([]*entities.CartItem, error) {
	var items []*entities.CartItem
	err := r.db.WithContext(ctx).Where("cart_id = ?", cartID).Find(&items).Error
	return items, err
}

func (r *cartItemRepository) GetByCartAndProduct(ctx context.Context, cartID, productID string) (*entities.CartItem, error) {
	var item entities.CartItem
	err := r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).First(&item).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &item, nil
}

func (r *cartItemRepository) Update(ctx context.Context, item *entities.CartItem) error {
	return r.db.WithContext(ctx).Save(item).Error
}

func (r *cartItemRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.CartItem{}, "id = ?", id).Error
}

func (r *cartItemRepository) DeleteByCartID(ctx context.Context, cartID string) error {
	return r.db.WithContext(ctx).Where("cart_id = ?", cartID).Delete(&entities.CartItem{}).Error
}

func (r *cartItemRepository) DeleteByCartAndProduct(ctx context.Context, cartID, productID string) error {
	return r.db.WithContext(ctx).Where("cart_id = ? AND product_id = ?", cartID, productID).Delete(&entities.CartItem{}).Error
}
