package repositories

import (
	"context"
	"errors"

	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/shopping-cart-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type cartRepository struct {
	db *gorm.DB
}

func NewCartRepository(db *gorm.DB) repositories.CartRepository {
	return &cartRepository{db: db}
}

func (r *cartRepository) Create(ctx context.Context, cart *entities.Cart) error {
	return r.db.WithContext(ctx).Create(cart).Error
}

func (r *cartRepository) GetByID(ctx context.Context, id string) (*entities.Cart, error) {
	var cart entities.Cart
	err := r.db.WithContext(ctx).Preload("Items").Where("id = ?", id).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) GetByUserID(ctx context.Context, userID string) (*entities.Cart, error) {
	var cart entities.Cart
	err := r.db.WithContext(ctx).Preload("Items").Where("user_id = ?", userID).First(&cart).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &cart, nil
}

func (r *cartRepository) Update(ctx context.Context, cart *entities.Cart) error {
	return r.db.WithContext(ctx).Save(cart).Error
}

func (r *cartRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("cart_id = ?", id).Delete(&entities.CartItem{}).Error; err != nil {
			return err
		}
		return tx.Delete(&entities.Cart{}, "id = ?", id).Error
	})
}

func (r *cartRepository) DeleteByUserID(ctx context.Context, userID string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sub := tx.Model(&entities.Cart{}).Select("id").Where("user_id = ?", userID)
		if err := tx.Where("cart_id IN (?)", sub).Delete(&entities.CartItem{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&entities.Cart{}).Error
	})
}
