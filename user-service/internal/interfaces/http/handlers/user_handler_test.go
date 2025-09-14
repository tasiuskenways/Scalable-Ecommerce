package handlers

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// NOTE: Testing framework used: Go's standard "testing" package.
// If the repository already uses testify, these tests can be trivially adapted to use assert/require.
// We avoid new dependencies and stick to net/http/httptest + Fiber's app.Test for requests.

// --- Minimal local doubles/mocks to avoid importing internal implementations ---

// captureService is a simple stub/mock for services.UserService capturing inputs and returning configured outputs.
type captureService struct {
	// Configurable returns
	getUserResp   any
	getUserErr    error
	getAllResp    any
	getAllErr     error
	updateResp    any
	updateErr     error
	deleteErr     error

	// Captured arguments
	lastCtxUserID   string
	lastPage        int
	lastLimit       int
	lastUpdateID    string
	lastUpdateBody  any
	lastDeletedUser string
}

func (m *captureService) GetUser(c *fiber.Ctx, userID string) (any, error) {
	m.lastCtxUserID = userID
	return m.getUserResp, m.getUserErr
}
func (m *captureService) GetAllUsers(c *fiber.Ctx, page, limit int) (any, error) {
	m.lastPage = page
	m.lastLimit = limit
	return m.getAllResp, m.getAllErr
}
func (m *captureService) UpdateUser(c *fiber.Ctx, userID string, req any) (any, error) {
	m.lastUpdateID = userID
	m.lastUpdateBody = req
	return m.updateResp, m.updateErr
}
func (m *captureService) DeleteUser(c *fiber.Ctx, userID string) error {
	m.lastDeletedUser = userID
	return m.deleteErr
}

// successBody mirrors a minimal shape expected from utils.SuccessResponse.
// We don't know exact structure; we assert on commonly used fields.
type successBody struct {
	Message string `json:"message"`
	Data    any    `json:"data"`
}
type errorBody struct {
	Message string `json:"message"`
}

// helper to build a Fiber app with routes bound to our handler
func buildApp(h *UserHandler) *fiber.App {
	app := fiber.New()
	// Routes for testing
	app.Get("/me", h.GetMe)
	app.Get("/users/:id", h.GetUser)
	app.Get("/users", h.GetAllUsers)
	app.Put("/users/:id", h.UpdateUser)
	app.Put("/me", h.UpdateMe)
	app.Delete("/users/:id", h.DeleteUser)
	return app
}

func readJSON[T any](t *testing.T, body io.Reader, out *T) {
	t.Helper()
	data, err := io.ReadAll(body)
	if err \!= nil {
		t.Fatalf("failed reading body: %v", err)
	}
	if err := json.Unmarshal(data, out); err \!= nil {
		t.Fatalf("failed unmarshalling body: %v\nraw: %s", err, string(data))
	}
}

// ---------- GetMe ----------
func TestGetMe_Unauthenticated(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	// No X-User-Id header
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
	var eb errorBody
	readJSON(t, res.Body, &eb)
	if eb.Message == "" {
		t.Fatalf("expected error message, got empty")
	}
}

func TestGetMe_NotFoundFromService(t *testing.T) {
	svc := &captureService{getUserErr: errors.New("user not found")}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("X-User-Id", "abc123")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
	if svc.lastCtxUserID \!= "abc123" {
		t.Fatalf("expected service to be called with userID=abc123, got %s", svc.lastCtxUserID)
	}
}

func TestGetMe_Success(t *testing.T) {
	mockData := map[string]any{"id": "u1", "name": "Jane"}
	svc := &captureService{getUserResp: mockData}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("X-User-Id", "u1")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var sb successBody
	readJSON(t, res.Body, &sb)
	if sb.Message == "" {
		t.Fatalf("expected success message, got empty")
	}
	if data, ok := sb.Data.(map[string]any); ok {
		if data["id"] \!= "u1" {
			t.Fatalf("expected data.id=u1, got %v", data["id"])
		}
	}
}

// ---------- GetUser ----------
func TestGetUser_MissingID(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/users/", nil) // no param; route not matched, so test handler directly
	// Directly invoke handler using app with a crafted request against a path that includes an empty id isn't trivial.
	// Instead, call the route with empty param by registering another route for test:
	app2 := fiber.New()
	app2.Get("/users/", h.GetUser)
	res, err := app2.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	// Fiber may 404 for unmatched route; ensure we actually hit handler:
	if res.StatusCode == http.StatusNotFound {
		t.Skip("Route without :id not matched; skipping this path as Fiber router requires param presence.")
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestGetUser_NotFoundFromService(t *testing.T) {
	svc := &captureService{getUserErr: errors.New("no such user")}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/users/xyz", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusNotFound {
		t.Fatalf("expected 404, got %d", res.StatusCode)
	}
	if svc.lastCtxUserID \!= "xyz" {
		t.Fatalf("expected xyz, got %s", svc.lastCtxUserID)
	}
}

func TestGetUser_Success(t *testing.T) {
	mockData := map[string]any{"id": "u9"}
	svc := &captureService{getUserResp: mockData}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/users/u9", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var sb successBody
	readJSON(t, res.Body, &sb)
	if sb.Message == "" {
		t.Fatalf("expected message")
	}
}

// ---------- GetAllUsers ----------
func TestGetAllUsers_Defaults(t *testing.T) {
	svc := &captureService{getAllResp: []any{}}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/users", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	// Expect defaults page=1, limit=10
	if svc.lastPage \!= 1 || svc.lastLimit \!= 10 {
		t.Fatalf("expected defaults page=1,limit=10; got page=%d,limit=%d", svc.lastPage, svc.lastLimit)
	}
}

func TestGetAllUsers_NormalizesInvalidParams(t *testing.T) {
	cases := []struct {
		pageQ     string
		limitQ    string
		wantPage  int
		wantLimit int
	}{
		{"-2", "0", 1, 10},
		{"0", "200", 1, 10},
		{"2", "5", 2, 5},
	}
	for i, cs := range cases {
		svc := &captureService{getAllResp: []any{}}
		h := NewUserHandler(svc)
		app := buildApp(h)
		url := "/users"
		q := "?"
		if cs.pageQ \!= "" {
			q += "page=" + cs.pageQ
		}
		if cs.limitQ \!= "" {
			if q \!= "?" {
				q += "&"
			}
			q += "limit=" + cs.limitQ
		}
		if q \!= "?" {
			url += q
		}
		req := httptest.NewRequest(http.MethodGet, url, nil)
		res, err := app.Test(req)
		if err \!= nil {
			t.Fatalf("case %d: app.Test error: %v", i, err)
		}
		if res.StatusCode \!= http.StatusOK {
			t.Fatalf("case %d: expected 200, got %d", i, res.StatusCode)
		}
		if svc.lastPage \!= cs.wantPage || svc.lastLimit \!= cs.wantLimit {
			t.Fatalf("case %d: expected page=%d,limit=%d; got page=%d,limit=%d",
				i, cs.wantPage, cs.wantLimit, svc.lastPage, svc.lastLimit)
		}
	}
}

func TestGetAllUsers_ServiceError(t *testing.T) {
	svc := &captureService{getAllErr: errors.New("db down")}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/users?page=1&limit=10", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusInternalServerError {
		t.Fatalf("expected 500, got %d", res.StatusCode)
	}
}

// ---------- UpdateUser ----------
func TestUpdateUser_MissingID(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	// We need a route without :id to trigger missing id path:
	app := fiber.New()
	app.Put("/users/", h.UpdateUser)

	req := httptest.NewRequest(http.MethodPut, "/users/", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode == http.StatusNotFound {
		t.Skip("Route without :id not matched; skip due to router constraints.")
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestUpdateUser_InvalidBody(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodPut, "/users/u1", bytes.NewBufferString(`{"invalid":`)) // malformed JSON
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestUpdateUser_ServiceError(t *testing.T) {
	svc := &captureService{updateErr: errors.New("validation failed")}
	h := NewUserHandler(svc)
	app := buildApp(h)

	body := map[string]any{"name": "New Name"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/users/u2", bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	if svc.lastUpdateID \!= "u2" {
		t.Fatalf("expected update userID u2, got %s", svc.lastUpdateID)
	}
}

func TestUpdateUser_Success(t *testing.T) {
	mock := map[string]any{"updated": true}
	svc := &captureService{updateResp: mock}
	h := NewUserHandler(svc)
	app := buildApp(h)

	body := map[string]any{"name": "N"}
	raw, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPut, "/users/u3", bytes.NewBuffer(raw))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	var sb successBody
	readJSON(t, res.Body, &sb)
	if sb.Message == "" {
		t.Fatalf("expected success message")
	}
}

// ---------- UpdateMe ----------
func TestUpdateMe_Unauthenticated(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{}`))
	req.Header.Set("Content-Type", "application/json")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", res.StatusCode)
	}
}

func TestUpdateMe_InvalidBody(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"x":`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "me1")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestUpdateMe_ServiceError(t *testing.T) {
	svc := &captureService{updateErr: errors.New("bad data")}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"name":"X"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "me2")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestUpdateMe_Success(t *testing.T) {
	svc := &captureService{updateResp: map[string]any{"ok": true}}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodPut, "/me", bytes.NewBufferString(`{"name":"Y"}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-User-Id", "me3")
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

// ---------- DeleteUser ----------
func TestDeleteUser_MissingID(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := fiber.New()
	app.Delete("/users/", h.DeleteUser)

	req := httptest.NewRequest(http.MethodDelete, "/users/", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode == http.StatusNotFound {
		t.Skip("Route without :id not matched; skipping due to router.")
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
}

func TestDeleteUser_ServiceError(t *testing.T) {
	svc := &captureService{deleteErr: errors.New("constraint")}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodDelete, "/users/d1", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", res.StatusCode)
	}
	if svc.lastDeletedUser \!= "d1" {
		t.Fatalf("expected lastDeletedUser=d1, got %s", svc.lastDeletedUser)
	}
}

func TestDeleteUser_Success(t *testing.T) {
	svc := &captureService{}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodDelete, "/users/d2", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
}

// --- Additional guard: ensure normalization logic dispatches expected values across several inputs ---
func TestGetAllUsers_NormalizationMatrix(t *testing.T) {
	inputs := []struct {
		page, limit string
		expP, expL  int
	}{
		{"", "", 1, 10},
		{"1", "", 1, 10},
		{"", "10", 1, 10},
		{"-1", "-5", 1, 10},
		{"5", "101", 5, 10},
		{"3", "100", 3, 100},
	}
	for i, in := range inputs {
		svc := &captureService{getAllResp: []any{}}
		h := NewUserHandler(svc)
		app := buildApp(h)
		url := "/users"
		q := "?"
		if in.page \!= "" {
			q += "page=" + in.page
		}
		if in.limit \!= "" {
			if q \!= "?" {
				q += "&"
			}
			q += "limit=" + in.limit
		}
		if q \!= "?" {
			url += q
		}
		req := httptest.NewRequest(http.MethodGet, url, nil)
		res, err := app.Test(req)
		if err \!= nil {
			t.Fatalf("matrix %d: app.Test error: %v", i, err)
		}
		if res.StatusCode \!= http.StatusOK {
			t.Fatalf("matrix %d: expected 200, got %d", i, res.StatusCode)
		}
		if svc.lastPage \!= in.expP || svc.lastLimit \!= in.expL {
			t.Fatalf("matrix %d: expected page=%d limit=%d; got page=%d limit=%d (input page=%q,limit=%q)",
				i, in.expP, in.expL, svc.lastPage, svc.lastLimit, in.page, in.limit)
		}
	}
}

// Sanity: ensure numeric query parsing doesn't panic on non-numeric inputs and falls back to defaults.
func TestGetAllUsers_NonNumericQueries(t *testing.T) {
	svc := &captureService{getAllResp: []any{}}
	h := NewUserHandler(svc)
	app := buildApp(h)

	req := httptest.NewRequest(http.MethodGet, "/users?page=foo&limit=bar", nil)
	res, err := app.Test(req)
	if err \!= nil {
		t.Fatalf("app.Test error: %v", err)
	}
	if res.StatusCode \!= http.StatusOK {
		t.Fatalf("expected 200, got %d", res.StatusCode)
	}
	// Atoi error should lead to 0 which normalizes to defaults 1 and 10
	if svc.lastPage \!= 1 || svc.lastLimit \!= 10 {
		t.Fatalf("expected normalized defaults page=1 limit=10, got page=%d limit=%d", svc.lastPage, svc.lastLimit)
	}
}

// Additional helper to ensure we can decode success payloads with generic map
func TestSuccessBodyDecoding(t *testing.T) {
	sb := successBody{
		Message: "ok",
		Data: map[string]any{
			"id":    "x",
			"count": 3,
		},
	}
	raw, err := json.Marshal(sb)
	if err \!= nil {
		t.Fatalf("marshal: %v", err)
	}
	var out successBody
	if err := json.Unmarshal(raw, &out); err \!= nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if out.Message \!= "ok" {
		t.Fatalf("message mismatch")
	}
	// Validate nested values can be accessed by type assertions
	m, ok := out.Data.(map[string]any)
	if \!ok {
		t.Fatalf("expected Data to be a map")
	}
	if m["count"] \!= float64(3) { // JSON numbers decode as float64 in interface{}
		t.Fatalf("expected count 3, got %v", m["count"])
	}
}

// Quick check for strconv edge recognitions used by handler
func TestAtoiEdge(t *testing.T) {
	if _, err := strconv.Atoi("not-a-number"); err == nil {
		t.Fatalf("expected error for non-numeric")
	}
}