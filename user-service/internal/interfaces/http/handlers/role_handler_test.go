package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"

	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/application/dto"
	"github.com/tasiuskenways/scalable-ecommerce/user-service/internal/domain/services"
)

type mockRoleService struct {
	CreateRoleFn         func(c *fiber.Ctx, req *dto.CreateRoleRequest) (interface{}, error)
	GetRoleFn            func(c *fiber.Ctx, id string) (interface{}, error)
	GetAllRolesFn        func(c *fiber.Ctx) (interface{}, error)
	UpdateRoleFn         func(c *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (interface{}, error)
	DeleteRoleFn         func(c *fiber.Ctx, id string) error
	AssignRolesToUserFn  func(c *fiber.Ctx, req *dto.AssignRoleRequest) error
	GetUserRolesFn       func(c *fiber.Ctx, userID string) (interface{}, error)
	GetAllPermissionsFn  func(c *fiber.Ctx) (interface{}, error)
}

func (m *mockRoleService) CreateRole(c *fiber.Ctx, req *dto.CreateRoleRequest) (interface{}, error) {
	return m.CreateRoleFn(c, req)
}
func (m *mockRoleService) GetRole(c *fiber.Ctx, id string) (interface{}, error) {
	return m.GetRoleFn(c, id)
}
func (m *mockRoleService) GetAllRoles(c *fiber.Ctx) (interface{}, error) {
	return m.GetAllRolesFn(c)
}
func (m *mockRoleService) UpdateRole(c *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (interface{}, error) {
	return m.UpdateRoleFn(c, id, req)
}
func (m *mockRoleService) DeleteRole(c *fiber.Ctx, id string) error {
	return m.DeleteRoleFn(c, id)
}
func (m *mockRoleService) AssignRolesToUser(c *fiber.Ctx, req *dto.AssignRoleRequest) error {
	return m.AssignRolesToUserFn(c, req)
}
func (m *mockRoleService) GetUserRoles(c *fiber.Ctx, userID string) (interface{}, error) {
	return m.GetUserRolesFn(c, userID)
}
func (m *mockRoleService) GetAllPermissions(c *fiber.Ctx) (interface{}, error) {
	return m.GetAllPermissionsFn(c)
}

// responseEnvelope mirrors utils.SuccessResponse/ErrorResponse default structure:
// { "success": bool, "message": string, "data": any }
type responseEnvelope struct {
	Success bool        `json:"success"`
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func setupApp(handler *RoleHandler) *fiber.App {
	app := fiber.New()
	// Register endpoints exactly as handlers expect to be called
	app.Post("/roles", handler.CreateRole)
	app.Get("/roles/:id", handler.GetRole)
	app.Get("/roles", handler.GetAllRoles)
	app.Put("/roles/:id", handler.UpdateRole)
	app.Delete("/roles/:id", handler.DeleteRole)
	app.Post("/roles/assign", handler.AssignRolesToUser)
	app.Get("/me/roles", handler.GetUserRoles)
	app.Get("/permissions", handler.GetAllPermissions)
	return app
}

func parseResponse(tb testing.TB, res *http.Response) responseEnvelope {
	tb.Helper()
	defer res.Body.Close()
	var env responseEnvelope
	dec := json.NewDecoder(res.Body)
	if err := dec.Decode(&env); err \!= nil {
		tb.Fatalf("failed to decode response body: %v", err)
	}
	return env
}

// CreateRole tests
func TestCreateRole_Success(t *testing.T) {
	mock := &mockRoleService{
		CreateRoleFn: func(c *fiber.Ctx, req *dto.CreateRoleRequest) (interface{}, error) {
			return map[string]any{"id": "r1", "name": req.Name}, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	body := `{"name":"admin","permissions":["read","write"]}`
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}

	if resp.StatusCode \!= http.StatusCreated {
		t.Fatalf("expected status %d, got %d", http.StatusCreated, resp.StatusCode)
	}
	env := parseResponse(t, resp)
	if \!env.Success {
		t.Fatalf("expected success true")
	}
	if env.Message == "" {
		t.Fatalf("expected non-empty message")
	}
	data := env.Data.(map[string]any)
	if data["id"] \!= "r1" || data["name"] \!= "admin" {
		t.Fatalf("unexpected data: %#v", data)
	}
}

func TestCreateRole_InvalidBody(t *testing.T) {
	mock := &mockRoleService{
		CreateRoleFn: func(c *fiber.Ctx, req *dto.CreateRoleRequest) (interface{}, error) {
			t.Fatalf("service should not be called on invalid body")
			return nil, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	env := parseResponse(t, resp)
	if env.Success {
		t.Fatalf("expected success false")
	}
}

func TestCreateRole_ServiceError(t *testing.T) {
	mock := &mockRoleService{
		CreateRoleFn: func(c *fiber.Ctx, req *dto.CreateRoleRequest) (interface{}, error) {
			return nil, errors.New("duplicate role")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	body := `{"name":"admin"}`
	req := httptest.NewRequest(http.MethodPost, "/roles", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)

	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
	env := parseResponse(t, resp)
	if env.Success || env.Message == "" {
		t.Fatalf("expected error message, got: %#v", env)
	}
}

// GetRole tests
func TestGetRole_Success(t *testing.T) {
	mock := &mockRoleService{
		GetRoleFn: func(c *fiber.Ctx, id string) (interface{}, error) {
			if id \!= "r1" {
				t.Fatalf("expected id r1, got %s", id)
			}
			return map[string]any{"id": id, "name": "admin"}, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/roles/r1", nil)
	resp, _ := app.Test(req)

	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
	env := parseResponse(t, resp)
	if \!env.Success {
		t.Fatalf("expected success")
	}
}

func TestGetRole_MissingID(t *testing.T) {
	mock := &mockRoleService{
		GetRoleFn: func(c *fiber.Ctx, id string) (interface{}, error) {
			t.Fatalf("service should not be called when id missing")
			return nil, nil
		},
	}
	h := NewRoleHandler(mock)
	app := fiber.New()
	// Route without :id to force missing param
	app.Get("/roles", h.GetRole)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetRole_NotFound(t *testing.T) {
	mock := &mockRoleService{
		GetRoleFn: func(c *fiber.Ctx, id string) (interface{}, error) {
			return nil, errors.New("role not found")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/roles/unknown", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

// GetAllRoles tests
func TestGetAllRoles_Success(t *testing.T) {
	mock := &mockRoleService{
		GetAllRolesFn: func(c *fiber.Ctx) (interface{}, error) {
			return []map[string]any{{"id": "r1"}, {"id": "r2"}}, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
	env := parseResponse(t, resp)
	if \!env.Success {
		t.Fatalf("expected success")
	}
}

func TestGetAllRoles_Error(t *testing.T) {
	mock := &mockRoleService{
		GetAllRolesFn: func(c *fiber.Ctx) (interface{}, error) {
			return nil, errors.New("db down")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/roles", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// UpdateRole tests
func TestUpdateRole_Success(t *testing.T) {
	mock := &mockRoleService{
		UpdateRoleFn: func(c *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (interface{}, error) {
			return map[string]any{"id": id, "name": "manager"}, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	body := `{"name":"manager"}`
	req := httptest.NewRequest(http.MethodPut, "/roles/r1", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestUpdateRole_MissingID(t *testing.T) {
	mock := &mockRoleService{
		UpdateRoleFn: func(c *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (interface{}, error) {
			t.Fatalf("service should not be called when id missing")
			return nil, nil
		},
	}
	h := NewRoleHandler(mock)
	app := fiber.New()
	app.Put("/roles", h.UpdateRole) // no :id

	req := httptest.NewRequest(http.MethodPut, "/roles", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUpdateRole_InvalidBody(t *testing.T) {
	mock := &mockRoleService{
		UpdateRoleFn: func(c *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (interface{}, error) {
			t.Fatalf("service should not be called on invalid body")
			return nil, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodPut, "/roles/r1", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestUpdateRole_ServiceError(t *testing.T) {
	mock := &mockRoleService{
		UpdateRoleFn: func(c *fiber.Ctx, id string, req *dto.UpdateRoleRequest) (interface{}, error) {
			return nil, errors.New("cannot update system role")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodPut, "/roles/r1", bytes.NewBufferString(`{"name":"x"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// DeleteRole tests
func TestDeleteRole_Success(t *testing.T) {
	mock := &mockRoleService{
		DeleteRoleFn: func(c *fiber.Ctx, id string) error {
			return nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodDelete, "/roles/r1", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestDeleteRole_MissingID(t *testing.T) {
	mock := &mockRoleService{
		DeleteRoleFn: func(c *fiber.Ctx, id string) error {
			t.Fatalf("service should not be called when id missing")
			return nil
		},
	}
	h := NewRoleHandler(mock)
	app := fiber.New()
	app.Delete("/roles", h.DeleteRole) // no :id

	req := httptest.NewRequest(http.MethodDelete, "/roles", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestDeleteRole_ServiceError(t *testing.T) {
	mock := &mockRoleService{
		DeleteRoleFn: func(c *fiber.Ctx, id string) error {
			return errors.New("in use")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodDelete, "/roles/r1", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// AssignRolesToUser tests
func TestAssignRolesToUser_Success(t *testing.T) {
	mock := &mockRoleService{
		AssignRolesToUserFn: func(c *fiber.Ctx, req *dto.AssignRoleRequest) error {
			return nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	body := `{"userId":"u1","roleIds":["r1","r2"]}`
	req := httptest.NewRequest(http.MethodPost, "/roles/assign", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestAssignRolesToUser_InvalidBody(t *testing.T) {
	mock := &mockRoleService{
		AssignRolesToUserFn: func(c *fiber.Ctx, req *dto.AssignRoleRequest) error {
			t.Fatalf("service should not be called on invalid body")
			return nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodPost, "/roles/assign", bytes.NewBufferString("{bad"))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestAssignRolesToUser_ServiceError(t *testing.T) {
	mock := &mockRoleService{
		AssignRolesToUserFn: func(c *fiber.Ctx, req *dto.AssignRoleRequest) error {
			return errors.New("invalid roles")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	body := `{"userId":"u1","roleIds":["bad"]}`
	req := httptest.NewRequest(http.MethodPost, "/roles/assign", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

// GetUserRoles tests
func TestGetUserRoles_Success(t *testing.T) {
	mock := &mockRoleService{
		GetUserRolesFn: func(c *fiber.Ctx, userID string) (interface{}, error) {
			return []map[string]any{{"id": "r1"}}, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/me/roles", nil)
	req.Header.Set("X-User-Id", "u1")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetUserRoles_MissingHeader(t *testing.T) {
	mock := &mockRoleService{
		GetUserRolesFn: func(c *fiber.Ctx, userID string) (interface{}, error) {
			t.Fatalf("service should not be called when header missing")
			return nil, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/me/roles", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected %d, got %d", http.StatusBadRequest, resp.StatusCode)
	}
}

func TestGetUserRoles_NotFound(t *testing.T) {
	mock := &mockRoleService{
		GetUserRolesFn: func(c *fiber.Ctx, userID string) (interface{}, error) {
			return nil, errors.New("user not found")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/me/roles", nil)
	req.Header.Set("X-User-Id", "u404")
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusNotFound {
		t.Fatalf("expected %d, got %d", http.StatusNotFound, resp.StatusCode)
	}
}

// GetAllPermissions tests
func TestGetAllPermissions_Success(t *testing.T) {
	mock := &mockRoleService{
		GetAllPermissionsFn: func(c *fiber.Ctx) (interface{}, error) {
			return []string{"read", "write"}, nil
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusOK {
		t.Fatalf("expected %d, got %d", http.StatusOK, resp.StatusCode)
	}
}

func TestGetAllPermissions_Error(t *testing.T) {
	mock := &mockRoleService{
		GetAllPermissionsFn: func(c *fiber.Ctx) (interface{}, error) {
			return nil, errors.New("permissions unavailable")
		},
	}
	h := NewRoleHandler(mock)
	app := setupApp(h)

	req := httptest.NewRequest(http.MethodGet, "/permissions", nil)
	resp, _ := app.Test(req)
	if resp.StatusCode \!= http.StatusInternalServerError {
		t.Fatalf("expected %d, got %d", http.StatusInternalServerError, resp.StatusCode)
	}
}

// Compile-time conformance check to ensure mock satisfies the interface
var _ services.RoleService = (*mockRoleService)(nil)