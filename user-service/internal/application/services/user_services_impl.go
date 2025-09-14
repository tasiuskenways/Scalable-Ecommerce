package services

import (
	"errors"
	"math"

	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
	"gorm.io/gorm"
)

type userService struct {
	userRepo repositories.UserRepository
	db       *gorm.DB
}

// NewUserService creates and returns a services.UserService backed by the provided
// UserRepository and GORM DB instance. The returned service uses the repository for
// data access and the DB for queries that require gorm operations.
func NewUserService(userRepo repositories.UserRepository, db *gorm.DB) services.UserService {
	return &userService{
		userRepo: userRepo,
		db:       db,
	}
}

func (s *userService) GetUser(ctx *fiber.Ctx, id string) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	return dto.NewUserResponse(user), nil
}

func (s *userService) GetAllUsers(ctx *fiber.Ctx, page, limit int) (*dto.PaginatedResponse, error) {
	offset := (page - 1) * limit

	var users []entities.User
	var totalCount int64

	// Get total count
	if err := s.db.WithContext(ctx.Context()).
		Model(&entities.User{}).
		Count(&totalCount).Error; err != nil {
		return nil, err
	}

	// Get users with pagination
	if err := s.db.WithContext(ctx.Context()).
		Preload("Roles").
		Preload("Profile").
		Offset(offset).
		Limit(limit).
		Find(&users).Error; err != nil {
		return nil, err
	}

	// Convert to DTOs
	userResponses := make([]dto.UserListResponse, len(users))
	for i, user := range users {
		userResponses[i] = *dto.NewUserListResponse(&user)
	}

	totalPages := int(math.Ceil(float64(totalCount) / float64(limit)))
	hasNext := page < totalPages
	hasPrev := page > 1

	return &dto.PaginatedResponse{
		Data:       userResponses,
		Page:       page,
		Limit:      limit,
		TotalCount: totalCount,
		TotalPages: totalPages,
		HasNext:    hasNext,
		HasPrev:    hasPrev,
	}, nil
}

func (s *userService) UpdateUser(ctx *fiber.Ctx, id string, req *dto.UpdateUserRequest) (*dto.UserResponse, error) {
	user, err := s.userRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Update fields
	if req.Name != nil {
		user.Name = *req.Name
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}

	if err := s.userRepo.Update(ctx.Context(), user); err != nil {
		return nil, err
	}

	return dto.NewUserResponse(user), nil
}

func (s *userService) DeleteUser(ctx *fiber.Ctx, id string) error {
	user, err := s.userRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return err
	}
	if user == nil {
		return errors.New("user not found")
	}

	return s.userRepo.Delete(ctx.Context(), id)
}

func (s *userService) GetUserRBACInfo(ctx *fiber.Ctx, id string) (*dto.UserRBACResponse, error) {
	user, err := s.userRepo.GetByID(ctx.Context(), id)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, errors.New("user not found")
	}

	// Extract role names
	roles := make([]string, len(user.Roles))
	for i, role := range user.Roles {
		roles[i] = role.Name
	}

	// Get all permissions
	permissions := user.GetPermissions()

	return &dto.UserRBACResponse{
		UserID:      user.ID,
		Roles:       roles,
		Permissions: permissions,
	}, nil
}
