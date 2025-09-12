package services

import (
	"context"
	"fmt"

	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/services"
)

type productService struct {
	productRepo  repositories.ProductRepository
	categoryRepo repositories.CategoryRepository
}

func NewProductService(productRepo repositories.ProductRepository, categoryRepo repositories.CategoryRepository) services.ProductService {
	return &productService{
		productRepo:  productRepo,
		categoryRepo: categoryRepo,
	}
}

func (s *productService) CreateProduct(ctx context.Context, product *entities.Product) error {
	// Check if category exists
	_, err := s.categoryRepo.GetByID(ctx, product.CategoryID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// Check if SKU already exists
	existingProduct, err := s.productRepo.GetBySKU(ctx, product.SKU)
	if err == nil && existingProduct != nil {
		return fmt.Errorf("product with SKU %s already exists", product.SKU)
	}

	return s.productRepo.Create(ctx, product)
}

func (s *productService) GetProduct(ctx context.Context, id string) (*entities.Product, error) {
	return s.productRepo.GetByID(ctx, id)
}

func (s *productService) GetProductBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	return s.productRepo.GetBySKU(ctx, sku)
}

func (s *productService) GetProducts(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	return s.productRepo.GetAll(ctx, limit, offset)
}

func (s *productService) GetProductsByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*entities.Product, error) {
	// Check if category exists
	_, err := s.categoryRepo.GetByID(ctx, categoryID)
	if err != nil {
		return nil, fmt.Errorf("category not found: %w", err)
	}

	return s.productRepo.GetByCategory(ctx, categoryID, limit, offset)
}

func (s *productService) UpdateProduct(ctx context.Context, product *entities.Product) error {
	// Check if product exists
	existingProduct, err := s.productRepo.GetByID(ctx, product.ID)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	// If category is being updated, check if new category exists
	if product.CategoryID != existingProduct.CategoryID {
		_, err := s.categoryRepo.GetByID(ctx, product.CategoryID)
		if err != nil {
			return fmt.Errorf("category not found: %w", err)
		}
	}

	// If SKU is being updated, check if new SKU already exists
	if product.SKU != existingProduct.SKU {
		existingBySKU, err := s.productRepo.GetBySKU(ctx, product.SKU)
		if err == nil && existingBySKU != nil && existingBySKU.ID != product.ID {
			return fmt.Errorf("product with SKU %s already exists", product.SKU)
		}
	}

	return s.productRepo.Update(ctx, product)
}

func (s *productService) DeleteProduct(ctx context.Context, id string) error {
	// Check if product exists
	_, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	return s.productRepo.Delete(ctx, id)
}

func (s *productService) UpdateProductStock(ctx context.Context, id string, stock int) error {
	// Check if product exists
	_, err := s.productRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("product not found: %w", err)
	}

	return s.productRepo.UpdateStock(ctx, id, stock)
}

func (s *productService) SearchProducts(ctx context.Context, query string, limit, offset int) ([]*entities.Product, error) {
	return s.productRepo.Search(ctx, query, limit, offset)
}

type categoryService struct {
	categoryRepo repositories.CategoryRepository
}

func NewCategoryService(categoryRepo repositories.CategoryRepository) services.CategoryService {
	return &categoryService{
		categoryRepo: categoryRepo,
	}
}

func (s *categoryService) CreateCategory(ctx context.Context, category *entities.Category) error {
	// Check if category name already exists
	existingCategory, err := s.categoryRepo.GetByName(ctx, category.Name)
	if err == nil && existingCategory != nil {
		return fmt.Errorf("category with name %s already exists", category.Name)
	}

	return s.categoryRepo.Create(ctx, category)
}

func (s *categoryService) GetCategory(ctx context.Context, id string) (*entities.Category, error) {
	return s.categoryRepo.GetByID(ctx, id)
}

func (s *categoryService) GetCategoryByName(ctx context.Context, name string) (*entities.Category, error) {
	return s.categoryRepo.GetByName(ctx, name)
}

func (s *categoryService) GetCategories(ctx context.Context, limit, offset int) ([]*entities.Category, error) {
	return s.categoryRepo.GetAll(ctx, limit, offset)
}

func (s *categoryService) UpdateCategory(ctx context.Context, category *entities.Category) error {
	// Check if category exists
	existingCategory, err := s.categoryRepo.GetByID(ctx, category.ID)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	// If name is being updated, check if new name already exists
	if category.Name != existingCategory.Name {
		existingByName, err := s.categoryRepo.GetByName(ctx, category.Name)
		if err == nil && existingByName != nil && existingByName.ID != category.ID {
			return fmt.Errorf("category with name %s already exists", category.Name)
		}
	}

	return s.categoryRepo.Update(ctx, category)
}

func (s *categoryService) DeleteCategory(ctx context.Context, id string) error {
	// Check if category exists
	_, err := s.categoryRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("category not found: %w", err)
	}

	return s.categoryRepo.Delete(ctx, id)
}
