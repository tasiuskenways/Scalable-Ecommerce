package dto

import (
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
)

type UserResponse struct {
	ID          string           `json:"id"`
	Email       string           `json:"email"`
	Name        string           `json:"name"`
	IsActive    bool             `json:"is_active"`
	Profile     *ProfileResponse `json:"profile,omitempty"`
	Roles       []RoleResponse   `json:"roles,omitempty"`
	Permissions []string         `json:"permissions,omitempty"`
}

type UserListResponse struct {
	ID       string         `json:"id"`
	Email    string         `json:"email"`
	Name     string         `json:"name"`
	IsActive bool           `json:"is_active"`
	Roles    []RoleResponse `json:"roles"`
}

type UpdateUserRequest struct {
	Name     *string `json:"name"`
	IsActive *bool   `json:"is_active"`
}

func NewUserResponse(user *entities.User) *UserResponse {
	response := &UserResponse{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: user.IsActive,
	}

	// Add profile if exists
	if user.Profile != nil {
		response.Profile = NewProfileResponse(user.Profile)
	}

	// Add roles
	if len(user.Roles) > 0 {
		response.Roles = make([]RoleResponse, len(user.Roles))
		for i, role := range user.Roles {
			response.Roles[i] = *NewRoleResponse(&role)
		}

		// Add permissions
		response.Permissions = user.GetPermissions()
	}

	return response
}

func NewUserListResponse(user *entities.User) *UserListResponse {
	response := &UserListResponse{
		ID:       user.ID,
		Email:    user.Email,
		Name:     user.Name,
		IsActive: user.IsActive,
	}

	// Add roles
	if len(user.Roles) > 0 {
		response.Roles = make([]RoleResponse, len(user.Roles))
		for i, role := range user.Roles {
			response.Roles[i] = *NewRoleResponse(&role)
		}
	}

	return response
}
