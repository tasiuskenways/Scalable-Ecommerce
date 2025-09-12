package repositories

import (
	"context"

	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/entities"
)

type ProductRepository interface {
	Create(ctx context.Context, product *entities.Product) error
	GetByID(ctx context.Context, id string) (*entities.Product, error)
	GetBySKU(ctx context.Context, sku string) (*entities.Product, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Product, error)
	GetByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*entities.Product, error)
	Update(ctx context.Context, product *entities.Product) error
	Delete(ctx context.Context, id string) error
	UpdateStock(ctx context.Context, id string, stock int) error
	Search(ctx context.Context, query string, limit, offset int) ([]*entities.Product, error)
}

type CategoryRepository interface {
	Create(ctx context.Context, category *entities.Category) error
	GetByID(ctx context.Context, id string) (*entities.Category, error)
	GetByName(ctx context.Context, name string) (*entities.Category, error)
	GetAll(ctx context.Context, limit, offset int) ([]*entities.Category, error)
	Update(ctx context.Context, category *entities.Category) error
	Delete(ctx context.Context, id string) error
}
