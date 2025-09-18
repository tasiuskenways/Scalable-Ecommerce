package repositories

import (
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
)

type StoreRepository interface {
	Create(store *entities.Store) error
	GetByID(id string) (*entities.Store, error)
	GetBySlug(slug string) (*entities.Store, error)
	GetByUserID(userID string, limit, offset int) ([]entities.Store, error)
	Update(store *entities.Store) error
	Delete(id string) error
	GetStoresByFilter(filter StoreFilter) ([]entities.Store, int64, error)
	SlugExists(slug string, excludeID ...string) (bool, error)
}

type StoreFilter struct {
	UserID   string
	IsActive *bool
	Search   string
	Limit    int
	Offset   int
}

type UserStoreRoleRepository interface {
	Create(role *entities.UserStoreRole) error
	GetByUserAndStore(userID, storeID string) (*entities.UserStoreRole, error)
	GetByStoreID(storeID string) ([]entities.UserStoreRole, error)
	GetByUserID(userID string) ([]entities.UserStoreRole, error)
	Update(role *entities.UserStoreRole) error
	Delete(userID, storeID string) error
	GetUserRole(userID, storeID string) (entities.StoreRole, error)
	HasPermission(userID, storeID string, requiredRole entities.StoreRole) (bool, error)
	IsStoreOwner(userID, storeID string) (bool, error)
}

type StoreInvitationRepository interface {
	Create(invitation *entities.StoreInvitation) error
	GetByID(id string) (*entities.StoreInvitation, error)
	GetByToken(token string) (*entities.StoreInvitation, error)
	GetByStoreID(storeID string) ([]entities.StoreInvitation, error)
	GetByEmail(email string) ([]entities.StoreInvitation, error)
	Update(invitation *entities.StoreInvitation) error
	Delete(id string) error
	ExpireOldInvitations() error
	GetPendingByEmailAndStore(email, storeID string) (*entities.StoreInvitation, error)
}