package dto

import (
	"github.com/tasiuskenways/scalable-ecommerce/store-service/internal/domain/entities"
)

type CreateStoreRequest struct {
	Name        string                 `json:"name" validate:"required,min=2,max=100"`
	Slug        string                 `json:"slug" validate:"required,min=2,max=100,alphanum"`
	Description string                 `json:"description" validate:"max=1000"`
	Logo        string                 `json:"logo,omitempty" validate:"omitempty,url"`
	Banner      string                 `json:"banner,omitempty" validate:"omitempty,url"`
	Website     string                 `json:"website,omitempty" validate:"omitempty,url"`
	Phone       string                 `json:"phone,omitempty"`
	Email       string                 `json:"email,omitempty" validate:"omitempty,email"`
	Address     string                 `json:"address,omitempty"`
	City        string                 `json:"city,omitempty"`
	State       string                 `json:"state,omitempty"`
	Country     string                 `json:"country,omitempty"`
	PostalCode  string                 `json:"postal_code,omitempty"`
	Settings    entities.StoreSettings `json:"settings,omitempty"`
}

type UpdateStoreRequest struct {
	Name        *string                 `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
	Description *string                 `json:"description,omitempty" validate:"omitempty,max=1000"`
	Logo        *string                 `json:"logo,omitempty" validate:"omitempty,url"`
	Banner      *string                 `json:"banner,omitempty" validate:"omitempty,url"`
	Website     *string                 `json:"website,omitempty" validate:"omitempty,url"`
	Phone       *string                 `json:"phone,omitempty"`
	Email       *string                 `json:"email,omitempty" validate:"omitempty,email"`
	Address     *string                 `json:"address,omitempty"`
	City        *string                 `json:"city,omitempty"`
	State       *string                 `json:"state,omitempty"`
	Country     *string                 `json:"country,omitempty"`
	PostalCode  *string                 `json:"postal_code,omitempty"`
	IsActive    *bool                   `json:"is_active,omitempty"`
	Settings    *entities.StoreSettings `json:"settings,omitempty"`
}

type StoreResponse struct {
	ID          string                    `json:"id"`
	Name        string                    `json:"name"`
	Slug        string                    `json:"slug"`
	Description string                    `json:"description"`
	Logo        string                    `json:"logo,omitempty"`
	Banner      string                    `json:"banner,omitempty"`
	Website     string                    `json:"website,omitempty"`
	Phone       string                    `json:"phone,omitempty"`
	Email       string                    `json:"email,omitempty"`
	Address     string                    `json:"address,omitempty"`
	City        string                    `json:"city,omitempty"`
	State       string                    `json:"state,omitempty"`
	Country     string                    `json:"country,omitempty"`
	PostalCode  string                    `json:"postal_code,omitempty"`
	IsActive    bool                      `json:"is_active"`
	Settings    entities.StoreSettings    `json:"settings"`
	CreatedAt   string                    `json:"created_at"`
	UpdatedAt   string                    `json:"updated_at"`
	UserRole    *entities.StoreRole       `json:"user_role,omitempty"`
	Permissions *entities.RolePermissions `json:"permissions,omitempty"`
}

type StoreListResponse struct {
	Stores     []StoreResponse `json:"stores"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PerPage    int             `json:"per_page"`
	TotalPages int             `json:"total_pages"`
}

type InviteMemberRequest struct {
	Email string             `json:"email" validate:"required,email"`
	Role  entities.StoreRole `json:"role" validate:"required,oneof=ADMIN MANAGER MEMBER"`
}

type UpdateMemberRoleRequest struct {
	Role entities.StoreRole `json:"role" validate:"required,oneof=ADMIN MANAGER MEMBER"`
}

type StoreInvitationResponse struct {
	ID          string                    `json:"id"`
	StoreID     string                    `json:"store_id"`
	Email       string                    `json:"email"`
	Role        entities.StoreRole        `json:"role"`
	Status      entities.InvitationStatus `json:"status"`
	ExpiresAt   string                    `json:"expires_at"`
	CreatedAt   string                    `json:"created_at"`
	Store       *StoreResponse            `json:"store,omitempty"`
	InviteToken *string                   `json:"invite_token,omitempty"`
}

type StoreMemberResponse struct {
	ID       string             `json:"id"`
	UserID   string             `json:"user_id"`
	Role     entities.StoreRole `json:"role"`
	IsActive bool               `json:"is_active"`
	JoinedAt string             `json:"joined_at"`
}

type AcceptInvitationRequest struct {
	Token string `json:"token" validate:"required"`
}
