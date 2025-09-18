package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/services"
	repoImpl "github.com/tasiuskenways/scalable-ecommerce/store-service/internal/infrastructure/repositories"
)

type storeService struct {
	storeRepo      repositories.StoreRepository
	roleRepo       repositories.UserStoreRoleRepository
	invitationRepo repositories.StoreInvitationRepository
}

func NewStoreService(
	storeRepo repositories.StoreRepository,
	roleRepo repositories.UserStoreRoleRepository,
	invitationRepo repositories.StoreInvitationRepository,
) services.StoreService {
	return &storeService{
		storeRepo:      storeRepo,
		roleRepo:       roleRepo,
		invitationRepo: invitationRepo,
	}
}

func (s *storeService) CreateStore(userID string, req dto.CreateStoreRequest) (*dto.StoreResponse, error) {
	// Validate and clean slug
	slug := s.generateSlug(req.Slug)

	// Check if slug exists
	exists, err := s.storeRepo.SlugExists(slug)
	if err != nil {
		return nil, fmt.Errorf("failed to check slug existence: %w", err)
	}
	if exists {
		return nil, errors.New("store slug already exists")
	}

	// Create store with default settings if none provided
	settings := req.Settings
	if settings == (entities.StoreSettings{}) {
		settings = entities.GetDefaultStoreSettings()
	}

	store := &entities.Store{
		Name:        req.Name,
		Slug:        slug,
		Description: req.Description,
		Logo:        req.Logo,
		Banner:      req.Banner,
		Website:     req.Website,
		Phone:       req.Phone,
		Email:       req.Email,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		PostalCode:  req.PostalCode,
		IsActive:    true,
		Settings:    settings,
	}

	if err := s.storeRepo.Create(store); err != nil {
		return nil, fmt.Errorf("failed to create store: %w", err)
	}

	// Make user the owner
	ownerRole := &entities.UserStoreRole{
		UserID:   userID,
		StoreID:  store.ID,
		Role:     entities.StoreRoleOwner,
		IsActive: true,
	}

	if err := s.roleRepo.Create(ownerRole); err != nil {
		return nil, fmt.Errorf("failed to create owner role: %w", err)
	}

	ownerRoleValue := entities.StoreRoleOwner
	return s.mapStoreToResponse(store, &ownerRoleValue), nil
}

func (s *storeService) GetStore(storeID, userID string) (*dto.StoreResponse, error) {
	store, err := s.storeRepo.GetByID(storeID)
	if err != nil {
		if errors.Is(err, repoImpl.ErrStoreNotFound) {
			return nil, services.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	// Get user role in store
	role, err := s.roleRepo.GetUserRole(userID, storeID)
	if err != nil {
		// User doesn't have access to this store
		return nil, errors.New("store not found or access denied")
	}

	return s.mapStoreToResponse(store, &role), nil
}

func (s *storeService) GetStoreBySlug(slug, userID string) (*dto.StoreResponse, error) {
	store, err := s.storeRepo.GetBySlug(slug)
	if err != nil {
		if errors.Is(err, repoImpl.ErrStoreNotFound) {
			return nil, services.ErrNotFound
		}
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	// Get user role in store
	role, err := s.roleRepo.GetUserRole(userID, store.ID)
	if err != nil {
		// User doesn't have access to this store
		return nil, errors.New("store not found or access denied")
	}

	return s.mapStoreToResponse(store, &role), nil
}

func (s *storeService) GetUserStores(userID string, page, perPage int) (*dto.StoreListResponse, error) {
	offset := (page - 1) * perPage

	stores, total, err := s.storeRepo.GetStoresByFilter(repositories.StoreFilter{
		UserID: userID,
		Limit:  perPage,
		Offset: offset,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get user stores: %w", err)
	}

	storeResponses := make([]dto.StoreResponse, len(stores))
	for i, store := range stores {
		// Get user role for each store
		role, _ := s.roleRepo.GetUserRole(userID, store.ID)
		storeResponses[i] = *s.mapStoreToResponse(&store, &role)
	}

	totalPages := int((total + int64(perPage) - 1) / int64(perPage))

	return &dto.StoreListResponse{
		Stores:     storeResponses,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: totalPages,
	}, nil
}

func (s *storeService) InviteMember(storeID, inviterID string, req dto.InviteMemberRequest) (*dto.StoreInvitationResponse, error) {
	// Check if inviter has permission to invite members (OWNER or ADMIN only)
	inviterRole, err := s.roleRepo.GetUserRole(inviterID, storeID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	permissions := entities.GetPermissions(inviterRole)
	if !permissions.CanInviteMembers {
		return nil, errors.New("insufficient permissions to invite members")
	}

	// Check if invitation already exists
	existing, err := s.invitationRepo.GetPendingByEmailAndStore(req.Email, storeID)
	if err == nil && existing != nil {
		return nil, errors.New("invitation already sent to this email")
	}

	// Generate invitation token
	token, err := s.generateToken()
	if err != nil {
		return nil, fmt.Errorf("failed to generate invitation token: %w", err)
	}

	// Create invitation
	invitation := &entities.StoreInvitation{
		StoreID:   storeID,
		InviterID: inviterID,
		Email:     req.Email,
		Role:      req.Role,
		Token:     token,
		ExpiresAt: time.Now().Add(7 * 24 * time.Hour), // 7 days
	}

	if err := s.invitationRepo.Create(invitation); err != nil {
		return nil, fmt.Errorf("failed to create invitation: %w", err)
	}

	return s.mapInvitationToResponse(invitation), nil
}

func (s *storeService) AcceptInvitation(userID string, req dto.AcceptInvitationRequest) error {
	invitation, err := s.invitationRepo.GetByToken(req.Token)
	if err != nil {
		return errors.New("invalid invitation token")
	}

	if !invitation.CanAccept() {
		return errors.New("invitation has expired or is no longer valid")
	}

	// Check if user is already a member
	existing, err := s.roleRepo.GetByUserAndStore(userID, invitation.StoreID)
	if err == nil && existing != nil {
		return errors.New("user is already a member of this store")
	}

	// Create user store role
	role := &entities.UserStoreRole{
		UserID:   userID,
		StoreID:  invitation.StoreID,
		Role:     invitation.Role,
		IsActive: true,
	}

	if err := s.roleRepo.Create(role); err != nil {
		return fmt.Errorf("failed to add user to store: %w", err)
	}

	// Update invitation status
	now := time.Now()
	invitation.Status = entities.InvitationStatusAccepted
	invitation.InviteeID = &userID
	invitation.AcceptedAt = &now

	if err := s.invitationRepo.Update(invitation); err != nil {
		return fmt.Errorf("failed to update invitation: %w", err)
	}

	return nil
}

// Helper methods
func (s *storeService) generateSlug(input string) string {
	// Convert to lowercase and replace spaces with hyphens
	slug := strings.ToLower(strings.TrimSpace(input))
	slug = regexp.MustCompile(`[^a-z0-9\-]`).ReplaceAllString(slug, "-")
	slug = regexp.MustCompile(`-+`).ReplaceAllString(slug, "-")
	slug = strings.Trim(slug, "-")
	return slug
}

func (s *storeService) generateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *storeService) mapStoreToResponse(store *entities.Store, userRole *entities.StoreRole) *dto.StoreResponse {
	response := &dto.StoreResponse{
		ID:          store.ID,
		Name:        store.Name,
		Slug:        store.Slug,
		Description: store.Description,
		Logo:        store.Logo,
		Banner:      store.Banner,
		Website:     store.Website,
		Phone:       store.Phone,
		Email:       store.Email,
		Address:     store.Address,
		City:        store.City,
		State:       store.State,
		Country:     store.Country,
		PostalCode:  store.PostalCode,
		IsActive:    store.IsActive,
		Settings:    store.Settings,
		CreatedAt:   store.CreatedAt.Format(time.RFC3339),
		UpdatedAt:   store.UpdatedAt.Format(time.RFC3339),
	}

	if userRole != nil {
		response.UserRole = userRole
		permissions := entities.GetPermissions(*userRole)
		response.Permissions = &permissions
	}

	return response
}

func (s *storeService) UpdateStore(storeID, userID string, req dto.UpdateStoreRequest) (*dto.StoreResponse, error) {
	// Get user's role
	userRole, err := s.roleRepo.GetUserRole(userID, storeID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	// Check if user has permission to edit store settings
	permissions := entities.GetPermissions(userRole)
	if !permissions.CanEditStoreSettings {
		return nil, errors.New("insufficient permissions to update store")
	}

	// Get existing store
	store, err := s.storeRepo.GetByID(storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get store: %w", err)
	}

	// Update fields if provided
	if req.Name != nil {
		store.Name = *req.Name
	}
	if req.Description != nil {
		store.Description = *req.Description
	}
	if req.Logo != nil {
		store.Logo = *req.Logo
	}
	if req.Banner != nil {
		store.Banner = *req.Banner
	}
	if req.Website != nil {
		store.Website = *req.Website
	}
	if req.Phone != nil {
		store.Phone = *req.Phone
	}
	if req.Email != nil {
		store.Email = *req.Email
	}
	if req.Address != nil {
		store.Address = *req.Address
	}
	if req.City != nil {
		store.City = *req.City
	}
	if req.State != nil {
		store.State = *req.State
	}
	if req.Country != nil {
		store.Country = *req.Country
	}
	if req.PostalCode != nil {
		store.PostalCode = *req.PostalCode
	}
	if req.IsActive != nil {
		store.IsActive = *req.IsActive
	}
	if req.Settings != nil {
		store.Settings = *req.Settings
	}

	if err := s.storeRepo.Update(store); err != nil {
		return nil, fmt.Errorf("failed to update store: %w", err)
	}

	// Get user role
	role, _ := s.roleRepo.GetUserRole(userID, storeID)
	return s.mapStoreToResponse(store, &role), nil
}

func (s *storeService) DeleteStore(storeID, userID string) error {
	// Only store owner can delete store
	isOwner, err := s.roleRepo.IsStoreOwner(userID, storeID)
	if err != nil || !isOwner {
		return errors.New("only store owner can delete the store")
	}

	return s.storeRepo.Delete(storeID)
}

func (s *storeService) GetStoreMembers(storeID, userID string) ([]dto.StoreMemberResponse, error) {
	// Check if user has access to store
	_, err := s.roleRepo.GetUserRole(userID, storeID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	members, err := s.roleRepo.GetByStoreID(storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get store members: %w", err)
	}

	responses := make([]dto.StoreMemberResponse, len(members))
	for i, member := range members {
		responses[i] = dto.StoreMemberResponse{
			ID:       member.ID,
			UserID:   member.UserID,
			Role:     member.Role,
			IsActive: member.IsActive,
			JoinedAt: member.JoinedAt.Format(time.RFC3339),
		}
	}

	return responses, nil
}

func (s *storeService) UpdateMemberRole(storeID, memberUserID, requesterID string, req dto.UpdateMemberRoleRequest) error {
	// Get requester's role
	requesterRole, err := s.roleRepo.GetUserRole(requesterID, storeID)
	if err != nil {
		return errors.New("access denied")
	}

	// Check if requester has permission to manage members
	permissions := entities.GetPermissions(requesterRole)
	if !permissions.CanManageMembers {
		return errors.New("insufficient permissions to update member role")
	}

	// Get current member role
	memberRole, err := s.roleRepo.GetByUserAndStore(memberUserID, storeID)
	if err != nil {
		return errors.New("member not found")
	}

	// Cannot change owner's role
	if memberRole.Role == entities.StoreRoleOwner {
		return errors.New("cannot change store owner's role")
	}

	// Cannot promote someone to owner (ownership transfer is a different operation)
	if req.Role == entities.StoreRoleOwner {
		return errors.New("cannot promote member to owner")
	}

	// Requester can only manage users with lower roles
	if !requesterRole.HasPermission(memberRole.Role) {
		return errors.New("cannot modify role of user with equal or higher permissions")
	}

	// Requester cannot assign a role higher than their own
	if !requesterRole.HasPermission(req.Role) && requesterRole != req.Role {
		return errors.New("cannot assign role higher than your own")
	}

	// Update role
	memberRole.Role = req.Role
	return s.roleRepo.Update(memberRole)
}

func (s *storeService) RemoveMember(storeID, memberUserID, requesterID string) error {
	// Get requester's role
	requesterRole, err := s.roleRepo.GetUserRole(requesterID, storeID)
	if err != nil {
		return errors.New("access denied")
	}

	// Check if requester has permission to manage members
	permissions := entities.GetPermissions(requesterRole)
	if !permissions.CanManageMembers {
		return errors.New("insufficient permissions to remove member")
	}

	// Get member to be removed role
	memberRole, err := s.roleRepo.GetUserRole(memberUserID, storeID)
	if err != nil {
		return errors.New("member not found")
	}

	// Cannot remove store owner
	if memberRole == entities.StoreRoleOwner {
		return errors.New("cannot remove store owner")
	}

	// Cannot remove self
	if memberUserID == requesterID {
		return errors.New("cannot remove yourself from the store")
	}

	// Requester can only remove users with lower roles
	if !requesterRole.HasPermission(memberRole) {
		return errors.New("cannot remove user with equal or higher permissions")
	}

	return s.roleRepo.Delete(memberUserID, storeID)
}

func (s *storeService) GetStoreInvitations(storeID, userID string) ([]dto.StoreInvitationResponse, error) {
	// Get user's role
	userRole, err := s.roleRepo.GetUserRole(userID, storeID)
	if err != nil {
		return nil, errors.New("access denied")
	}

	// Check if user has permission to manage members (only admin/owner can view invitations)
	permissions := entities.GetPermissions(userRole)
	if !permissions.CanManageMembers {
		return nil, errors.New("insufficient permissions to view invitations")
	}

	invitations, err := s.invitationRepo.GetByStoreID(storeID)
	if err != nil {
		return nil, fmt.Errorf("failed to get invitations: %w", err)
	}

	responses := make([]dto.StoreInvitationResponse, len(invitations))
	for i, invitation := range invitations {
		responses[i] = *s.mapInvitationToResponse(&invitation)
	}

	return responses, nil
}

func (s *storeService) GetUserInvitations(userEmail string) ([]dto.StoreInvitationResponse, error) {
	invitations, err := s.invitationRepo.GetByEmail(userEmail)
	if err != nil {
		return nil, fmt.Errorf("failed to get user invitations: %w", err)
	}

	responses := make([]dto.StoreInvitationResponse, len(invitations))
	for i, invitation := range invitations {
		responses[i] = *s.mapInvitationToResponse(&invitation)
	}

	return responses, nil
}

func (s *storeService) mapInvitationToResponse(invitation *entities.StoreInvitation) *dto.StoreInvitationResponse {
	return &dto.StoreInvitationResponse{
		ID:          invitation.ID,
		StoreID:     invitation.StoreID,
		Email:       invitation.Email,
		Role:        invitation.Role,
		Status:      invitation.Status,
		ExpiresAt:   invitation.ExpiresAt.Format(time.RFC3339),
		CreatedAt:   invitation.CreatedAt.Format(time.RFC3339),
		InviteToken: &invitation.Token,
	}
}
