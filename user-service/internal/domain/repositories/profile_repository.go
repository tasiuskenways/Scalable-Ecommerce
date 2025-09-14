package repositories

import (
	"context"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
)

type ProfileRepository interface {
	Create(ctx context.Context, profile *entities.UserProfile) error
	GetByUserID(ctx context.Context, userID string) (*entities.UserProfile, error)
	GetByID(ctx context.Context, id string) (*entities.UserProfile, error)
	Update(ctx context.Context, profile *entities.UserProfile) error
	Delete(ctx context.Context, id string) error
	ExistsByUserID(ctx context.Context, userID string) (bool, error)
}
