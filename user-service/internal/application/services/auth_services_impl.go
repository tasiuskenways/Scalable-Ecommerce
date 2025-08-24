package services

import (
	"errors"
	"log"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"tasius.my.id/SE/user-service/internal/application/dto"
	"tasius.my.id/SE/user-service/internal/config"
	"tasius.my.id/SE/user-service/internal/domain/entities"
	"tasius.my.id/SE/user-service/internal/domain/repositories"
	"tasius.my.id/SE/user-service/internal/domain/services"
	"tasius.my.id/SE/user-service/internal/interfaces/http/middleware"
	"tasius.my.id/SE/user-service/internal/utils/jwt"
	"tasius.my.id/SE/user-service/internal/utils/password"
)

type authService struct {
	userRepo    repositories.UserRepository
	redisClient *redis.Client
	jwtConfig   *config.JWTConfig
	jwtManager  *jwt.TokenManager
}

func NewAuthService(userRepo repositories.UserRepository, redisClient *redis.Client, jwtConfig *config.JWTConfig, jwtManager *jwt.TokenManager) services.AuthService {
	return &authService{
		userRepo:    userRepo,
		redisClient: redisClient,
		jwtConfig:   jwtConfig,
		jwtManager:  jwtManager,
	}
}

// Login implements services.AuthService.
func (s *authService) Login(ctx *fiber.Ctx, req *dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.GetByEmail(ctx.Context(), req.Email)
	if err != nil {
		return nil, err
	}

	log.Println("password", req.Password)
	log.Println("user password", user.Password)

	if err := password.CheckPassword(req.Password, user.Password); err != nil {
		return nil, errors.New("invalid password")
	}

	return s.generateAuthResponse(user)
}

// RefreshToken implements services.AuthService.
func (s *authService) RefreshToken(ctx *fiber.Ctx, refreshToken string) (*dto.AuthResponse, error) {
	token := ctx.Get("Authorization")
	if token == "" {
		return nil, errors.New("Refresh token is required")
	}

	// Remove "Bearer " prefix if present
	token = strings.TrimPrefix(token, "Bearer ")

	tokens, err := s.jwtManager.RefreshToken(token)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		AccessToken:  tokens[jwt.AccessToken],
		RefreshToken: tokens[jwt.RefreshToken],
		TokenType:    "Bearer",
		ExpiresIn:    15 * 60, // 15 minutes in seconds
	}, nil
}

// Register implements services.AuthService.
func (s *authService) Register(ctx *fiber.Ctx, req *dto.RegisterRequest) (*dto.AuthResponse, error) {
	exsist, err := s.userRepo.ExistsByEmail(ctx.Context(), req.Email)
	if err != nil {
		return nil, err
	}
	if exsist {
		return nil, errors.New("email already exists")
	}

	hashPassword, err := password.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &entities.User{
		Email:    req.Email,
		Password: hashPassword,
		Name:     req.Name,
		IsActive: true,
	}

	if err := s.userRepo.Create(ctx.Context(), user); err != nil {
		return nil, errors.New("failed to create user")
	}

	return s.generateAuthResponse(user)
}

// ValidateToken implements services.AuthService.
func (s *authService) ValidateToken(ctx *fiber.Ctx, token string) (*dto.UserResponse, error) {
	claims, err := s.jwtManager.ValidateToken(token, jwt.AccessToken)
	if err != nil {
		return nil, err
	}

	user, err := s.userRepo.GetByID(ctx.Context(), claims.UserID)
	if err != nil {
		return nil, err
	}

	return dto.NewUserResponse(user), nil
}

func (s *authService) Logout(ctx *fiber.Ctx) error {
	userID, err := middleware.GetUserIDFromContext(ctx)
	if err != nil {
		return err
	}

	// Invalidate all tokens for this user
	err = s.jwtManager.Logout(userID)
	if err != nil {
		return err
	}

	return nil
}

func (s *authService) generateAuthResponse(user *entities.User) (*dto.AuthResponse, error) {

	tokens, err := s.jwtManager.GenerateTokenPair(user)
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		User:         dto.NewUserResponse(user),	
		AccessToken:  tokens[jwt.AccessToken],
		RefreshToken: tokens[jwt.RefreshToken],
		TokenType:    "Bearer",
		ExpiresIn:    s.jwtConfig.Expiration.Milliseconds(),
	}, nil
}
