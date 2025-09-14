package repositories

import (
	"context"
	"errors"
	"fmt"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

// NewUserRepository creates and returns a repositories.UserRepository backed by the provided *gorm.DB.
// The returned repository uses the given DB connection for all user-related persistence operations.
func NewUserRepository(db *gorm.DB) repositories.UserRepository {
	return &userRepository{
		db: db,
	}
}

func (r *userRepository) Create(ctx context.Context, user *entities.User) error {
	// Start transaction
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Create user
		if err := tx.Create(user).Error; err != nil {
			return err
		}

		// Assign default "customer" role if no roles specified
		if len(user.Roles) == 0 {
			var defaultRole entities.Role
			if err := tx.Where("name = ?", "customer").First(&defaultRole).Error; err != nil {
				return fmt.Errorf("failed to get default role: %w", err)
			}

			if err := tx.Model(user).Association("Roles").Append(&defaultRole); err != nil {
				return fmt.Errorf("failed to assign default role: %w", err)
			}
		}

		return nil
	})
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("Profile").
		Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) GetByID(ctx context.Context, id string) (*entities.User, error) {
	var user entities.User
	if err := r.db.WithContext(ctx).
		Preload("Roles.Permissions").
		Preload("Profile").
		Where("id = ?", id).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Update(ctx context.Context, user *entities.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Where("id = ?", id).Delete(&entities.User{}).Error
}

func (r *userRepository) ExistsByEmail(ctx context.Context, email string) (bool, error) {
	var count int64
	if err := r.db.WithContext(ctx).Model(&entities.User{}).Where("email = ?", email).Count(&count).Error; err != nil {
		return false, err
	}
	return count > 0, nil
}
