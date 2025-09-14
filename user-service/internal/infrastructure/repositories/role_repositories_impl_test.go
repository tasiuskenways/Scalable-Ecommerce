package repositories_test

import (
	"context"
	"testing"
	"time"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	domainrepos "github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/repositories"
	repoimpl "github.com/tasiuskenways/scalable-ecommerce/user-service/internal/infrastructure/repositories"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type testDeps struct {
	DB   *gorm.DB
	Repo domainrepos.RoleRepository
	Ctx  context.Context
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:role_repo_test?mode=memory&cache=shared"), &gorm.Config{
		SkipDefaultTransaction: true,
	})
	if err \!= nil {
		t.Fatalf("failed to open sqlite in-memory DB: %v", err)
	}
	// Try to discover whether Role and Permission exist with fields.
	// Since we don't know exact shapes, rely on entities.Role and entities.Permission
	// and AutoMigrate both. If Permission not present, migration will still work if it's defined.
	if err := db.AutoMigrate(&entities.Role{}, &entities.Permission{}); err \!= nil {
		t.Fatalf("failed to automigrate: %v", err)
	}
	return db
}

func newDeps(t *testing.T) testDeps {
	db := newTestDB(t)
	return testDeps{
		DB:   db,
		Repo: repoimpl.NewRoleRepository(db),
		Ctx:  context.Background(),
	}
}

// helper to seed permissions and return slice
func seedPermissions(t *testing.T, db *gorm.DB, names ...string) []entities.Permission {
	t.Helper()
	perms := make([]entities.Permission, 0, len(names))
	for _, n := range names {
		p := entities.Permission{
			Name:        n,
			Description: n + " desc",
		}
		if err := db.Create(&p).Error; err \!= nil {
			t.Fatalf("failed creating permission %s: %v", n, err)
		}
		perms = append(perms, p)
	}
	return perms
}

func mustCreateRole(t *testing.T, db *gorm.DB, name string, perms []entities.Permission) entities.Role {
	t.Helper()
	r := entities.Role{
		Name:        name,
		Description: name + " role",
		Permissions: perms,
	}
	if err := db.Create(&r).Error; err \!= nil {
		t.Fatalf("failed to create role %s: %v", name, err)
	}
	return r
}

func TestCreate_Success(t *testing.T) {
	deps := newDeps(t)
	perms := seedPermissions(t, deps.DB, "read", "write")
	role := &entities.Role{
		Name:        "admin",
		Description: "administrator",
		Permissions: perms,
	}

	if err := deps.Repo.Create(deps.Ctx, role); err \!= nil {
		t.Fatalf("Create returned error: %v", err)
	}

	// Verify persisted with permissions preloaded by fetching
	got, err := deps.Repo.GetByID(deps.Ctx, role.ID)
	if err \!= nil {
		t.Fatalf("GetByID after Create error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected role, got nil")
	}
	if got.Name \!= role.Name {
		t.Errorf("name mismatch: want %q got %q", role.Name, got.Name)
	}
	if len(got.Permissions) \!= len(perms) {
		t.Errorf("permissions length mismatch: want %d got %d", len(perms), len(got.Permissions))
	}
}

func TestGetByID_NotFoundReturnsNil(t *testing.T) {
	deps := newDeps(t)
	got, err := deps.Repo.GetByID(deps.Ctx, "non-existent-id")
	if err \!= nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if got \!= nil {
		t.Fatalf("expected nil role on not found, got %+v", got)
	}
}

func TestGetByName_SuccessAndNotFound(t *testing.T) {
	deps := newDeps(t)
	perms := seedPermissions(t, deps.DB, "read")
	created := mustCreateRole(t, deps.DB, "moderator", perms)

	// success
	got, err := deps.Repo.GetByName(deps.Ctx, "moderator")
	if err \!= nil {
		t.Fatalf("GetByName error: %v", err)
	}
	if got == nil || got.ID \!= created.ID {
		t.Fatalf("expected role id %s, got %+v", created.ID, got)
	}
	if len(got.Permissions) \!= 1 || got.Permissions[0].Name \!= "read" {
		t.Errorf("expected 1 permission 'read', got %+v", got.Permissions)
	}

	// not found -> nil, nil
	none, err := deps.Repo.GetByName(deps.Ctx, "ghost")
	if err \!= nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if none \!= nil {
		t.Fatalf("expected nil on not found, got %+v", none)
	}
}

func TestGetAll_EmptyAndPopulated(t *testing.T) {
	deps := newDeps(t)

	// empty
	roles, err := deps.Repo.GetAll(deps.Ctx)
	if err \!= nil {
		t.Fatalf("GetAll error on empty: %v", err)
	}
	if len(roles) \!= 0 {
		t.Fatalf("expected 0 roles, got %d", len(roles))
	}

	// populated
	perms := seedPermissions(t, deps.DB, "read", "write", "delete")
	mustCreateRole(t, deps.DB, "r1", perms[:1])
	mustCreateRole(t, deps.DB, "r2", perms[:2])
	mustCreateRole(t, deps.DB, "r3", perms[:3])

	roles, err = deps.Repo.GetAll(deps.Ctx)
	if err \!= nil {
		t.Fatalf("GetAll error: %v", err)
	}
	if len(roles) \!= 3 {
		t.Fatalf("expected 3 roles, got %d", len(roles))
	}
	// ensure permissions preloaded (not zero for at least one)
	foundPreload := false
	for _, r := range roles {
		if len(r.Permissions) > 0 {
			foundPreload = true
			break
		}
	}
	if \!foundPreload {
		t.Errorf("expected at least one role with preloaded permissions")
	}
}

func TestUpdate_SimpleFieldsAndPermissions(t *testing.T) {
	deps := newDeps(t)
	perms := seedPermissions(t, deps.DB, "read", "write")
	r := mustCreateRole(t, deps.DB, "member", perms[:1])

	// update description and swap permissions
	newPerms := seedPermissions(t, deps.DB, "export")
	r.Description = "updated description"
	r.Permissions = append([]entities.Permission{}, newPerms...)
	if err := deps.Repo.Update(deps.Ctx, &r); err \!= nil {
		t.Fatalf("Update error: %v", err)
	}

	got, err := deps.Repo.GetByID(deps.Ctx, r.ID)
	if err \!= nil {
		t.Fatalf("GetByID after Update error: %v", err)
	}
	if got.Description \!= "updated description" {
		t.Errorf("desc mismatch: want %q got %q", "updated description", got.Description)
	}
	if len(got.Permissions) \!= 1 || got.Permissions[0].Name \!= "export" {
		t.Errorf("permissions not updated: %+v", got.Permissions)
	}
}

func TestDelete_ByID(t *testing.T) {
	deps := newDeps(t)
	perms := seedPermissions(t, deps.DB, "read")
	r := mustCreateRole(t, deps.DB, "temp", perms)

	if err := deps.Repo.Delete(deps.Ctx, r.ID); err \!= nil {
		t.Fatalf("Delete error: %v", err)
	}

	got, err := deps.Repo.GetByID(deps.Ctx, r.ID)
	if err \!= nil {
		t.Fatalf("GetByID after delete error: %v", err)
	}
	if got \!= nil {
		t.Fatalf("expected nil after delete, got %+v", got)
	}
}

func TestExistsByName(t *testing.T) {
	deps := newDeps(t)
	perms := seedPermissions(t, deps.DB, "read")
	mustCreateRole(t, deps.DB, "exists", perms)

	ok, err := deps.Repo.ExistsByName(deps.Ctx, "exists")
	if err \!= nil {
		t.Fatalf("ExistsByName error: %v", err)
	}
	if \!ok {
		t.Fatalf("expected exists=true for name 'exists'")
	}

	no, err := deps.Repo.ExistsByName(deps.Ctx, "missing")
	if err \!= nil {
		t.Fatalf("ExistsByName error for missing: %v", err)
	}
	if no {
		t.Fatalf("expected exists=false for missing name")
	}
}

func TestGetByIDs(t *testing.T) {
	deps := newDeps(t)
	perms := seedPermissions(t, deps.DB, "read", "write")
	r1 := mustCreateRole(t, deps.DB, "A", perms[:1])
	r2 := mustCreateRole(t, deps.DB, "B", perms)

	got, err := deps.Repo.GetByIDs(deps.Ctx, []string{r1.ID, r2.ID})
	if err \!= nil {
		t.Fatalf("GetByIDs error: %v", err)
	}
	if len(got) \!= 2 {
		t.Fatalf("expected 2 roles, got %d", len(got))
	}
	// ensure permissions preloaded
	for _, r := range got {
		if r.ID == r1.ID && len(r.Permissions) \!= 1 {
			t.Errorf("r1 permissions expected 1, got %d", len(r.Permissions))
		}
		if r.ID == r2.ID && len(r.Permissions) \!= 2 {
			t.Errorf("r2 permissions expected 2, got %d", len(r.Permissions))
		}
	}
}

func TestContextPropagation(t *testing.T) {
	deps := newDeps(t)
	// Use a context with deadline to exercise WithContext path
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()
	perms := seedPermissions(t, deps.DB, "read")
	role := &entities.Role{Name: "ctx-check", Permissions: perms}
	if err := deps.Repo.Create(ctx, role); err \!= nil {
		t.Fatalf("Create with context error: %v", err)
	}
	if _, err := deps.Repo.GetByName(ctx, "ctx-check"); err \!= nil {
		t.Fatalf("GetByName with context error: %v", err)
	}
}