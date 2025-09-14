package repositories

// Testing library/framework note:
// - Using Go's standard testing package ("testing") with table/sub-tests.
// - No new test dependencies introduced; we use GORM's sqlite driver for ephemeral DB state.

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/entities"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err \!= nil {
		t.Fatalf("failed to open sqlite in-memory db: %v", err)
	}

	// Enable FKs (noop if not supported)
	if res := db.Exec("PRAGMA foreign_keys = ON"); res.Error \!= nil {
		t.Fatalf("failed to enable foreign keys: %v", res.Error)
	}

	// Auto-migrate minimal set required by repository logic and preloads
	if err := db.AutoMigrate(&entities.Permission{}, &entities.Role{}, &entities.Profile{}, &entities.User{}); err \!= nil {
		t.Fatalf("failed to automigrate schema: %v", err)
	}

	// Close underlying connection when test ends
	sqlDB, err := db.DB()
	if err == nil && sqlDB \!= nil {
		t.Cleanup(func() { _ = sqlDB.Close() })
	}
	return db
}

func newRepo(t *testing.T, db *gorm.DB) *userRepository {
	t.Helper()
	r := NewUserRepository(db)
	ur, ok := r.(*userRepository)
	if \!ok || ur == nil {
		t.Fatalf("NewUserRepository returned unexpected type: %T", r)
	}
	return ur
}

func setIDIfZero(obj any) string {
	v := reflect.ValueOf(obj)
	if v.Kind() \!= reflect.Ptr || v.Elem().Kind() \!= reflect.Struct {
		return ""
	}
	f := v.Elem().FieldByName("ID")
	if \!f.IsValid() || \!f.CanSet() {
		return ""
	}

	idStr := ""
	switch f.Kind() {
	case reflect.String:
		if f.String() == "" {
			idStr = uuid.NewString()
			f.SetString(idStr)
		} else {
			idStr = f.String()
		}
	case reflect.Struct:
		// Handle github.com/google/uuid.UUID
		if f.Type().PkgPath() == "github.com/google/uuid" && f.Type().Name() == "UUID" {
			zero := reflect.Zero(f.Type()).Interface()
			if reflect.DeepEqual(f.Interface(), zero) {
				id := uuid.New()
				f.Set(reflect.ValueOf(id))
				idStr = id.String()
			} else {
				idStr = fmt.Sprint(f.Interface())
			}
		}
	case reflect.Uint, reflect.Uint64, reflect.Uint32, reflect.Int, reflect.Int64:
		// Leave zero; DB will autoincrement. Return string rendering.
		idStr = fmt.Sprint(f.Interface())
	default:
		idStr = fmt.Sprint(f.Interface())
	}
	return idStr
}

func getIDString(obj any) string {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() \!= reflect.Struct {
		return ""
	}
	f := v.FieldByName("ID")
	if \!f.IsValid() {
		return ""
	}
	return fmt.Sprint(f.Interface())
}

func setStringField(obj any, field, value string) {
	v := reflect.ValueOf(obj)
	if v.Kind() \!= reflect.Ptr || v.Elem().Kind() \!= reflect.Struct {
		return
	}
	f := v.Elem().FieldByName(field)
	if f.IsValid() && f.CanSet() && f.Kind() == reflect.String {
		f.SetString(value)
	}
}

func seedRole(t *testing.T, db *gorm.DB, name string) entities.Role {
	t.Helper()
	role := entities.Role{}
	setIDIfZero(&role)
	setStringField(&role, "Name", name)
	if err := db.Create(&role).Error; err \!= nil {
		t.Fatalf("seedRole create %q failed: %v", name, err)
	}
	return role
}

func hasRole(u *entities.User, roleName string) bool {
	for _, r := range u.Roles {
		if strings.EqualFold(r.Name, roleName) {
			return true
		}
	}
	return false
}

// -------- Tests --------

func TestNewUserRepository_NotNil(t *testing.T) {
	db := openTestDB(t)
	if got := NewUserRepository(db); got == nil {
		t.Fatalf("NewUserRepository returned nil")
	}
}

func TestCreate_AssignsDefaultRole_WhenNoRolesProvided(t *testing.T) {
	db := openTestDB(t)
	_ = seedRole(t, db, "customer") // default role expected by repository

	repo := newRepo(t, db)

	u := &entities.User{}
	setStringField(u, "Email", "cust1@example.com")
	setStringField(u, "Password", "secret")      // best-effort if present
	setStringField(u, "PasswordHash", "hashed\!") // best-effort if present
	setIDIfZero(u)                                // ensure ID for UUID schemas

	ctx := context.Background()
	if err := repo.Create(ctx, u); err \!= nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.GetByEmail(ctx, "cust1@example.com")
	if err \!= nil {
		t.Fatalf("GetByEmail returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("GetByEmail returned nil user")
	}
	if \!hasRole(got, "customer") {
		t.Fatalf("expected default role 'customer' to be assigned")
	}
}

func TestCreate_ReturnsError_WhenDefaultRoleMissing(t *testing.T) {
	db := openTestDB(t)
	repo := newRepo(t, db)

	u := &entities.User{}
	setStringField(u, "Email", "nobody@example.com")
	setIDIfZero(u)

	ctx := context.Background()
	err := repo.Create(ctx, u)
	if err == nil {
		t.Fatalf("expected error when default 'customer' role is missing")
	}
	if \!strings.Contains(strings.ToLower(err.Error()), "failed to get default role") {
		t.Fatalf("unexpected error: %v", err)
	}

	var count int64
	if derr := db.Model(&entities.User{}).Where("email = ?", "nobody@example.com").Count(&count).Error; derr \!= nil {
		t.Fatalf("count users failed: %v", derr)
	}
	if count \!= 0 {
		t.Fatalf("transaction should rollback on error; found %d users", count)
	}
}

func TestCreate_WithExplicitRoles_DoesNotAssignDefault(t *testing.T) {
	db := openTestDB(t)
	_ = seedRole(t, db, "customer") // present but should not be used
	admin := seedRole(t, db, "admin")

	repo := newRepo(t, db)

	u := &entities.User{
		Roles: []entities.Role{admin},
	}
	setStringField(u, "Email", "admin1@example.com")
	setIDIfZero(u)

	ctx := context.Background()
	if err := repo.Create(ctx, u); err \!= nil {
		t.Fatalf("Create returned error: %v", err)
	}

	got, err := repo.GetByEmail(ctx, "admin1@example.com")
	if err \!= nil {
		t.Fatalf("GetByEmail returned error: %v", err)
	}
	if got == nil {
		t.Fatalf("GetByEmail returned nil user")
	}
	if \!hasRole(got, "admin") {
		t.Fatalf("expected explicit role 'admin' to persist")
	}
	if hasRole(got, "customer") {
		t.Fatalf("did not expect default role 'customer' when explicit roles are provided")
	}
}

func TestGetByEmail_NotFound_ReturnsNil(t *testing.T) {
	db := openTestDB(t)
	repo := newRepo(t, db)

	ctx := context.Background()
	got, err := repo.GetByEmail(ctx, "missing@example.com")
	if err \!= nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got \!= nil {
		t.Fatalf("expected nil user for not found; got: %+v", got)
	}
}

func TestGetByEmail_ContextCanceled_PropagatesError(t *testing.T) {
	db := openTestDB(t)
	repo := newRepo(t, db)

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel before call

	got, err := repo.GetByEmail(ctx, "anything@example.com")
	if err == nil {
		t.Fatalf("expected error due to canceled context")
	}
	if got \!= nil {
		t.Fatalf("expected nil user on error")
	}
}

func TestGetByID_HappyPath(t *testing.T) {
	db := openTestDB(t)
	_ = seedRole(t, db, "customer")
	repo := newRepo(t, db)

	u := &entities.User{}
	setStringField(u, "Email", "byid@example.com")
	setIDIfZero(u)

	if err := repo.Create(context.Background(), u); err \!= nil {
		t.Fatalf("Create error: %v", err)
	}
	created, err := repo.GetByEmail(context.Background(), "byid@example.com")
	if err \!= nil || created == nil {
		t.Fatalf("failed to fetch by email: %v, user=%v", err, created)
	}

	idStr := getIDString(created)
	got, err := repo.GetByID(context.Background(), idStr)
	if err \!= nil {
		t.Fatalf("GetByID error: %v", err)
	}
	if got == nil || strings.ToLower(got.Email) \!= "byid@example.com" {
		t.Fatalf("unexpected GetByID result: %+v", got)
	}
}

func TestUpdate_PersistsChanges(t *testing.T) {
	db := openTestDB(t)
	_ = seedRole(t, db, "customer")
	repo := newRepo(t, db)

	u := &entities.User{}
	setStringField(u, "Email", "old@example.com")
	setIDIfZero(u)

	if err := repo.Create(context.Background(), u); err \!= nil {
		t.Fatalf("Create error: %v", err)
	}

	// Fetch, mutate, update
	stored, err := repo.GetByEmail(context.Background(), "old@example.com")
	if err \!= nil || stored == nil {
		t.Fatalf("failed to fetch stored user: %v, u=%v", err, stored)
	}
	setStringField(stored, "Email", "new@example.com")

	if err := repo.Update(context.Background(), stored); err \!= nil {
		t.Fatalf("Update error: %v", err)
	}

	got, err := repo.GetByEmail(context.Background(), "new@example.com")
	if err \!= nil {
		t.Fatalf("GetByEmail error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected user after update by new email")
	}
}

func TestDelete_RemovesUser(t *testing.T) {
	db := openTestDB(t)
	_ = seedRole(t, db, "customer")
	repo := newRepo(t, db)

	u := &entities.User{}
	setStringField(u, "Email", "del@example.com")
	setIDIfZero(u)

	if err := repo.Create(context.Background(), u); err \!= nil {
		t.Fatalf("Create error: %v", err)
	}
	created, err := repo.GetByEmail(context.Background(), "del@example.com")
	if err \!= nil || created == nil {
		t.Fatalf("failed to fetch user pre-delete: %v, u=%v", err, created)
	}

	idStr := getIDString(created)
	if err := repo.Delete(context.Background(), idStr); err \!= nil {
		t.Fatalf("Delete error: %v", err)
	}

	got, err := repo.GetByID(context.Background(), idStr)
	if err \!= nil {
		t.Fatalf("GetByID after delete error: %v", err)
	}
	if got \!= nil {
		t.Fatalf("expected nil after delete, got: %+v", got)
	}
}

func TestExistsByEmail(t *testing.T) {
	db := openTestDB(t)
	_ = seedRole(t, db, "customer")
	repo := newRepo(t, db)

	email := "exists@example.com"
	u := &entities.User{}
	setStringField(u, "Email", email)
	setIDIfZero(u)
	if err := repo.Create(context.Background(), u); err \!= nil {
		t.Fatalf("Create error: %v", err)
	}

	ctx := context.Background()
	ok, err := repo.ExistsByEmail(ctx, email)
	if err \!= nil {
		t.Fatalf("ExistsByEmail error: %v", err)
	}
	if \!ok {
		t.Fatalf("expected ExistsByEmail to be true for existing email")
	}

	ok, err = repo.ExistsByEmail(ctx, "missing@example.com")
	if err \!= nil {
		t.Fatalf("ExistsByEmail error (missing): %v", err)
	}
	if ok {
		t.Fatalf("expected ExistsByEmail to be false for missing email")
	}

	// Canceled context propagates error
	cctx, cancel := context.WithTimeout(context.Background(), time.Millisecond)
	cancel()
	ok, err = repo.ExistsByEmail(cctx, email)
	if err == nil {
		t.Fatalf("expected error with canceled context")
	}
	if ok {
		t.Fatalf("expected ok=false on error")
	}
}