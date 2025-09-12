package jwt

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
	"github.com/tasiuskenways/Scalable-Ecommerce/user-service/internal/config"
	"github.com/tasiuskenways/Scalable-Ecommerce/user-service/internal/domain/entities"
)

const (
	// Redis key prefixes
	refreshTokenPrefix = "refresh:%d"
	accessTokenPrefix  = "access:%d"
	blacklistPrefix    = "blacklist:%s"
)

type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

type TokenManager struct {
	secretKey []byte
	config    *config.JWTConfig
	redis     *redis.Client
}

type Claims struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	jwt.RegisteredClaims
}

func NewTokenManager(cfg *config.JWTConfig, redisClient *redis.Client) (*TokenManager, error) {
	if cfg.PrivateKey == "" {
		return nil, errors.New("JWT secret key is required")
	}

	return &TokenManager{
		secretKey: []byte(cfg.PrivateKey),
		config:    cfg,
		redis:     redisClient,
	}, nil
}

func (tm *TokenManager) GenerateTokenPair(user *entities.User) (map[TokenType]string, error) {
	// Create access token
	accessToken, accessClaims, err := tm.generateToken(user, AccessToken, tm.config.Expiration)
	if err != nil {
		return nil, err
	}

	// Create refresh token
	refreshToken, refreshClaims, err := tm.generateToken(user, RefreshToken, tm.config.RefreshExpiration)
	if err != nil {
		return nil, err
	}

	// Store refresh token in Redis
	err = tm.redis.Set(
		context.Background(),
		fmt.Sprintf("%s%s", refreshTokenPrefix, user.ID),
		refreshToken,
		time.Until(refreshClaims.ExpiresAt.Time),
	).Err()

	if err != nil {
		return nil, fmt.Errorf("failed to store refresh token: %w", err)
	}

	// Store access token in Redis with shorter TTL
	err = tm.redis.Set(
		context.Background(),
		fmt.Sprintf("%s%s", accessTokenPrefix, user.ID),
		accessToken,
		time.Until(accessClaims.ExpiresAt.Time),
	).Err()

	if err != nil {
		return nil, fmt.Errorf("failed to store access token: %w", err)
	}

	return map[TokenType]string{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (tm *TokenManager) generateToken(user *entities.User, tokenType TokenType, expiration time.Duration) (string, *Claims, error) {
	now := time.Now()
	expiresAt := now.Add(expiration)
	claims := &Claims{
		UserID: user.ID,
		Email:  user.Email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "todolistapi",
			Subject:   string(tokenType),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(tm.secretKey)
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}

	return tokenString, claims, nil
}

func (tm *TokenManager) isTokenBlacklisted(tokenString string) (bool, error) {
	isBlacklisted, err := tm.redis.Exists(
		context.Background(),
		fmt.Sprintf(blacklistPrefix, tokenString),
	).Result()

	if err != nil && err != redis.Nil {
		return false, fmt.Errorf("failed to check token status: %w", err)
	}

	return isBlacklisted == 1, nil
}

func (tm *TokenManager) verifyRefreshTokenInRedis(claims *Claims, tokenString string) error {
	tokenInRedis, err := tm.redis.Get(
		context.Background(),
		fmt.Sprintf("%s%s", refreshTokenPrefix, claims.UserID),
	).Result()

	if err != nil || tokenInRedis != tokenString {
		return errors.New("invalid or expired refresh token")
	}

	return nil
}

func (tm *TokenManager) ValidateToken(tokenString string, tokenType TokenType) (*Claims, error) {
	// Check if token is blacklisted
	blacklisted, err := tm.isTokenBlacklisted(tokenString)
	if err != nil {
		return nil, err
	}
	if blacklisted {
		return nil, errors.New("token has been invalidated")
	}

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return tm.secretKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token")
	}

	// Verify token type
	if claims.Subject != string(tokenType) {
		return nil, errors.New("invalid token type")
	}

	// For refresh tokens, verify it exists in Redis
	if tokenType == RefreshToken {
		if err := tm.verifyRefreshTokenInRedis(claims, tokenString); err != nil {
			return nil, err
		}
	}

	return claims, nil
}

// Logout invalidates all tokens for a user
func (tm *TokenManager) Logout(userID string) error {
	ctx := context.Background()

	// Get the refresh token before deleting it
	refreshToken, err := tm.redis.Get(ctx, refreshTokenPrefix+userID).Result()
	if err == nil {
		// Add refresh token to blacklist
		tm.redis.Set(ctx, blacklistPrefix+refreshToken, "1", tm.config.RefreshExpiration)
	}

	// Delete all tokens for this user
	_, err = tm.redis.Del(
		ctx,
		refreshTokenPrefix+userID,
		accessTokenPrefix+userID,
	).Result()

	return err
}

// InvalidateToken adds a token to the blacklist
func (tm *TokenManager) InvalidateToken(tokenString string, tokenType TokenType, expiration time.Duration) error {
	// Add token to blacklist
	return tm.redis.Set(
		context.Background(),
		blacklistPrefix+tokenString,
		"1",
		expiration,
	).Err()
}

func (tm *TokenManager) RefreshToken(refreshToken string) (map[TokenType]string, error) {
	// Validate refresh token
	claims, err := tm.ValidateToken(refreshToken, RefreshToken)
	if err != nil {
		return nil, err
	}

	// Create a new token pair
	user := &entities.User{
		ID:    claims.UserID,
		Email: claims.Email,
	}

	return tm.GenerateTokenPair(user)
}
