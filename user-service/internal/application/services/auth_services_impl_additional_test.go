package services

import (
	"errors"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/config"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/utils/jwt"
)

//
// Lightweight fakes/mocks for dependencies (no new libraries introduced)
//

type fakeUserRepo struct {
	getByEmailFunc     func(email string) (*entities.User, error)
	getByIDFunc        func(id string) (*entities.User, error)
	existsByEmailFunc  func(email string) (bool, error)
	createFunc         func(u *entities.User) error
}

func (f *fakeUserRepo) GetByEmail(_ interface{}, email string) (*entities.User, error) {
	if f.getByEmailFunc \!= nil {
		return f.getByEmailFunc(email)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepo) GetByID(_ interface{}, id string) (*entities.User, error) {
	if f.getByIDFunc \!= nil {
		return f.getByIDFunc(id)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeUserRepo) ExistsByEmail(_ interface{}, email string) (bool, error) {
	if f.existsByEmailFunc \!= nil {
		return f.existsByEmailFunc(email)
	}
	return false, errors.New("not implemented")
}

func (f *fakeUserRepo) Create(_ interface{}, u *entities.User) error {
	if f.createFunc \!= nil {
		return f.createFunc(u)
	}
	return errors.New("not implemented")
}

// Minimal TokenManager stub conforming to methods used by service.
// We avoid importing real implementation details; we simulate behavior deterministically.
type fakeTokenManager struct {
	generatePairFunc func(user *entities.User) (map[string]string, error)
	validateFunc     func(token string, tokenType string) (*jwt.Claims, error)
	refreshFunc      func(refreshToken string) (map[string]string, error)
	logoutFunc       func(userID string) error
}

func (f *fakeTokenManager) GenerateTokenPair(user *entities.User) (map[string]string, error) {
	if f.generatePairFunc \!= nil {
		return f.generatePairFunc(user)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeTokenManager) ValidateToken(token string, tokenType string) (*jwt.Claims, error) {
	if f.validateFunc \!= nil {
		return f.validateFunc(token, tokenType)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeTokenManager) RefreshToken(refreshToken string) (map[string]string, error) {
	if f.refreshFunc \!= nil {
		return f.refreshFunc(refreshToken)
	}
	return nil, errors.New("not implemented")
}

func (f *fakeTokenManager) Logout(userID string) error {
	if f.logoutFunc \!= nil {
		return f.logoutFunc(userID)
	}
	return errors.New("not implemented")
}

// Helpers
func newFiberCtxWithHeaders(headers map[string]string) (*fiber.App, *fiber.Ctx) {
	app := fiber.New()
	// Create a test route that just returns 200 so app.Test builds a context we can reuse.
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendStatus(200)
	})
	req := httptest.NewRequest("GET", "/", nil)
	for k, v := range headers {
		req.Header.Set(k, v)
	}
	resp, _ := app.Test(req, -1)
	_ = resp
	// Build another context by invoking a no-op handler to capture ctx
	var captured *fiber.Ctx
	app.Get("/capture", func(c *fiber.Ctx) error {
		captured = c
		return c.SendStatus(204)
	})
	req2 := httptest.NewRequest("GET", "/capture", nil)
	for k, v := range headers {
		req2.Header.Set(k, v)
	}
	_, _ = app.Test(req2, -1)
	return app, captured
}

func mustJWTConfig(exp time.Duration) *config.JWTConfig {
	return &config.JWTConfig{
		Secret:     "test-secret",
		Expiration: exp,
	}
}

//
// Tests
//

func TestNewAuthService_ConstructsInstance(t *testing.T) {
	svc := NewAuthService(&fakeUserRepo{}, &redis.Client{}, mustJWTConfig(15*time.Minute), &fakeTokenManager{})
	if svc == nil {
		t.Fatal("expected non-nil auth service")
	}
}

func TestAuthService_Login_Success(t *testing.T) {
	// Arrange
	user := &entities.User{
		ID:       "u1",
		Email:    "user@example.com",
		Password: "$2a$10$hashmatches", // We won't invoke real bcrypt; CheckPassword is called in impl.
		Name:     "User",
		IsActive: true,
	}
	repo := &fakeUserRepo{
		getByEmailFunc: func(email string) (*entities.User, error) {
			if email \!= user.Email {
				return nil, errors.New("not found")
			}
			return &entities.User{
				ID:       user.ID,
				Email:    email,
				Password: user.Password,
				Name:     user.Name,
				IsActive: true,
			}, nil
		},
	}
	tokens := map[string]string{
		jwt.AccessToken:  "access-123",
		jwt.RefreshToken: "refresh-123",
	}
	fm := &fakeTokenManager{
		generatePairFunc: func(u *entities.User) (map[string]string, error) {
			if u.Email \!= user.Email {
				return nil, errors.New("wrong user")
			}
			return tokens, nil
		},
	}
	s := &authService{
		userRepo:    repo,
		redisClient: &redis.Client{},
		jwtConfig:   mustJWTConfig(30 * time.Minute),
		jwtManager:  (*jwt.TokenManager)(fm), // type-compat: using same method set
	}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	// We bypass real password hashing by stubbing password.CheckPassword via the stored hash being same as input.
	// Since we cannot patch functions easily without extra libs, use the exact same string for input and stored "hash".
	req := &dto.LoginRequest{Email: user.Email, Password: user.Password}

	// Act
	resp, err := s.Login(c, req)

	// Assert
	if err \!= nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response, got nil")
	}
	if resp.AccessToken \!= tokens[jwt.AccessToken] || resp.RefreshToken \!= tokens[jwt.RefreshToken] {
		t.Errorf("unexpected tokens: %#v", resp)
	}
	if resp.User == nil || resp.User.Email \!= user.Email {
		t.Errorf("expected user email %s, got %#v", user.Email, resp.User)
	}
}

func TestAuthService_Login_InvalidPassword(t *testing.T) {
	// Arrange: return a user with different "hash" so CheckPassword fails.
	repo := &fakeUserRepo{
		getByEmailFunc: func(email string) (*entities.User, error) {
			return &entities.User{ID: "u2", Email: email, Password: "stored-hash"}, nil
		},
	}
	fm := &fakeTokenManager{}
	s := &authService{
		userRepo:    repo,
		redisClient: &redis.Client{},
		jwtConfig:   mustJWTConfig(10 * time.Minute),
		jwtManager:  (*jwt.TokenManager)(fm),
	}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	req := &dto.LoginRequest{Email: "x@example.com", Password: "different-plain"}

	// Act
	resp, err := s.Login(c, req)

	// Assert
	if err == nil {
		t.Fatalf("expected error, got nil and resp=%#v", resp)
	}
	if \!strings.Contains(err.Error(), "invalid password") {
		t.Errorf("expected invalid password error, got %v", err)
	}
}

func TestAuthService_RefreshToken_MissingHeader(t *testing.T) {
	s := &authService{
		userRepo:    &fakeUserRepo{},
		redisClient: &redis.Client{},
		jwtConfig:   mustJWTConfig(15 * time.Minute),
		jwtManager:  (*jwt.TokenManager)(&fakeTokenManager{}),
	}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	resp, err := s.RefreshToken(c, "")
	if err == nil || \!strings.Contains(err.Error(), "Refresh token is required") {
		t.Fatalf("expected missing header error, got resp=%#v err=%v", resp, err)
	}
}

func TestAuthService_RefreshToken_SuccessWithBearerPrefix(t *testing.T) {
	tokens := map[string]string{
		jwt.AccessToken:  "access-new",
		jwt.RefreshToken: "refresh-new",
	}
	fm := &fakeTokenManager{
		refreshFunc: func(refreshToken string) (map[string]string, error) {
			if refreshToken \!= "r-abc" {
				return nil, errors.New("bad token")
			}
			return tokens, nil
		},
	}
	s := &authService{
		userRepo:    &fakeUserRepo{},
		redisClient: &redis.Client{},
		jwtConfig:   mustJWTConfig(15 * time.Minute),
		jwtManager:  (*jwt.TokenManager)(fm),
	}
	app, c := newFiberCtxWithHeaders(map[string]string{"Authorization": "Bearer r-abc"})
	defer app.Shutdown()

	resp, err := s.RefreshToken(c, "")
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken \!= "access-new" || resp.RefreshToken \!= "refresh-new" {
		t.Fatalf("unexpected tokens: %#v", resp)
	}
	// ExpiresIn in RefreshToken is hard-coded 15*60 seconds in service
	if resp.ExpiresIn \!= 15*60 {
		t.Errorf("expected ExpiresIn=900, got %d", resp.ExpiresIn)
	}
}

func TestAuthService_Register_EmailExists(t *testing.T) {
	repo := &fakeUserRepo{
		existsByEmailFunc: func(email string) (bool, error) {
			return true, nil
		},
	}
	s := &authService{
		userRepo:    repo,
		redisClient: &redis.Client{},
		jwtConfig:   mustJWTConfig(1 * time.Hour),
		jwtManager:  (*jwt.TokenManager)(&fakeTokenManager{}),
	}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	_, err := s.Register(c, &dto.RegisterRequest{Email: "exists@example.com", Password: "p", Name: "N"})
	if err == nil || \!strings.Contains(err.Error(), "email already exists") {
		t.Fatalf("expected email exists error, got %v", err)
	}
}

func TestAuthService_Register_CreateFails(t *testing.T) {
	repo := &fakeUserRepo{
		existsByEmailFunc: func(email string) (bool, error) { return false, nil },
		createFunc:        func(u *entities.User) error { return errors.New("db fail") },
	}
	s := &authService{
		userRepo:    repo,
		redisClient: &redis.Client{},
		jwtConfig:   mustJWTConfig(30 * time.Minute),
		jwtManager:  (*jwt.TokenManager)(&fakeTokenManager{}),
	}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	_, err := s.Register(c, &dto.RegisterRequest{Email: "ok@example.com", Password: "pw", Name: "Ok"})
	if err == nil || \!strings.Contains(err.Error(), "failed to create user") {
		t.Fatalf("expected wrapped create error, got %v", err)
	}
}

func TestAuthService_Register_Success_GeneratesTokens(t *testing.T) {
	repo := &fakeUserRepo{
		existsByEmailFunc: func(email string) (bool, error) { return false, nil },
		createFunc:        func(u *entities.User) error { return nil },
	}
	tokens := map[string]string{
		jwt.AccessToken:  "acc-777",
		jwt.RefreshToken: "ref-777",
	}
	fm := &fakeTokenManager{
		generatePairFunc: func(user *entities.User) (map[string]string, error) {
			if user.Email == "" || user.Password == "" {
				return nil, errors.New("invalid user")
			}
			return tokens, nil
		},
	}
	cfg := mustJWTConfig(45 * time.Minute)
	s := &authService{
		userRepo:    repo,
		redisClient: &redis.Client{},
		jwtConfig:   cfg,
		jwtManager:  (*jwt.TokenManager)(fm),
	}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	resp, err := s.Register(c, &dto.RegisterRequest{Email: "new@example.com", Password: "pw", Name: "New"})
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.AccessToken \!= tokens[jwt.AccessToken] || resp.RefreshToken \!= tokens[jwt.RefreshToken] {
		t.Fatalf("unexpected tokens: %#v", resp)
	}
	if resp.TokenType \!= "Bearer" {
		t.Errorf("expected token type Bearer, got %s", resp.TokenType)
	}
	// ExpiresIn uses cfg.Expiration.Milliseconds() in generateAuthResponse
	want := cfg.Expiration.Milliseconds()
	if resp.ExpiresIn \!= want {
		t.Errorf("expected ExpiresIn=%d, got %d", want, resp.ExpiresIn)
	}
}

func TestAuthService_ValidateToken_Success(t *testing.T) {
	fm := &fakeTokenManager{
		validateFunc: func(token string, tokenType string) (*jwt.Claims, error) {
			if tokenType \!= jwt.AccessToken {
				return nil, errors.New("wrong type")
			}
			return &jwt.Claims{UserID: "uid-9"}, nil
		},
	}
	repo := &fakeUserRepo{
		getByIDFunc: func(id string) (*entities.User, error) {
			if id \!= "uid-9" {
				return nil, errors.New("not found")
			}
			return &entities.User{ID: id, Email: "u9@example.com", Name: "Nine", IsActive: true}, nil
		},
	}
	s := &authService{userRepo: repo, jwtManager: (*jwt.TokenManager)(fm), jwtConfig: mustJWTConfig(15 * time.Minute)}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	resp, err := s.ValidateToken(c, "token-abc")
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil || resp.Email \!= "u9@example.com" {
		t.Fatalf("unexpected user resp: %#v", resp)
	}
}

func TestAuthService_ValidateToken_FailsOnInvalid(t *testing.T) {
	fm := &fakeTokenManager{
		validateFunc: func(token string, tokenType string) (*jwt.Claims, error) {
			return nil, errors.New("invalid token")
		},
	}
	s := &authService{userRepo: &fakeUserRepo{}, jwtManager: (*jwt.TokenManager)(fm), jwtConfig: mustJWTConfig(1 * time.Minute)}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	resp, err := s.ValidateToken(c, "bad")
	if err == nil || \!strings.Contains(err.Error(), "invalid token") {
		t.Fatalf("expected validation error, got resp=%#v err=%v", resp, err)
	}
}

func TestAuthService_Logout_MissingUserIDHeader(t *testing.T) {
	s := &authService{jwtManager: (*jwt.TokenManager)(&fakeTokenManager{})}
	app, c := newFiberCtxWithHeaders(nil)
	defer app.Shutdown()

	err := s.Logout(c)
	if err == nil || \!strings.Contains(err.Error(), "user not authenticated") {
		t.Fatalf("expected missing user header error, got %v", err)
	}
}

func TestAuthService_Logout_Success(t *testing.T) {
	called := false
	fm := &fakeTokenManager{
		logoutFunc: func(userID string) error {
			if userID \!= "U-123" {
				return errors.New("bad id")
			}
			called = true
			return nil
		},
	}
	s := &authService{jwtManager: (*jwt.TokenManager)(fm)}
	app, c := newFiberCtxWithHeaders(map[string]string{"X-User-Id": "U-123"})
	defer app.Shutdown()

	err := s.Logout(c)
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if \!called {
		t.Fatalf("expected logout to be called")
	}
}

func TestGenerateAuthResponse_PropagatesErrors(t *testing.T) {
	u := &entities.User{ID: "x", Email: "x@example.com"}
	fm := &fakeTokenManager{
		generatePairFunc: func(user *entities.User) (map[string]string, error) {
			return nil, errors.New("generation failed")
		},
	}
	s := &authService{jwtManager: (*jwt.TokenManager)(fm), jwtConfig: mustJWTConfig(5 * time.Minute)}

	resp, err := s.generateAuthResponse(u)
	if err == nil || \!strings.Contains(err.Error(), "generation failed") {
		t.Fatalf("expected error, got resp=%#v err=%v", resp, err)
	}
}

func TestGenerateAuthResponse_SetsFieldsFromConfigAndDTO(t *testing.T) {
	u := &entities.User{ID: "y", Email: "y@example.com", Name: "Y"}
	cfg := mustJWTConfig(12 * time.Minute)
	fm := &fakeTokenManager{
		generatePairFunc: func(user *entities.User) (map[string]string, error) {
			return map[string]string{
				jwt.AccessToken:  "a-token",
				jwt.RefreshToken: "r-token",
			}, nil
		},
	}
	s := &authService{jwtManager: (*jwt.TokenManager)(fm), jwtConfig: cfg}

	resp, err := s.generateAuthResponse(u)
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.TokenType \!= "Bearer" {
		t.Errorf("expected Bearer, got %s", resp.TokenType)
	}
	if resp.ExpiresIn \!= cfg.Expiration.Milliseconds() {
		t.Errorf("expires mismatch: got %d want %d", resp.ExpiresIn, cfg.Expiration.Milliseconds())
	}
	if resp.User == nil || resp.User.Email \!= u.Email {
		t.Errorf("user mapping incorrect: %#v", resp.User)
	}
	if resp.AccessToken == "" || resp.RefreshToken == "" {
		t.Errorf("expected non-empty tokens: %#v", resp)
	}
}