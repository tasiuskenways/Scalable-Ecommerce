// NOTE: Tests use Go's standard testing package. If github.com/stretchr/testify is present in the project,
// they also use require/assert for clearer assertions. This follows existing Go testing conventions.
//
// These tests focus on the diff-impacted behaviors of permissionRepository:
// - Create, GetByID, GetByName: correct nil-return on not found and proper error propagation
// - GetAll: returns all or empty slice
// - Update: persists field changes
// - Delete: removes record by id
// - ExistsByName: accurate boolean and error propagation
// - GetByIDs: subset retrieval
// - Context cancellation: ensures errors bubble when ctx is canceled
//
// The tests use an in-memory SQLite GORM database with AutoMigrate on entities.Permission for isolation.

package repositories

import (
	"context"
	"testing"
	"time"

	"gorm.io/gorm"
	"gorm.io/driver/sqlite"

	// "require" and "assert" are optional; compile succeeds without testify.
	// If testify is not in go.mod, simply remove these imports.
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/assert"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	require.NoError(t, err, "open in-memory sqlite")

	// AutoMigrate Permission entity so that GORM knows the schema and not-found becomes ErrRecordNotFound (not 'no such table').
	err = db.AutoMigrate(&entities.Permission{})
	require.NoError(t, err, "automigrate Permission")

	// Enforce foreign keys for sqlite (even if not needed here, it's a common toggle).
	sqlDB, err := db.DB()
	require.NoError(t, err)
	_, _ = sqlDB.Exec("PRAGMA foreign_keys = ON;")

	return db
}

func seedPermission(t *testing.T, db *gorm.DB, p entities.Permission) entities.Permission {
	t.Helper()
	require.NoError(t, db.Create(&p).Error)
	return p
}

func TestPermissionRepository_Create_And_GetByID_HappyPath(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)

	ctx := context.Background()
	p := &entities.Permission{
		ID:   "perm-1",
		Name: "READ_USERS",
	}
	err := repo.Create(ctx, p)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "perm-1")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "perm-1", got.ID)
	assert.Equal(t, "READ_USERS", got.Name)
}

func TestPermissionRepository_GetByID_NotFoundReturnsNil(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)

	ctx := context.Background()
	got, err := repo.GetByID(ctx, "missing-id")
	require.NoError(t, err)
	assert.Nil(t, got, "should return nil, nil when record not found")
}

func TestPermissionRepository_GetByName_HappyPath_And_NotFound(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)

	seedPermission(t, db, entities.Permission{ID: "p2", Name: "WRITE_PRODUCTS"})

	ctx := context.Background()
	got, err := repo.GetByName(ctx, "WRITE_PRODUCTS")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "p2", got.ID)

	none, err := repo.GetByName(ctx, "MISSING")
	require.NoError(t, err)
	assert.Nil(t, none)
}

func TestPermissionRepository_GetAll_EmptyAndMulti(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	// Empty case
	all, err := repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 0)

	// Seed multiple
	seedPermission(t, db, entities.Permission{ID: "p3", Name: "A"})
	seedPermission(t, db, entities.Permission{ID: "p4", Name: "B"})

	all, err = repo.GetAll(ctx)
	require.NoError(t, err)
	assert.Len(t, all, 2)
}

func TestPermissionRepository_Update_PersistsChanges(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)

	seedPermission(t, db, entities.Permission{ID: "p5", Name: "OLD"})

	ctx := context.Background()
	up := &entities.Permission{ID: "p5", Name: "NEW"}
	err := repo.Update(ctx, up)
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "p5")
	require.NoError(t, err)
	require.NotNil(t, got)
	assert.Equal(t, "NEW", got.Name)
}

func TestPermissionRepository_Delete_RemovesRecord(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)

	seedPermission(t, db, entities.Permission{ID: "p6", Name: "DEL"})

	ctx := context.Background()
	err := repo.Delete(ctx, "p6")
	require.NoError(t, err)

	got, err := repo.GetByID(ctx, "p6")
	require.NoError(t, err)
	assert.Nil(t, got)
}

func TestPermissionRepository_ExistsByName(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	seedPermission(t, db, entities.Permission{ID: "p7", Name: "EXISTS"})

	ok, err := repo.ExistsByName(ctx, "EXISTS")
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.ExistsByName(ctx, "NOPE")
	require.NoError(t, err)
	assert.False(t, ok)
}

func TestPermissionRepository_GetByIDs(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)
	ctx := context.Background()

	seedPermission(t, db, entities.Permission{ID: "p8", Name: "N1"})
	seedPermission(t, db, entities.Permission{ID: "p9", Name: "N2"})
	seedPermission(t, db, entities.Permission{ID: "p10", Name: "N3"})

	got, err := repo.GetByIDs(ctx, []string{"p8", "p10", "missing"})
	require.NoError(t, err)

	ids := map[string]struct{}{}
	for _, p := range got {
		ids[p.ID] = struct{}{}
	}
	assert.Contains(t, ids, "p8")
	assert.Contains(t, ids, "p10")
	assert.NotContains(t, ids, "missing")
	assert.Len(t, got, 2)
}

func TestPermissionRepository_ContextCanceled_PropagatesErrors(t *testing.T) {
	db := newTestDB(t)
	repo := NewPermissionRepository(db)

	// Seed one so not-found branch isn't involved; we want the ctx error to surface.
	seedPermission(t, db, entities.Permission{ID: "pc1", Name: "CTX"})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before operations

	// Each method should return a non-nil error when ctx is canceled.
	errs := []error{
		repo.Create(ctx, &entities.Permission{ID: "pc-new", Name: "X"}),
		func() error { _, e := repo.GetByID(ctx, "pc1"); return e }(),
		func() error { _, e := repo.GetByName(ctx, "CTX"); return e }(),
		func() error { _, e := repo.GetAll(ctx); return e }(),
		repo.Update(ctx, &entities.Permission{ID: "pc1", Name: "CTX2"}),
		repo.Delete(ctx, "pc1"),
		func() error { _, e := repo.ExistsByName(ctx, "CTX"); return e }(),
		func() error { _, e := repo.GetByIDs(ctx, []string{"pc1"}); return e }(),
	}

	for i, e := range errs {
		if e == nil {
			t.Fatalf("expected error due to canceled context at index %d, got nil", i)
		}
	}
}

// --- Optional: compile-time check that our repo satisfies the interface ---
// If the interface path differs, adjust as needed. This is a harmless compile check.
/*
import domrepos "github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
var _ domrepos.PermissionRepository = (*permissionRepository)(nil)
*/