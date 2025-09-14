package services

import (
	"context"
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

//
// Test strategy and framework:
// - Framework: Go testing with stretchr/testify (assert/require) for readability.
// - We use lightweight in-memory sqlite via gorm to satisfy s.db calls (Transaction, Preload, First).
// - Repositories are mocked with local fake implementations to cover happy paths, edge cases, and failures.
// - We focus on the service public methods and their behaviors per the diff.
//

// ---------------------------
// Fake repositories for tests
// ---------------------------

type fakeRoleRepo struct {
	// behavior configuration
	existsByName    map[string]bool
	existsErr       error
	createErr       error
	updateErr       error
	deleteErr       error
	getByIDMap      map[string]*entities.Role
	getByIDErr      error
	getAllList      []entities.Role
	getAllErr       error
	getByNameMap    map[string]*entities.Role
	getByNameErr    error
	getByIDsMap     map[string]entities.Role // keyed by ID
	getByIDsErr     error
}

func (f *fakeRoleRepo) ExistsByName(ctx context.Context, name string) (bool, error) {
	if f.existsErr \!= nil {
		return false, f.existsErr
	}
	return f.existsByName[name], nil
}

func (f *fakeRoleRepo) Create(ctx context.Context, role *entities.Role) error {
	return f.createErr
}

func (f *fakeRoleRepo) Update(ctx context.Context, role *entities.Role) error {
	return f.updateErr
}

func (f *fakeRoleRepo) Delete(ctx context.Context, id string) error {
	return f.deleteErr
}

func (f *fakeRoleRepo) GetByID(ctx context.Context, id string) (*entities.Role, error) {
	if f.getByIDErr \!= nil {
		return nil, f.getByIDErr
	}
	if r, ok := f.getByIDMap[id]; ok {
		return r, nil
	}
	return nil, nil
}

func (f *fakeRoleRepo) GetAll(ctx context.Context) ([]entities.Role, error) {
	if f.getAllErr \!= nil {
		return nil, f.getAllErr
	}
	return f.getAllList, nil
}

func (f *fakeRoleRepo) GetByName(ctx context.Context, name string) (*entities.Role, error) {
	if f.getByNameErr \!= nil {
		return nil, f.getByNameErr
	}
	if r, ok := f.getByNameMap[name]; ok {
		return r, nil
	}
	return nil, nil
}

func (f *fakeRoleRepo) GetByIDs(ctx context.Context, ids []string) ([]entities.Role, error) {
	if f.getByIDsErr \!= nil {
		return nil, f.getByIDsErr
	}
	var out []entities.Role
	for _, id := range ids {
		if r, ok := f.getByIDsMap[id]; ok {
			out = append(out, r)
		}
	}
	return out, nil
}

type fakePermissionRepo struct {
	getAllList []entities.Permission
	getAllErr  error

	getByIDsMap map[string]entities.Permission
	getByIDsErr error
}

func (f *fakePermissionRepo) GetAll(ctx context.Context) ([]entities.Permission, error) {
	if f.getAllErr \!= nil {
		return nil, f.getAllErr
	}
	return f.getAllList, nil
}

func (f *fakePermissionRepo) GetByIDs(ctx context.Context, ids []string) ([]entities.Permission, error) {
	if f.getByIDsErr \!= nil {
		return nil, f.getByIDsErr
	}
	var out []entities.Permission
	for _, id := range ids {
		if p, ok := f.getByIDsMap[id]; ok {
			out = append(out, p)
		}
	}
	return out, nil
}

type fakeUserRepo struct {
	getByIDMap map[string]*entities.User
	getByIDErr error
}

func (f *fakeUserRepo) GetByID(ctx context.Context, id string) (*entities.User, error) {
	if f.getByIDErr \!= nil {
		return nil, f.getByIDErr
	}
	if u, ok := f.getByIDMap[id]; ok {
		return u, nil
	}
	return nil, nil
}

// Ensure fake types satisfy interfaces
var (
	_ repositories.RoleRepository       = (*fakeRoleRepo)(nil)
	_ repositories.PermissionRepository = (*fakePermissionRepo)(nil)
	_ repositories.UserRepository       = (*fakeUserRepo)(nil)
)

// ---------------------------
// Test helpers
// ---------------------------

func newFiberCtx(t *testing.T) *fiber.Ctx {
	t.Helper()
	app := fiber.New()
	ctx := app.AcquireCtx(&fiber.Ctx{})
	return ctx
}

func newGormInMemory(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)
	// Auto-migrate minimal schema for User, Role, Permission and join table to support Associations and Preload
	type Permission struct {
		ID   string `gorm:"primaryKey"`
		Name string
	}
	type Role struct {
		ID          string `gorm:"primaryKey"`
		Name        string
		Description string
		Permissions []Permission `gorm:"many2many:role_permissions;"`
	}
	type User struct {
		ID    string `gorm:"primaryKey"`
		Email string
		Roles []Role `gorm:"many2many:user_roles;"`
	}
	require.NoError(t, db.AutoMigrate(&Permission{}, &Role{}, &User{}))
	return db
}

// seedUserWithRoles seeds GORM DB with provided user, roles, and permissions.
func seedUserWithRoles(t *testing.T, db *gorm.DB, user entities.User, roles []entities.Role) {
	t.Helper()
	// Map entities.* to local migratable structs
	type Permission struct {
		ID   string `gorm:"primaryKey"`
		Name string
	}
	type Role struct {
		ID          string `gorm:"primaryKey"`
		Name        string
		Description string
		Permissions []Permission `gorm:"many2many:role_permissions;"`
	}
	type User struct {
		ID    string `gorm:"primaryKey"`
		Email string
		Roles []Role `gorm:"many2many:user_roles;"`
	}

	var convRoles []Role
	for _, r := range roles {
		var convPerms []Permission
		for _, p := range r.Permissions {
			convPerms = append(convPerms, Permission{ID: p.ID, Name: p.Name})
		}
		convRoles = append(convRoles, Role{
			ID:          r.ID,
			Name:        r.Name,
			Description: r.Description,
			Permissions: convPerms,
		})
	}
	u := User{ID: user.ID, Email: user.Email, Roles: convRoles}
	require.NoError(t, db.Create(&u).Error)
}

// ---------------------------
// Tests for CreateRole
// ---------------------------

func TestRoleService_CreateRole_Success(t *testing.T) {
	ctx := newFiberCtx(t)

	roleRepo := &fakeRoleRepo{
		existsByName: map[string]bool{"admin": false},
	}
	permRepo := &fakePermissionRepo{
		getByIDsMap: map[string]entities.Permission{
			"p1": {ID: "p1", Name: "read"},
			"p2": {ID: "p2", Name: "write"},
		},
	}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	req := &dto.CreateRoleRequest{
		Name:        "admin",
		Description: "Administrator",
		Permissions: []string{"p1", "p2"},
	}

	res, err := svc.CreateRole(ctx, req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "admin", res.Name)
	assert.Equal(t, "Administrator", res.Description)
	assert.Len(t, res.Permissions, 2)
}

func TestRoleService_CreateRole_AlreadyExists(t *testing.T) {
	ctx := newFiberCtx(t)

	roleRepo := &fakeRoleRepo{
		existsByName: map[string]bool{"admin": true},
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	_, err := svc.CreateRole(ctx, &dto.CreateRoleRequest{Name: "admin"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRoleService_CreateRole_PermissionLookupError(t *testing.T) {
	ctx := newFiberCtx(t)

	roleRepo := &fakeRoleRepo{
		existsByName: map[string]bool{"staff": false},
	}
	permRepo := &fakePermissionRepo{
		getByIDsErr: errors.New("perm-err"),
	}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	_, err := svc.CreateRole(ctx, &dto.CreateRoleRequest{Name: "staff", Permissions: []string{"pX"}})
	require.Error(t, err)
	assert.Equal(t, "perm-err", err.Error())
}

func TestRoleService_CreateRole_RepoCreateError(t *testing.T) {
	ctx := newFiberCtx(t)

	roleRepo := &fakeRoleRepo{
		existsByName: map[string]bool{"staff": false},
		createErr:    errors.New("create-err"),
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	_, err := svc.CreateRole(ctx, &dto.CreateRoleRequest{Name: "staff"})
	require.Error(t, err)
	assert.Equal(t, "create-err", err.Error())
}

// ---------------------------
// Tests for GetRole
// ---------------------------

func TestRoleService_GetRole_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	role := &entities.Role{ID: "r1", Name: "admin"}

	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": role},
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	res, err := svc.GetRole(ctx, "r1")
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "admin", res.Name)
}

func TestRoleService_GetRole_NotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	_, err := svc.GetRole(ctx, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRoleService_GetRole_RepoError(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{getByIDErr: errors.New("boom")}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	_, err := svc.GetRole(ctx, "r1")
	require.Error(t, err)
	assert.Equal(t, "boom", err.Error())
}

// ---------------------------
// Tests for GetAllRoles
// ---------------------------

func TestRoleService_GetAllRoles_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{
		getAllList: []entities.Role{
			{ID: "r1", Name: "admin"},
			{ID: "r2", Name: "staff"},
		},
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	res, err := svc.GetAllRoles(ctx)
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "admin", res[0].Name)
	assert.Equal(t, "staff", res[1].Name)
}

func TestRoleService_GetAllRoles_Error(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{getAllErr: errors.New("list-err")}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	_, err := svc.GetAllRoles(ctx)
	require.Error(t, err)
	assert.Equal(t, "list-err", err.Error())
}

// ---------------------------
// Tests for UpdateRole
// ---------------------------

func TestRoleService_UpdateRole_Success_AllFields(t *testing.T) {
	ctx := newFiberCtx(t)
	existing := &entities.Role{ID: "r1", Name: "old", Description: "old desc"}

	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": existing},
		getByNameMap: map[string]*entities.Role{
			// Will return nil for "new" to indicate no conflict
		},
	}
	permRepo := &fakePermissionRepo{
		getByIDsMap: map[string]entities.Permission{
			"p1": {ID: "p1", Name: "read"},
		},
	}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	newName := "new"
	newDesc := "new desc"
	newPerms := []string{"p1"}
	req := &dto.UpdateRoleRequest{
		Name:        &newName,
		Description: &newDesc,
		Permissions: newPerms,
	}
	res, err := svc.UpdateRole(ctx, "r1", req)
	require.NoError(t, err)
	require.NotNil(t, res)
	assert.Equal(t, "new", res.Name)
	assert.Equal(t, "new desc", res.Description)
	assert.Len(t, res.Permissions, 1)
}

func TestRoleService_UpdateRole_NameConflict(t *testing.T) {
	ctx := newFiberCtx(t)
	existing := &entities.Role{ID: "r1", Name: "old"}

	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": existing},
		getByNameMap: map[string]*entities.Role{
			"taken": {ID: "r2", Name: "taken"},
		},
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	newName := "taken"
	req := &dto.UpdateRoleRequest{Name: &newName}
	_, err := svc.UpdateRole(ctx, "r1", req)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "already exists")
}

func TestRoleService_UpdateRole_NotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{} // no such ID
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	_, err := svc.UpdateRole(ctx, "missing", &dto.UpdateRoleRequest{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRoleService_UpdateRole_PermissionsLookupError(t *testing.T) {
	ctx := newFiberCtx(t)
	existing := &entities.Role{ID: "r1", Name: "old"}
	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": existing},
	}
	permRepo := &fakePermissionRepo{getByIDsErr: errors.New("perm-err")}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	req := &dto.UpdateRoleRequest{Permissions: []string{"p1"}}
	_, err := svc.UpdateRole(ctx, "r1", req)
	require.Error(t, err)
	assert.Equal(t, "perm-err", err.Error())
}

func TestRoleService_UpdateRole_UpdateError(t *testing.T) {
	ctx := newFiberCtx(t)
	existing := &entities.Role{ID: "r1", Name: "old"}
	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": existing},
		updateErr:  errors.New("update-err"),
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	_, err := svc.UpdateRole(ctx, "r1", &dto.UpdateRoleRequest{})
	require.Error(t, err)
	assert.Equal(t, "update-err", err.Error())
}

// ---------------------------
// Tests for DeleteRole
// ---------------------------

func TestRoleService_DeleteRole_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": {ID: "r1"}},
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	err := svc.DeleteRole(ctx, "r1")
	require.NoError(t, err)
}

func TestRoleService_DeleteRole_NotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{} // none
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	err := svc.DeleteRole(ctx, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestRoleService_DeleteRole_GetError(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{getByIDErr: errors.New("get-err")}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	err := svc.DeleteRole(ctx, "r1")
	require.Error(t, err)
	assert.Equal(t, "get-err", err.Error())
}

func TestRoleService_DeleteRole_DeleteError(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{
		getByIDMap: map[string]*entities.Role{"r1": {ID: "r1"}},
		deleteErr:  errors.New("del-err"),
	}
	permRepo := &fakePermissionRepo{}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	err := svc.DeleteRole(ctx, "r1")
	require.Error(t, err)
	assert.Equal(t, "del-err", err.Error())
}

// ---------------------------
// Tests for AssignRolesToUser
// ---------------------------

func TestRoleService_AssignRolesToUser_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	db := newGormInMemory(t)

	user := &entities.User{ID: "u1", Email: "u1@example.com"}
	userRepo := &fakeUserRepo{getByIDMap: map[string]*entities.User{"u1": user}}

	roleRepo := &fakeRoleRepo{
		getByIDsMap: map[string]entities.Role{
			"r1": {ID: "r1", Name: "admin"},
			"r2": {ID: "r2", Name: "staff"},
		},
	}
	permRepo := &fakePermissionRepo{}

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	req := &dto.AssignRoleRequest{
		UserID:  "u1",
		RoleIDs: []string{"r1", "r2"},
	}
	err := svc.AssignRolesToUser(ctx, req)
	require.NoError(t, err)
}

func TestRoleService_AssignRolesToUser_UserNotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	db := newGormInMemory(t)
	userRepo := &fakeUserRepo{} // nil
	roleRepo := &fakeRoleRepo{}
	permRepo := &fakePermissionRepo{}

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	err := svc.AssignRolesToUser(ctx, &dto.AssignRoleRequest{UserID: "missing"})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

func TestRoleService_AssignRolesToUser_RoleLookupError(t *testing.T) {
	ctx := newFiberCtx(t)
	db := newGormInMemory(t)
	userRepo := &fakeUserRepo{getByIDMap: map[string]*entities.User{"u1": {ID: "u1"}}}
	roleRepo := &fakeRoleRepo{getByIDsErr: errors.New("roles-err")}
	permRepo := &fakePermissionRepo{}

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	err := svc.AssignRolesToUser(ctx, &dto.AssignRoleRequest{UserID: "u1", RoleIDs: []string{"r1"}})
	require.Error(t, err)
	assert.Equal(t, "roles-err", err.Error())
}

// ---------------------------
// Tests for GetUserRoles
// ---------------------------

func TestRoleService_GetUserRoles_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	db := newGormInMemory(t)

	// Seed user with roles and permissions
	user := entities.User{ID: "u1", Email: "u1@example.com"}
	roles := []entities.Role{
		{ID: "r1", Name: "admin", Description: "Administrator", Permissions: []entities.Permission{{ID: "p1", Name: "read"}}},
		{ID: "r2", Name: "staff", Description: "Staff", Permissions: []entities.Permission{{ID: "p2", Name: "write"}}},
	}
	seedUserWithRoles(t, db, user, roles)

	userRepo := &fakeUserRepo{getByIDMap: map[string]*entities.User{"u1": &user}}
	roleRepo := &fakeRoleRepo{}
	permRepo := &fakePermissionRepo{}

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	res, err := svc.GetUserRoles(ctx, "u1")
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.ElementsMatch(t, []string{"admin", "staff"}, []string{res[0].Name, res[1].Name})
}

func TestRoleService_GetUserRoles_UserNotFound(t *testing.T) {
	ctx := newFiberCtx(t)
	db := newGormInMemory(t)
	userRepo := &fakeUserRepo{} // nil
	roleRepo := &fakeRoleRepo{}
	permRepo := &fakePermissionRepo{}

	svc := NewRoleService(roleRepo, permRepo, userRepo, db)
	_, err := svc.GetUserRoles(ctx, "missing")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "user not found")
}

// ---------------------------
// Tests for GetAllPermissions
// ---------------------------

func TestRoleService_GetAllPermissions_Success(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{}
	permRepo := &fakePermissionRepo{
		getAllList: []entities.Permission{
			{ID: "p1", Name: "read"},
			{ID: "p2", Name: "write"},
		},
	}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	res, err := svc.GetAllPermissions(ctx)
	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "read", res[0].Name)
	assert.Equal(t, "write", res[1].Name)
}

func TestRoleService_GetAllPermissions_Error(t *testing.T) {
	ctx := newFiberCtx(t)
	roleRepo := &fakeRoleRepo{}
	permRepo := &fakePermissionRepo{getAllErr: errors.New("perm-err")}
	userRepo := &fakeUserRepo{}
	db := newGormInMemory(t)
	svc := NewRoleService(roleRepo, permRepo, userRepo, db)

	_, err := svc.GetAllPermissions(ctx)
	require.Error(t, err)
	assert.Equal(t, "perm-err", err.Error())
}