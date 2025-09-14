package repositories

import (
	"context"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
)

type PermissionRepository interface {
	Create(ctx context.Context, permission *entities.Permission) error
	GetByID(ctx context.Context, id string) (*entities.Permission, error)
	GetByName(ctx context.Context, name string) (*entities.Permission, error)
	GetAll(ctx context.Context) ([]entities.Permission, error)
	Update(ctx context.Context, permission *entities.Permission) error
	Delete(ctx context.Context, id string) error
	ExistsByName(ctx context.Context, name string) (bool, error)
	GetByIDs(ctx context.Context, ids []string) ([]entities.Permission, error)
}
