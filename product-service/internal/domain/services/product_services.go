package services

import (
	"context"

	"github.com/tasiuskenways/Scalable-Ecommerce/product-service/internal/domain/entities"
)

type ProductService interface {
	CreateProduct(ctx context.Context, product *entities.Product) error
	GetProduct(ctx context.Context, id string) (*entities.Product, error)
	GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error)
	GetProducts(ctx context.Context, limit, offset int) ([]*entities.Product, error)
	GetProductsByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*entities.Product, error)
	UpdateProduct(ctx context.Context, product *entities.Product) error
	DeleteProduct(ctx context.Context, id string) error
	UpdateProductStock(ctx context.Context, id string, stock int) error
	SearchProducts(ctx context.Context, query string, limit, offset int) ([]*entities.Product, error)
}

type CategoryService interface {
	CreateCategory(ctx context.Context, category *entities.Category) error
	GetCategory(ctx context.Context, id string) (*entities.Category, error)
	GetCategoryByName(ctx context.Context, name string) (*entities.Category, error)
	GetCategories(ctx context.Context, limit, offset int) ([]*entities.Category, error)
	UpdateCategory(ctx context.Context, category *entities.Category) error
	DeleteCategory(ctx context.Context, id string) error
}
