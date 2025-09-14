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

// NewUserResponse creates a UserResponse DTO from a domain User.
// 
// It copies ID, Email, Name, and IsActive. If the domain user has a Profile,
// the response.Profile is populated via NewProfileResponse. If the domain user
// has Roles, those are converted to RoleResponse entries and response.Permissions
// is set from user.GetPermissions().
// 
// The input `user` must be non-nil.
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

// NewUserListResponse creates a UserListResponse DTO from a domain User.
// 
// The returned DTO contains the user's ID, email, name, active status, and a
// slice of role DTOs (empty if user has no roles). The input `user` must be
// non-nil.
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
