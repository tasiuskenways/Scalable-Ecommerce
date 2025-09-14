package services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
	"gorm.io/gorm"
)

type roleService struct {
	roleRepo       repositories.RoleRepository
	permissionRepo repositories.PermissionRepository
	userRepo       repositories.UserRepository
	db             *gorm.DB
}

// NewRoleService creates and returns a services.RoleService backed by the provided
// role, permission and user repositories and the given GORM database connection.
func NewRoleService(
	roleRepo repositories.RoleRepository,
	permissionRepo repositories.PermissionRepository,
	userRepo repositories.UserRepository,
	db *gorm.DB,
) services.RoleService {
	return &roleService{
		roleRepo:       roleRepo,
		permissionRepo: permissionRepo,
		userRepo:       userRepo,
		db:             db,
	}
}

func (s *roleService) CreateRole(ctx *fiber.Ctx, req *dto.CreateRoleRequest) (*dto.RoleResponse, error) {
	// Check if role already exists
	exists, err := s.roleRepo.ExistsByName(ctx.Context(), req.Name)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("role with this name already exists")
	}

	// Get permissions
	permissions, err := s.permissionRepo.GetByIDs(ctx.Context(), req.Permissions)
	if err != nil {
		return nil, err
	}

	role := &entities.Role{
		Name:        req.Name,
		Description: req.Description,
		Permissions: permissions,
	}

	if err := s.roleRepo.Create(ctx.Context(), role); err != nil {
		return nil, err
	}

	return dto.NewRoleResponse(role), nil
}

func (s *roleService) GetRole(ctx *fiber.Ctx, id string) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	return dto.NewRoleResponse(role), nil
}

func (s *roleService) GetAllRoles(ctx *fiber.Ctx) ([]dto.RoleResponse, error) {
	roles, err := s.roleRepo.GetAll(ctx.Context())
	if err != nil {
		return nil, err
	}

	responses := make([]dto.RoleResponse, len(roles))
	for i, role := range roles {
		responses[i] = *dto.NewRoleResponse(&role)
	}

	return responses, nil
}

func (s *roleService) UpdateRole(ctx *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (*dto.RoleResponse, error) {
	role, err := s.roleRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return nil, err
	}
	if role == nil {
		return nil, errors.New("role not found")
	}

	// Update fields
	if req.Name != nil {
		// Check if new name already exists (excluding current role)
		existingRole, err := s.roleRepo.GetByName(ctx.Context(), *req.Name)
		if err != nil {
			return nil, err
		}
		if existingRole != nil && existingRole.ID != id {
			return nil, errors.New("role with this name already exists")
		}
		role.Name = *req.Name
	}
	if req.Description != nil {
		role.Description = *req.Description
	}
	if req.Permissions != nil {
		permissions, err := s.permissionRepo.GetByIDs(ctx.Context(), req.Permissions)
		if err != nil {
			return nil, err
		}
		role.Permissions = permissions
	}

	if err := s.roleRepo.Update(ctx.Context(), role); err != nil {
		return nil, err
	}

	return dto.NewRoleResponse(role), nil
}

func (s *roleService) DeleteRole(ctx *fiber.Ctx, id string) error {
	role, err := s.roleRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	if role == nil {
		return errors.New("role not found")
	}

	return s.roleRepo.Delete(ctx.Context(), id)
}

func (s *roleService) AssignRolesToUser(ctx *fiber.Ctx, req *dto.AssignRoleRequest) error {
	// Get user with current roles
	user, err := s.userRepo.GetByID(ctx.Context(), req.UserID)
	if err != nil || user == nil {
		return errors.New("user not found")
	}

	// Get roles
	roles, err := s.roleRepo.GetByIDs(ctx.Context(), req.RoleIDs)
	if err != nil {
		return err
	}

	// Use transaction to assign roles
	return s.db.Transaction(func(tx *gorm.DB) error {
		// Clear existing roles
		if err := tx.Model(user).Association("Roles").Clear(); err != nil {
			return err
		}

		// Assign new roles
		if len(roles) > 0 {
			if err := tx.Model(user).Association("Roles").Append(roles); err != nil {
				return err
			}
		}

		return nil
	})
}

func (s *roleService) GetUserRoles(ctx *fiber.Ctx, userID string) ([]dto.RoleResponse, error) {
	user, err := s.userRepo.GetByID(ctx.Context(), userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Load user roles
	var userWithRoles entities.User
	if err := s.db.Preload("Roles.Permissions").Where("id = ?", userID).First(&userWithRoles).Error; err != nil {
		return nil, err
	}

	responses := make([]dto.RoleResponse, len(userWithRoles.Roles))
	for i, role := range userWithRoles.Roles {
		responses[i] = *dto.NewRoleResponse(&role)
	}

	return responses, nil
}

func (s *roleService) GetAllPermissions(ctx *fiber.Ctx) ([]dto.PermissionResponse, error) {
	permissions, err := s.permissionRepo.GetAll(ctx.Context())
	if err != nil {
		return nil, err
	}

	responses := make([]dto.PermissionResponse, len(permissions))
	for i, perm := range permissions {
		responses[i] = *dto.NewPermissionResponse(&perm)
	}

	return responses, nil
}
