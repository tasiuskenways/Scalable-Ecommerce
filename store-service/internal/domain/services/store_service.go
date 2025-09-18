package services

import (
	"errors"

	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/application/dto"
)

type StoreService interface {
	CreateStore(userID string, req dto.CreateStoreRequest) (*dto.StoreResponse, error)
	GetStore(storeID, userID string) (*dto.StoreResponse, error)
	GetStoreBySlug(slug, userID string) (*dto.StoreResponse, error)
	GetUserStores(userID string, page, perPage int) (*dto.StoreListResponse, error)
	UpdateStore(storeID, userID string, req dto.UpdateStoreRequest) (*dto.StoreResponse, error)
	DeleteStore(storeID, userID string) error

	// Member management
	InviteMember(storeID, inviterID string, req dto.InviteMemberRequest) (*dto.StoreInvitationResponse, error)
	AcceptInvitation(userID string, req dto.AcceptInvitationRequest) error
	GetStoreMembers(storeID, userID string) ([]dto.StoreMemberResponse, error)
	UpdateMemberRole(storeID, memberUserID, requesterID string, req dto.UpdateMemberRoleRequest) error
	RemoveMember(storeID, memberUserID, requesterID string) error
	GetStoreInvitations(storeID, userID string) ([]dto.StoreInvitationResponse, error)
	GetUserInvitations(userEmail string) ([]dto.StoreInvitationResponse, error)
}

var ErrNotFound = errors.New("resource not found")
