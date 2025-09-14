package repositories

import (
	"context"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
)

type RoleRepository interface {
	Create(ctx context.Context, role *entities.Role) error
	GetByID(ctx context.Context, id string) (*entities.Role, error)
	GetByName(ctx context.Context, name string) (*entities.Role, error)
	GetAll(ctx context.Context) ([]entities.Role, error)
	Update(ctx context.Context, role *entities.Role) error
	Delete(ctx context.Context, id string) error
	ExistsByName(ctx context.Context, name string) (bool, error)
	GetByIDs(ctx context.Context, ids []string) ([]entities.Role, error)
}
