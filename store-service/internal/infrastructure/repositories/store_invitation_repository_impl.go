package repositories

import (
	"time"

	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/repositories"
	"gorm.io/gorm"
)

type storeInvitationRepository struct {
	db *gorm.DB
}

func NewStoreInvitationRepository(db *gorm.DB) repositories.StoreInvitationRepository {
	return &storeInvitationRepository{db: db}
}

func (r *storeInvitationRepository) Create(invitation *entities.StoreInvitation) error {
	return r.db.Create(invitation).Error
}

func (r *storeInvitationRepository) GetByID(id string) (*entities.StoreInvitation, error) {
	var invitation entities.StoreInvitation
	err := r.db.Preload("Store").First(&invitation, "id = ?", id).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *storeInvitationRepository) GetByToken(token string) (*entities.StoreInvitation, error) {
	var invitation entities.StoreInvitation
	err := r.db.Preload("Store").First(&invitation, "token = ?", token).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}

func (r *storeInvitationRepository) GetByStoreID(storeID string) ([]entities.StoreInvitation, error) {
	var invitations []entities.StoreInvitation
	err := r.db.Where("store_id = ?", storeID).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *storeInvitationRepository) GetByEmail(email string) ([]entities.StoreInvitation, error) {
	var invitations []entities.StoreInvitation
	err := r.db.Preload("Store").
		Where("email = ?", email).
		Order("created_at DESC").
		Find(&invitations).Error
	return invitations, err
}

func (r *storeInvitationRepository) Update(invitation *entities.StoreInvitation) error {
	return r.db.Save(invitation).Error
}

func (r *storeInvitationRepository) Delete(id string) error {
	return r.db.Delete(&entities.StoreInvitation{}, "id = ?", id).Error
}

func (r *storeInvitationRepository) ExpireOldInvitations() error {
	return r.db.Model(&entities.StoreInvitation{}).
		Where("expires_at < ? AND status = ?", time.Now(), entities.InvitationStatusPending).
		Update("status", entities.InvitationStatusExpired).Error
}

func (r *storeInvitationRepository) GetPendingByEmailAndStore(email, storeID string) (*entities.StoreInvitation, error) {
	var invitation entities.StoreInvitation
	err := r.db.Where("email = ? AND store_id = ? AND status = ?",
		email, storeID, entities.InvitationStatusPending).
		First(&invitation).Error
	if err != nil {
		return nil, err
	}
	return &invitation, nil
}