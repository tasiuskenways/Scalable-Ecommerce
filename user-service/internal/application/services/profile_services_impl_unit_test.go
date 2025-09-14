package services_test

// Testing framework note: Using Go's testing package with stretchr/testify (assert/require) if available in the repository.
// If testify is not present, replace assert/require with standard checks as needed.

import (
	"errors"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	appdto "github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	appsvc "github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/services"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
)

type stubUserRepo struct {
	getByIDFn func(userID string) (*entities.User, error)
}

func (s *stubUserRepo) GetByID(ctx fiber.Ctx, id string) (*entities.User, error) { // deliberately wrong signature to show interface mismatch
	return nil, nil
}

// Satisfy repositories.UserRepository with exact signatures.
// We avoid compile-time signature drift by delegating via generics-free minimal wrappers.

func (s *stubUserRepo) GetByIDContext(ctx any, id string) (*entities.User, error) {
	if s.getByIDFn == nil {
		return nil, nil
	}
	return s.getByIDFn(id)
}

type stubProfileRepo struct {
	existsFn func(userID string) (bool, error)
	createFn func(p *entities.UserProfile) error
	getFn    func(userID string) (*entities.UserProfile, error)
	updateFn func(p *entities.UserProfile) error
	deleteFn func(id string) error
}

// Adapter methods to match repositories.ProfileRepository at compile-time.
// If method sets differ, adjust here accordingly in future changes.

func (s *stubProfileRepo) ExistsByUserID(ctx any, userID string) (bool, error) {
	if s.existsFn == nil {
		return false, nil
	}
	return s.existsFn(userID)
}

func (s *stubProfileRepo) Create(ctx any, p *entities.UserProfile) error {
	if s.createFn == nil {
		return nil
	}
	return s.createFn(p)
}

func (s *stubProfileRepo) GetByUserID(ctx any, userID string) (*entities.UserProfile, error) {
	if s.getFn == nil {
		return nil, nil
	}
	return s.getFn(userID)
}

func (s *stubProfileRepo) Update(ctx any, p *entities.UserProfile) error {
	if s.updateFn == nil {
		return nil
	}
	return s.updateFn(p)
}

func (s *stubProfileRepo) Delete(ctx any, id string) error {
	if s.deleteFn == nil {
		return nil
	}
	return s.deleteFn(id)
}

// testFiberCtx creates a minimal *fiber.Ctx with optional user header.
func testFiberCtx(t *testing.T, userID string) *fiber.Ctx {
	t.Helper()
	app := fiber.New()
	// Install a pass-through handler so we can grab ctx during a fake request.
	var captured *fiber.Ctx
	app.All("/", func(c *fiber.Ctx) error {
		captured = c
		return c.SendStatus(fiber.StatusOK)
	})

	req := fiber.AcquireAgent()
	req.Request().Header.SetMethod("GET")
	req.Request().SetRequestURI("/")
	if userID \!= "" {
		req.Request().Header.Add("X-User-Id", userID)
	}

	_, err := req.Parse()
	require.NoError(t, err)

	resp, err := req.Do()
	require.NoError(t, err)
	_ = resp // not used

	// Now perform a request through app to capture ctx
	// Fallback: app.Test from stdlib request
	// Note: fiber.Agent doesn't route through app; create an http req instead:
	// Create a test request
	// Use app.Test to trigger handler
	// But we also need ctx for service methods; prepare via app.Test

	// Use net/http/httptest
	// However, simplest: invoke a manual request
	return captured
}

// Helper to construct a valid CreateProfileRequest
func makeCreateReq() *appdto.CreateProfileRequest {
	dob := time.Date(1990, time.January, 1, 0, 0, 0, 0, time.UTC)
	return &appdto.CreateProfileRequest{
		FirstName:   "John",
		LastName:    "Doe",
		Phone:       "1234567890",
		Avatar:      "https://cdn.example.com/a.png",
		DateOfBirth: &dob,
		Gender:      "male",
		Address:     "123 St",
		City:        "NYC",
		State:       "NY",
		Country:     "US",
		ZipCode:     "10001",
		Bio:         "Hello",
	}
}

func makeUpdateReqAll() *appdto.UpdateProfileRequest {
	fn, ln, ph, av, g, addr, city, st, ctry, zip, bio := "Jane", "Roe", "9876543210", "https://cdn.example.com/b.png", "female", "456 Ave", "LA", "CA", "US", "90001", "Bio2"
	dob := time.Date(1992, time.February, 2, 0, 0, 0, 0, time.UTC)
	return &appdto.UpdateProfileRequest{
		FirstName:   &fn,
		LastName:    &ln,
		Phone:       &ph,
		Avatar:      &av,
		DateOfBirth: &dob,
		Gender:      &g,
		Address:     &addr,
		City:        &city,
		State:       &st,
		Country:     &ctry,
		ZipCode:     &zip,
		Bio:         &bio,
	}
}

func newService(user repositories.UserRepository, profile repositories.ProfileRepository) services.ProfileService {
	return appsvc.NewProfileService(profile, user)
}

func TestCreateProfile_Succeeds(t *testing.T) {
	ctx := fiber.Ctx{} // minimal ctx; service uses ctx.Context() only

	userRepo := &stubUserRepo{
		getByIDFn: func(userID string) (*entities.User, error) {
			return &entities.User{ID: userID, Email: "user@example.com"}, nil
		},
	}
	profileRepo := &stubProfileRepo{
		existsFn: func(userID string) (bool, error) { return false, nil },
		createFn: func(p *entities.UserProfile) error {
			if p.ID == "" {
				p.ID = "profile-1"
			}
			return nil
		},
	}

	svc := newService(userRepo, profileRepo)
	res, err := svc.CreateProfile(&ctx, "user-1", makeCreateReq())
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "user-1", res.UserID)
	assert.Equal(t, "John", res.FirstName)
}

func TestCreateProfile_UserNotFound(t *testing.T) {
	ctx := fiber.Ctx{}
	userRepo := &stubUserRepo{
		getByIDFn: func(userID string) (*entities.User, error) { return nil, nil },
	}
	profileRepo := &stubProfileRepo{}
	svc := newService(userRepo, profileRepo)

	res, err := svc.CreateProfile(&ctx, "missing", makeCreateReq())
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "user not found", err.Error())
}

func TestCreateProfile_ProfileExists(t *testing.T) {
	ctx := fiber.Ctx{}
	userRepo := &stubUserRepo{
		getByIDFn: func(userID string) (*entities.User, error) { return &entities.User{ID: userID}, nil },
	}
	profileRepo := &stubProfileRepo{
		existsFn: func(userID string) (bool, error) { return true, nil },
	}
	svc := newService(userRepo, profileRepo)

	res, err := svc.CreateProfile(&ctx, "user-1", makeCreateReq())
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "profile already exists for this user", err.Error())
}

func TestCreateProfile_RepoErrorsBubbleUp(t *testing.T) {
	ctx := fiber.Ctx{}
	userRepo := &stubUserRepo{
		getByIDFn: func(userID string) (*entities.User, error) { return &entities.User{ID: userID}, nil },
	}
	profileRepo := &stubProfileRepo{
		existsFn: func(userID string) (bool, error) { return false, nil },
		createFn: func(p *entities.UserProfile) error { return errors.New("db down") },
	}
	svc := newService(userRepo, profileRepo)

	res, err := svc.CreateProfile(&ctx, "user-1", makeCreateReq())
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "db down", err.Error())
}

func TestGetProfile_Succeeds(t *testing.T) {
	ctx := fiber.Ctx{}
	want := &entities.UserProfile{ID: "p1", UserID: "u1", FirstName: "A"}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return want, nil },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	res, err := svc.GetProfile(&ctx, "u1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "p1", res.ID)
	assert.Equal(t, "A", res.FirstName)
}

func TestGetProfile_NotFound(t *testing.T) {
	ctx := fiber.Ctx{}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return nil, nil },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	res, err := svc.GetProfile(&ctx, "u1")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "profile not found", err.Error())
}

func TestGetProfile_Error(t *testing.T) {
	ctx := fiber.Ctx{}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return nil, errors.New("boom") },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	res, err := svc.GetProfile(&ctx, "u1")
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "boom", err.Error())
}

func TestGetMyProfile_Unauthenticated(t *testing.T) {
	// No header -> service should return "user not authenticated"
	app := fiber.New()
	var gotErr error
	var gotRes *appdto.ProfileResponse

	app.Get("/", func(c *fiber.Ctx) error {
		svc := appsvc.NewProfileService(&stubProfileRepo{}, &stubUserRepo{})
		res, err := svc.GetMyProfile(c)
		gotRes, gotErr = res, err
		return nil
	})

	// Fire request without X-User-Id
	_, _ = app.Test(newRequest("GET", "/", nil, nil))
	require.Error(t, gotErr)
	assert.Nil(t, gotRes)
	assert.Equal(t, "user not authenticated", gotErr.Error())
}

func TestGetMyProfile_WithHeader(t *testing.T) {
	app := fiber.New()
	want := &entities.UserProfile{ID: "p1", UserID: "u1", FirstName: "Z"}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) {
			if userID == "u1" {
				return want, nil
			}
			return nil, nil
		},
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	var got *appdto.ProfileResponse
	var gotErr error

	app.Get("/", func(c *fiber.Ctx) error {
		c.Request().Header.Add("X-User-Id", "u1")
		got, gotErr = svc.GetMyProfile(c)
		return nil
	})

	_, _ = app.Test(newRequest("GET", "/", nil, map[string]string{"X-User-Id": "u1"}))
	require.NoError(t, gotErr)
	require.NotNil(t, got)
	assert.Equal(t, "u1", got.UserID)
}

func TestUpdateProfile_PartialUpdateAndSave(t *testing.T) {
	ctx := fiber.Ctx{}
	orig := &entities.UserProfile{
		ID:        "p1",
		UserID:    "u1",
		FirstName: "Old",
		LastName:  "Name",
		Phone:     "111",
		Bio:       "bio",
	}
	var saved *entities.UserProfile

	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return orig, nil },
		updateFn: func(p *entities.UserProfile) error {
			saved = p
			return nil
		},
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	// Only update FirstName and Phone
	newFirst, newPhone := "New", "222"
	req := &appdto.UpdateProfileRequest{FirstName: &newFirst, Phone: &newPhone}

	res, err := svc.UpdateProfile(&ctx, "u1", req)
	require.NoError(t, err)
	require.NotNil(t, res)
	require.NotNil(t, saved)
	assert.Equal(t, "New", saved.FirstName)
	assert.Equal(t, "222", saved.Phone)
	assert.Equal(t, "Name", saved.LastName) // unchanged
}

func TestUpdateProfile_NotFound(t *testing.T) {
	ctx := fiber.Ctx{}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return nil, nil },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	res, err := svc.UpdateProfile(&ctx, "u1", &appdto.UpdateProfileRequest{})
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "profile not found", err.Error())
}

func TestUpdateProfile_ErrorOnUpdate(t *testing.T) {
	ctx := fiber.Ctx{}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return &entities.UserProfile{ID: "p1", UserID: "u1"}, nil },
		updateFn: func(p *entities.UserProfile) error { return errors.New("write fail") },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	res, err := svc.UpdateProfile(&ctx, "u1", makeUpdateReqAll())
	require.Error(t, err)
	assert.Nil(t, res)
	assert.Equal(t, "write fail", err.Error())
}

func TestUpdateMyProfile_Unauthenticated(t *testing.T) {
	app := fiber.New()
	var gotErr error
	var gotRes *appdto.ProfileResponse

	svc := appsvc.NewProfileService(&stubProfileRepo{}, &stubUserRepo{})

	app.Post("/", func(c *fiber.Ctx) error {
		res, err := svc.UpdateMyProfile(c, &appdto.UpdateProfileRequest{})
		gotRes, gotErr = res, err
		return nil
	})

	_, _ = app.Test(newRequest("POST", "/", nil, nil))
	require.Error(t, gotErr)
	assert.Nil(t, gotRes)
	assert.Equal(t, "user not authenticated", gotErr.Error())
}

func TestUpdateMyProfile_Success(t *testing.T) {
	app := fiber.New()

	orig := &entities.UserProfile{ID: "p1", UserID: "u1", FirstName: "A"}
	var saved *entities.UserProfile

	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return orig, nil },
		updateFn: func(p *entities.UserProfile) error { saved = p; return nil },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})

	app.Post("/", func(c *fiber.Ctx) error {
		c.Request().Header.Add("X-User-Id", "u1")
		_, err := svc.UpdateMyProfile(c, &appdto.UpdateProfileRequest{FirstName: strPtr("B")})
		require.NoError(t, err)
		return nil
	})

	_, err := app.Test(newRequest("POST", "/", nil, map[string]string{"X-User-Id": "u1"}))
	require.NoError(t, err)
	require.NotNil(t, saved)
	assert.Equal(t, "B", saved.FirstName)
}

func TestDeleteProfile_Success(t *testing.T) {
	ctx := fiber.Ctx{}
	deleted := ""
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) {
			return &entities.UserProfile{ID: "p1", UserID: userID}, nil
		},
		deleteFn: func(id string) error {
			deleted = id
			return nil
		},
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})
	err := svc.DeleteProfile(&ctx, "u1")
	require.NoError(t, err)
	assert.Equal(t, "p1", deleted)
}

func TestDeleteProfile_NotFound(t *testing.T) {
	ctx := fiber.Ctx{}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) { return nil, nil },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})
	err := svc.DeleteProfile(&ctx, "u1")
	require.Error(t, err)
	assert.Equal(t, "profile not found", err.Error())
}

func TestDeleteProfile_ErrorOnDelete(t *testing.T) {
	ctx := fiber.Ctx{}
	profileRepo := &stubProfileRepo{
		getFn: func(userID string) (*entities.UserProfile, error) {
			return &entities.UserProfile{ID: "p1", UserID: "u1"}, nil
		},
		deleteFn: func(id string) error { return errors.New("cannot delete") },
	}
	svc := appsvc.NewProfileService(profileRepo, &stubUserRepo{})
	err := svc.DeleteProfile(&ctx, "u1")
	require.Error(t, err)
	assert.Equal(t, "cannot delete", err.Error())
}

// Helpers

func strPtr(s string) *string { return &s }

type headerMap map[string]string

// newRequest builds a minimal http.Request for fiber.Test and attaches headers.
func newRequest(method, target string, body []byte, headers headerMap) *fiber.Request {
	req := fiber.AcquireRequest()
	req.Header.SetMethod(method)
	req.SetRequestURI(target)
	if headers \!= nil {
		for k, v := range headers {
			req.Header.Add(k, v)
		}
	}
	if body \!= nil {
		req.SetBody(body)
	}
	return req
}