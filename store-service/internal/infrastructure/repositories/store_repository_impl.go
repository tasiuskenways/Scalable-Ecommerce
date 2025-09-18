package repositories

import (
	"errors"
	"strings"

	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/repositories"
	"gorm.io/gorm"
)

var ErrStoreNotFound = errors.New("store not found")
var ErrStoreSlugExists = errors.New("store slug already exists")

type storeRepository struct {
	db *gorm.DB
}

func NewStoreRepository(db *gorm.DB) repositories.StoreRepository {
	return &storeRepository{db: db}
}

func (r *storeRepository) Create(store *entities.Store) error {
	exists, err := r.SlugExists(store.Slug)
	if err != nil {
		return err
	}
	if exists {
		return errors.New("store slug already exists")
	}
	return r.db.Create(store).Error
}

func (r *storeRepository) GetByID(id string) (*entities.Store, error) {
	var store entities.Store
	err := r.db.First(&store, "id = ?", id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) GetBySlug(slug string) (*entities.Store, error) {
	var store entities.Store
	err := r.db.First(&store, "slug = ?", slug).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, ErrStoreNotFound
		}
		return nil, err
	}
	return &store, nil
}

func (r *storeRepository) GetByUserID(userID string, limit, offset int) ([]entities.Store, error) {
	var stores []entities.Store

	query := r.db.
		Joins("INNER JOIN user_store_roles ON stores.id = user_store_roles.store_id").
		Where("user_store_roles.user_id = ? AND user_store_roles.is_active = ?", userID, true).
		Order("stores.created_at DESC")

	if limit > 0 {
		query = query.Limit(limit)
	}
	if offset > 0 {
		query = query.Offset(offset)
	}

	err := query.Find(&stores).Error
	return stores, err
}

func (r *storeRepository) Update(store *entities.Store) error {
	return r.db.Save(store).Error
}

func (r *storeRepository) Delete(id string) error {
	return r.db.Delete(&entities.Store{}, "id = ?", id).Error
}

func (r *storeRepository) GetStoresByFilter(filter repositories.StoreFilter) ([]entities.Store, int64, error) {
	var stores []entities.Store
	var total int64

	query := r.db.Model(&entities.Store{})

	if filter.UserID != "" {
		query = query.
			Joins("INNER JOIN user_store_roles ON stores.id = user_store_roles.store_id").
			Where("user_store_roles.user_id = ? AND user_store_roles.is_active = ?", filter.UserID, true)
	}

	if filter.IsActive != nil {
		query = query.Where("stores.is_active = ?", *filter.IsActive)
	}

	if filter.Search != "" {
		searchTerm := "%" + strings.ToLower(filter.Search) + "%"
		query = query.Where(
			"LOWER(stores.name) LIKE ? OR LOWER(stores.description) LIKE ?",
			searchTerm, searchTerm,
		)
	}

	// Count total
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// Apply pagination
	if filter.Limit > 0 {
		query = query.Limit(filter.Limit)
	}
	if filter.Offset > 0 {
		query = query.Offset(filter.Offset)
	}

	query = query.Order("stores.created_at DESC")

	err := query.Find(&stores).Error
	return stores, total, err
}

func (r *storeRepository) SlugExists(slug string, excludeID ...string) (bool, error) {
	var count int64
	query := r.db.Model(&entities.Store{}).Where("slug = ?", slug)

	if len(excludeID) > 0 && excludeID[0] != "" {
		query = query.Where("id != ?", excludeID[0])
	}

	err := query.Count(&count).Error
	return count > 0, err
}
