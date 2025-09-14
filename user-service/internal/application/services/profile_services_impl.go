package services

import (
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
)

type profileService struct {
	profileRepo repositories.ProfileRepository
	userRepo    repositories.UserRepository
}

func NewProfileService(profileRepo repositories.ProfileRepository, userRepo repositories.UserRepository) services.ProfileService {
	return &profileService{
		profileRepo: profileRepo,
		userRepo:    userRepo,
	}
}

func (s *profileService) CreateProfile(ctx *fiber.Ctx, userID string, req *dto.CreateProfileRequest) (*dto.ProfileResponse, error) {
	// Check if user exists
	user, err := s.userRepo.GetByID(ctx.Context(), userID)
	if err != nil || user == nil {
		return nil, errors.New("user not found")
	}

	// Check if profile already exists
	exists, err := s.profileRepo.ExistsByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("profile already exists for this user")
	}

	profile := &entities.UserProfile{
		UserID:      userID,
		FirstName:   req.FirstName,
		LastName:    req.LastName,
		Phone:       req.Phone,
		Avatar:      req.Avatar,
		DateOfBirth: req.DateOfBirth,
		Gender:      req.Gender,
		Address:     req.Address,
		City:        req.City,
		State:       req.State,
		Country:     req.Country,
		ZipCode:     req.ZipCode,
		Bio:         req.Bio,
	}

	if err := s.profileRepo.Create(ctx.Context(), profile); err != nil {
		return nil, err
	}

	return dto.NewProfileResponse(profile), nil
}

func (s *profileService) GetProfile(ctx *fiber.Ctx, userID string) (*dto.ProfileResponse, error) {
	profile, err := s.profileRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("profile not found")
	}

	return dto.NewProfileResponse(profile), nil
}

func (s *profileService) GetMyProfile(ctx *fiber.Ctx) (*dto.ProfileResponse, error) {
	userID := ctx.Get("X-User-Id")
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	return s.GetProfile(ctx, userID)
}

func (s *profileService) UpdateProfile(ctx *fiber.Ctx, userID string, req *dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	profile, err := s.profileRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return nil, err
	}
	if profile == nil {
		return nil, errors.New("profile not found")
	}

	// Update only non-nil fields
	if req.FirstName != nil {
		profile.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		profile.LastName = *req.LastName
	}
	if req.Phone != nil {
		profile.Phone = *req.Phone
	}
	if req.Avatar != nil {
		profile.Avatar = *req.Avatar
	}
	if req.DateOfBirth != nil {
		profile.DateOfBirth = req.DateOfBirth
	}
	if req.Gender != nil {
		profile.Gender = *req.Gender
	}
	if req.Address != nil {
		profile.Address = *req.Address
	}
	if req.City != nil {
		profile.City = *req.City
	}
	if req.State != nil {
		profile.State = *req.State
	}
	if req.Country != nil {
		profile.Country = *req.Country
	}
	if req.ZipCode != nil {
		profile.ZipCode = *req.ZipCode
	}
	if req.Bio != nil {
		profile.Bio = *req.Bio
	}

	if err := s.profileRepo.Update(ctx.Context(), profile); err != nil {
		return nil, err
	}

	return dto.NewProfileResponse(profile), nil
}

func (s *profileService) UpdateMyProfile(ctx *fiber.Ctx, req *dto.UpdateProfileRequest) (*dto.ProfileResponse, error) {
	userID := ctx.Get("X-User-Id")
	if userID == "" {
		return nil, errors.New("user not authenticated")
	}

	return s.UpdateProfile(ctx, userID, req)
}

func (s *profileService) DeleteProfile(ctx *fiber.Ctx, userID string) error {
	profile, err := s.profileRepo.GetByUserID(ctx.Context(), userID)
	if err != nil {
		return err
	}
	if profile == nil {
		return errors.New("profile not found")
	}

	return s.profileRepo.Delete(ctx.Context(), profile.ID)
}
