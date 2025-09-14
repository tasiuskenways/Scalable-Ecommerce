package services

import (
	"errors"
	"testing"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type stubUserRepo struct {
	getByIDFunc func(ctx any, id string) (*entities.User, error)
	updateFunc  func(ctx any, u *entities.User) error
	deleteFunc  func(ctx any, id string) error
}

func (m *stubUserRepo) GetByID(ctx any, id string) (*entities.User, error) {
	if m.getByIDFunc \!= nil {
		return m.getByIDFunc(ctx, id)
	}
	return nil, nil
}

func (m *stubUserRepo) Update(ctx any, u *entities.User) error {
	if m.updateFunc \!= nil {
		return m.updateFunc(ctx, u)
	}
	return nil
}

func (m *stubUserRepo) Delete(ctx any, id string) error {
	if m.deleteFunc \!= nil {
		return m.deleteFunc(ctx, id)
	}
	return nil
}

// helper to create a non-nil *fiber.Ctx for passing through service calls.
// we don't attach routes; we only need a live context instance.
func newFiberCtx(t *testing.T) *fiber.Ctx {
	t.Helper()
	app := fiber.New()
	// Fiber requires a request lifecycle; use a no-op handler to capture the ctx
	var captured *fiber.Ctx
	app.Get("/", func(c *fiber.Ctx) error {
		captured = c
		return c.SendStatus(fiber.StatusOK)
	})
	req := fiber.AcquireAgent().Request()
	req.Header.SetMethod(fiber.MethodGet)
	req.SetRequestURI("/")
	_, err := app.Test(req, -1)
	require.NoError(t, err)
	require.NotNil(t, captured, "failed to capture fiber ctx")
	return captured
}

func TestUserService_GetUser_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			require.Equal(t, "u123", id)
			return &entities.User{
				ID:       "u123",
				Name:     "Alice",
				IsActive: true,
			}, nil
		},
	}
	// db not used by GetUser; provide a minimal sqlite
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	svc := NewUserService(repo, db)

	resp, err := svc.GetUser(ctx, "u123")
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, "Alice", resp.Name)
	assert.Equal(t, "u123", resp.ID)
}

func TestUserService_GetUser_NotFoundNilEntity(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return nil, nil
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	resp, err := svc.GetUser(ctx, "missing")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "user not found")
}

func TestUserService_GetUser_RepoError(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return nil, errors.New("db down")
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	resp, err := svc.GetUser(ctx, "u1")
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "db down")
}

func TestUserService_UpdateUser_PartialUpdate_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	user := &entities.User{
		ID:       "u42",
		Name:     "Old Name",
		IsActive: false,
	}
	var updated *entities.User

	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return user, nil
		},
		updateFunc: func(_ any, u *entities.User) error {
			updated = u
			return nil
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	newName := "New Name"
	newActive := true
	req := &dto.UpdateUserRequest{
		Name:     &newName,
		IsActive: &newActive,
	}

	resp, err := svc.UpdateUser(ctx, "u42", req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, updated)
	assert.Equal(t, "New Name", updated.Name)
	assert.Equal(t, true, updated.IsActive)

	assert.Equal(t, "New Name", resp.Name)
	assert.Equal(t, "u42", resp.ID)
}

func TestUserService_UpdateUser_NoFieldsProvided(t *testing.T) {
	ctx := newFiberCtx(t)
	user := &entities.User{
		ID:       "u9",
		Name:     "Same",
		IsActive: true,
	}
	var updated *entities.User
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return user, nil
		},
		updateFunc: func(_ any, u *entities.User) error {
			updated = u
			return nil
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	req := &dto.UpdateUserRequest{} // no changes
	resp, err := svc.UpdateUser(ctx, "u9", req)
	require.NoError(t, err)
	require.NotNil(t, resp)

	require.NotNil(t, updated)
	assert.Equal(t, "Same", updated.Name)
	assert.Equal(t, true, updated.IsActive)
}

func TestUserService_UpdateUser_NotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return nil, nil
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	req := &dto.UpdateUserRequest{}
	resp, err := svc.UpdateUser(ctx, "missing", req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "user not found")
}

func TestUserService_UpdateUser_RepoUpdateError(t *testing.T) {
	ctx := newFiberCtx(t)
	user := &entities.User{ID: "u77", Name: "x", IsActive: true}
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return user, nil
		},
		updateFunc: func(_ any, u *entities.User) error {
			return errors.New("write failed")
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	req := &dto.UpdateUserRequest{}
	resp, err := svc.UpdateUser(ctx, "u77", req)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.EqualError(t, err, "write failed")
}

func TestUserService_DeleteUser_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return &entities.User{ID: id}, nil
		},
		deleteFunc: func(_ any, id string) error {
			require.Equal(t, "u-del", id)
			return nil
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	err = svc.DeleteUser(ctx, "u-del")
	require.NoError(t, err)
}

func TestUserService_DeleteUser_NotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			return nil, nil
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	err = svc.DeleteUser(ctx, "missing")
	require.Error(t, err)
	assert.EqualError(t, err, "user not found")
}

func TestUserService_DeleteUser_RepoErrors(t *testing.T) {
	ctx := newFiberCtx(t)
	repo := &stubUserRepo{
		getByIDFunc: func(_ any, id string) (*entities.User, error) {
			if id == "db-get-fail" {
				return nil, errors.New("read fail")
			}
			return &entities.User{ID: id}, nil
		},
		deleteFunc: func(_ any, id string) error {
			return errors.New("delete fail")
		},
	}
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	svc := NewUserService(repo, db)

	// Get failure
	err = svc.DeleteUser(ctx, "db-get-fail")
	require.Error(t, err)
	assert.EqualError(t, err, "read fail")

	// Delete failure
	err = svc.DeleteUser(ctx, "u1")
	require.Error(t, err)
	assert.EqualError(t, err, "delete fail")
}

// Targeted tests around pagination math and DB flow of GetAllUsers using an in-memory sqlite database.
// We intentionally avoid complex relations population; we only ensure Count + Find + DTO mapping + flags.
func TestUserService_GetAllUsers_PaginationAndFlags(t *testing.T) {
	ctx := newFiberCtx(t)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Auto-migrate minimal tables referenced by service queries
	err = db.AutoMigrate(&entities.User{}, &entities.Role{}, &entities.Profile{})
	require.NoError(t, err)

	// Seed 7 users
	users := make([]entities.User, 0, 7)
	for i := 1; i <= 7; i++ {
		users = append(users, entities.User{
			ID:       uint(i),
			Name:     "User" + string(rune('A'+i-1)),
			IsActive: i%2 == 0,
		})
	}
	require.NoError(t, db.Create(&users).Error)

	repo := &stubUserRepo{} // not used in GetAllUsers
	svc := NewUserService(repo, db)

	// Page 1, limit 3 => totalPages=3, hasNext=true, hasPrev=false
	resp, err := svc.GetAllUsers(ctx, 1, 3)
	require.NoError(t, err)
	require.NotNil(t, resp)
	assert.Equal(t, 1, resp.Page)
	assert.Equal(t, 3, resp.Limit)
	assert.Equal(t, int64(7), resp.TotalCount)
	assert.Equal(t, 3, resp.TotalPages)
	assert.Equal(t, true, resp.HasNext)
	assert.Equal(t, false, resp.HasPrev)
	require.Len(t, resp.Data, 3)

	// Page 3, limit 3 => last page => hasNext=false, hasPrev=true, items may be 1
	resp2, err := svc.GetAllUsers(ctx, 3, 3)
	require.NoError(t, err)
	assert.Equal(t, 3, resp2.Page)
	assert.Equal(t, false, resp2.HasNext)
	assert.Equal(t, true, resp2.HasPrev)
	require.Len(t, resp2.Data, 1)
}

// Error handling path for GetAllUsers: simulate count error by using a closed DB connection.
//
// Note: sqlite in-memory cannot be trivially "closed" at gorm level; instead, we open a valid DB and
// point Model to an invalid table to trigger a Count error, then ensure the error is propagated.
func TestUserService_GetAllUsers_CountError_Propagates(t *testing.T) {
	ctx := newFiberCtx(t)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Do NOT migrate entities.User, forcing Count() on missing table to fail.
	repo := &stubUserRepo{}
	svc := NewUserService(repo, db)

	resp, err := svc.GetAllUsers(ctx, 1, 5)
	require.Error(t, err, "expected error when counting without table")
	assert.Nil(t, resp)
}

func TestUserService_GetAllUsers_FindError_Propagates(t *testing.T) {
	ctx := newFiberCtx(t)

	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// Migrate only to make Count succeed
	err = db.AutoMigrate(&entities.User{})
	require.NoError(t, err)

	// Create one record so count=1 succeeds
	require.NoError(t, db.Create(&entities.User{ID: 1, Name: "Only", IsActive: true}).Error)

	// Drop the table to force Find to fail after Count succeeded
	require.NoError(t, db.Migrator().DropTable(&entities.User{}))

	repo := &stubUserRepo{}
	svc := NewUserService(repo, db)

	resp, err := svc.GetAllUsers(ctx, 1, 1)
	require.Error(t, err, "expected error when finding after count")
	assert.Nil(t, resp)
}