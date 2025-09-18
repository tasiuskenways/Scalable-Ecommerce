package repositories

import (
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type userStoreRoleRepository struct {
	db *gorm.DB
}

func NewUserStoreRoleRepository(db *gorm.DB) repositories.UserStoreRoleRepository {
	return &userStoreRoleRepository{db: db}
}

func (r *userStoreRoleRepository) Create(role *entities.UserStoreRole) error {
	return r.db.Create(role).Error
}

func (r *userStoreRoleRepository) GetByUserAndStore(userID, storeID string) (*entities.UserStoreRole, error) {
	var role entities.UserStoreRole
	err := r.db.Where("user_id = ? AND store_id = ? AND is_active = ?", userID, storeID, true).
		First(&role).Error
	if err != nil {
		return nil, err
	}
	return &role, nil
}

func (r *userStoreRoleRepository) GetByStoreID(storeID string) ([]entities.UserStoreRole, error) {
	var roles []entities.UserStoreRole
	err := r.db.Where("store_id = ? AND is_active = ?", storeID, true).
		Order("created_at ASC").
		Find(&roles).Error
	return roles, err
}

func (r *userStoreRoleRepository) GetByUserID(userID string) ([]entities.UserStoreRole, error) {
	var roles []entities.UserStoreRole
	err := r.db.Preload("Store").
		Where("user_id = ? AND is_active = ?", userID, true).
		Order("created_at DESC").
		Find(&roles).Error
	return roles, err
}

func (r *userStoreRoleRepository) Update(role *entities.UserStoreRole) error {
	return r.db.Save(role).Error
}

func (r *userStoreRoleRepository) Delete(userID, storeID string) error {
	return r.db.Where("user_id = ? AND store_id = ?", userID, storeID).
		Delete(&entities.UserStoreRole{}).Error
}

func (r *userStoreRoleRepository) GetUserRole(userID, storeID string) (entities.StoreRole, error) {
	var role entities.UserStoreRole
	err := r.db.Where("user_id = ? AND store_id = ? AND is_active = ?", userID, storeID, true).
		First(&role).Error
	if err != nil {
		return "", err
	}
	return role.Role, nil
}

func (r *userStoreRoleRepository) HasPermission(userID, storeID string, requiredRole entities.StoreRole) (bool, error) {
	userRole, err := r.GetUserRole(userID, storeID)
	if err != nil {
		return false, err
	}

	hierarchy := entities.GetRoleHierarchy()
	userLevel := hierarchy[userRole]
	requiredLevel := hierarchy[requiredRole]

	return userLevel >= requiredLevel, nil
}

func (r *userStoreRoleRepository) IsStoreOwner(userID, storeID string) (bool, error) {
	userRole, err := r.GetUserRole(userID, storeID)
	if err != nil {
		return false, err
	}

	return userRole == entities.StoreRoleOwner, nil
}
