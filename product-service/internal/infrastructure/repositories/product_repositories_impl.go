package repositories

import (
	"context"
	"strings"

	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/product-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type productRepository struct {
	db *gorm.DB
}

func NewProductRepository(db *gorm.DB) repositories.ProductRepository {
	return &productRepository{db: db}
}

func (r *productRepository) Create(ctx context.Context, product *entities.Product) error {
	return r.db.WithContext(ctx).Create(product).Error
}

func (r *productRepository) GetByID(ctx context.Context, id string) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("id = ?", id).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetBySKU(ctx context.Context, sku string) (*entities.Product, error) {
	var product entities.Product
	err := r.db.WithContext(ctx).Preload("Category").Where("sku = ?", sku).First(&product).Error
	if err != nil {
		return nil, err
	}
	return &product, nil
}

func (r *productRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	query := r.db.WithContext(ctx).Preload("Category").Where("is_active = ?", true)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) GetByCategory(ctx context.Context, categoryID string, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	query := r.db.WithContext(ctx).Preload("Category").Where("category_id = ? AND is_active = ?", categoryID, true)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&products).Error
	return products, err
}

func (r *productRepository) Update(ctx context.Context, product *entities.Product) error {
	return r.db.WithContext(ctx).Save(product).Error
}

func (r *productRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.Product{}, "id = ?", id).Error
}

func (r *productRepository) UpdateStock(ctx context.Context, id string, stock int) error {
	return r.db.WithContext(ctx).Model(&entities.Product{}).Where("id = ?", id).Update("stock", stock).Error
}

func (r *productRepository) Search(ctx context.Context, query string, limit, offset int) ([]*entities.Product, error) {
	var products []*entities.Product
	searchQuery := r.db.WithContext(ctx).Preload("Category").Where("is_active = ?", true)

	// Search in name and description
	searchTerm := "%" + strings.ToLower(query) + "%"
	searchQuery = searchQuery.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)

	if limit > 0 {
		searchQuery = searchQuery.Limit(limit)
	}
	if offset > 0 {
		searchQuery = searchQuery.Offset(offset)
	}

	err := searchQuery.Find(&products).Error
	return products, err
}

type categoryRepository struct {
	db *gorm.DB
}

func NewCategoryRepository(db *gorm.DB) *categoryRepository {
	return &categoryRepository{db: db}
}

func (r *categoryRepository) Create(ctx context.Context, category *entities.Category) error {
	return r.db.WithContext(ctx).Create(category).Error
}

func (r *categoryRepository) GetByID(ctx context.Context, id string) (*entities.Category, error) {
	var category entities.Category
	err := r.db.WithContext(ctx).Where("id = ?", id).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetByName(ctx context.Context, name string) (*entities.Category, error) {
	var category entities.Category
	err := r.db.WithContext(ctx).Where("name = ?", name).First(&category).Error
	if err != nil {
		return nil, err
	}
	return &category, nil
}

func (r *categoryRepository) GetAll(ctx context.Context, limit, offset int) ([]*entities.Category, error) {
	var categories []*entities.Category
	query := r.db.WithContext(ctx).Where("is_active = ?", true)

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&categories).Error
	return categories, err
}

func (r *categoryRepository) Update(ctx context.Context, category *entities.Category) error {
	return r.db.WithContext(ctx).Save(category).Error
}

func (r *categoryRepository) Delete(ctx context.Context, id string) error {
	return r.db.WithContext(ctx).Delete(&entities.Category{}, "id = ?", id).Error
}

func (r *productRepository) GetByIDs(ctx context.Context, ids []string) ([]*entities.Product, error) {
	var products []*entities.Product
	err := r.db.WithContext(ctx).Where("id IN (?)", ids).Find(&products).Error
	return products, err
}
