package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/database"
	"dojo6/backend/internal/models"
)

func setupTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatalf("open test db: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db, database.MigrationsFS); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return db
}

func newTestJWTService() *auth.JWTService {
	return auth.NewJWTService("test-secret-key", time.Hour)
}

func tokenForUser(t *testing.T, jwtSvc *auth.JWTService, id int64, email, role string) string {
	t.Helper()
	token, err := jwtSvc.GenerateToken(id, email, role)
	if err != nil {
		t.Fatalf("generate token: %v", err)
	}
	return token
}

func seedUser(t *testing.T, repo *models.UserRepository, name, email, role string) *models.User {
	t.Helper()
	hash, err := auth.HashPassword("password123")
	if err != nil {
		t.Fatalf("hash password: %v", err)
	}
	u := &models.User{
		Name:         name,
		Email:        email,
		Phone:        "555-0100",
		Role:         role,
		PasswordHash: hash,
	}
	if err := repo.Create(u); err != nil {
		t.Fatalf("seed user %s: %v", email, err)
	}
	return u
}

type testEnv struct {
	router  chi.Router
	jwtSvc  *auth.JWTService
	repo    *models.UserRepository
	handler *UserHandler
}

func setupTestEnv(t *testing.T) *testEnv {
	t.Helper()
	db := setupTestDB(t)
	jwtSvc := newTestJWTService()
	repo := models.NewUserRepository(db)
	h := NewUserHandler(repo)

	r := chi.NewRouter()
	r.Route("/api/users", func(ur chi.Router) {
		ur.Use(auth.JWTMiddleware(jwtSvc))
		ur.Get("/", h.List)
		ur.Post("/", h.Create)
		ur.Get("/{id}", h.GetByID)
		ur.Put("/{id}", h.Update)
		ur.Delete("/{id}", h.Delete)
		ur.Put("/{id}/role", h.ChangeRole)
		ur.Put("/{id}/password", h.ChangePassword)
	})

	return &testEnv{router: r, jwtSvc: jwtSvc, repo: repo, handler: h}
}

func doRequest(t *testing.T, env *testEnv, method, path, token string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var reqBody *bytes.Buffer
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshal body: %v", err)
		}
		reqBody = bytes.NewBuffer(b)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	req := httptest.NewRequest(method, path, reqBody)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	rec := httptest.NewRecorder()
	env.router.ServeHTTP(rec, req)
	return rec
}

func decodeJSON(t *testing.T, rec *httptest.ResponseRecorder, v interface{}) {
	t.Helper()
	if err := json.NewDecoder(rec.Body).Decode(v); err != nil {
		t.Fatalf("decode response: %v (body: %s)", err, rec.Body.String())
	}
}

// --- List Tests ---

func TestList_AdminCanListUsers(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	seedUser(t, env.repo, "Alice", "alice@test.com", "user")
	seedUser(t, env.repo, "Bob", "bob@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodGet, "/api/users", token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var users []userResponse
	decodeJSON(t, rec, &users)
	if len(users) != 3 {
		t.Errorf("expected 3 users, got %d", len(users))
	}
}

func TestList_InstructorCanListUsers(t *testing.T) {
	env := setupTestEnv(t)
	instructor := seedUser(t, env.repo, "Instructor", "inst@test.com", "instructor")
	seedUser(t, env.repo, "Alice", "alice@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, instructor.ID, instructor.Email, instructor.Role)
	rec := doRequest(t, env, http.MethodGet, "/api/users", token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestList_RegularUserForbidden(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	rec := doRequest(t, env, http.MethodGet, "/api/users", token, nil)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestList_Unauthenticated(t *testing.T) {
	env := setupTestEnv(t)
	rec := doRequest(t, env, http.MethodGet, "/api/users", "", nil)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestList_ExcludesDeletedUsers(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	toDelete := seedUser(t, env.repo, "Deleted", "deleted@test.com", "user")

	if err := env.repo.SoftDelete(toDelete.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodGet, "/api/users", token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var users []userResponse
	decodeJSON(t, rec, &users)
	if len(users) != 1 {
		t.Errorf("expected 1 user (deleted excluded), got %d", len(users))
	}
}

// --- Create Tests ---

func TestCreate_AdminCanCreateUser(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{
		"name":     "New User",
		"email":    "new@test.com",
		"phone":    "555-0200",
		"password": "secret123",
		"role":     "user",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp userResponse
	decodeJSON(t, rec, &resp)
	if resp.Name != "New User" {
		t.Errorf("expected name 'New User', got %s", resp.Name)
	}
	if resp.Email != "new@test.com" {
		t.Errorf("expected email 'new@test.com', got %s", resp.Email)
	}
	if resp.Role != "user" {
		t.Errorf("expected role 'user', got %s", resp.Role)
	}
}

func TestCreate_DefaultsToUserRole(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{
		"name":     "No Role",
		"email":    "norole@test.com",
		"phone":    "555-0200",
		"password": "secret123",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp userResponse
	decodeJSON(t, rec, &resp)
	if resp.Role != "user" {
		t.Errorf("expected default role 'user', got %s", resp.Role)
	}
}

func TestCreate_NonAdminForbidden(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	body := map[string]interface{}{
		"name":     "New",
		"email":    "new@test.com",
		"phone":    "555-0200",
		"password": "secret",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCreate_InstructorForbidden(t *testing.T) {
	env := setupTestEnv(t)
	inst := seedUser(t, env.repo, "Inst", "inst@test.com", "instructor")

	token := tokenForUser(t, env.jwtSvc, inst.ID, inst.Email, inst.Role)
	body := map[string]interface{}{
		"name":     "New",
		"email":    "new@test.com",
		"phone":    "555-0200",
		"password": "secret",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestCreate_DuplicateEmail(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{
		"name":     "Dup",
		"email":    "admin@test.com",
		"phone":    "555-0200",
		"password": "secret",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestCreate_MissingRequiredFields(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)

	tests := []struct {
		name string
		body map[string]interface{}
	}{
		{name: "missing name", body: map[string]interface{}{"email": "a@b.com", "phone": "555", "password": "pass"}},
		{name: "missing email", body: map[string]interface{}{"name": "A", "phone": "555", "password": "pass"}},
		{name: "missing phone", body: map[string]interface{}{"name": "A", "email": "a@b.com", "password": "pass"}},
		{name: "missing password", body: map[string]interface{}{"name": "A", "email": "a@b.com", "phone": "555"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := doRequest(t, env, http.MethodPost, "/api/users", token, tt.body)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}

func TestCreate_InvalidRole(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)

	body := map[string]interface{}{
		"name":     "Bad Role",
		"email":    "bad@test.com",
		"phone":    "555",
		"password": "pass",
		"role":     "superadmin",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestCreate_DoesNotExposePasswordHash(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)

	body := map[string]interface{}{
		"name":     "New",
		"email":    "new@test.com",
		"phone":    "555",
		"password": "secret",
	}
	rec := doRequest(t, env, http.MethodPost, "/api/users", token, body)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rec.Code)
	}

	// Ensure password_hash is not in the response.
	var raw map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&raw); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if _, ok := raw["password_hash"]; ok {
		t.Error("response should not contain password_hash")
	}
}

// --- GetByID Tests ---

func TestGetByID_AdminCanGetAnyUser(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodGet, fmt.Sprintf("/api/users/%d", user.ID), token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp userResponse
	decodeJSON(t, rec, &resp)
	if resp.Email != "user@test.com" {
		t.Errorf("expected email user@test.com, got %s", resp.Email)
	}
}

func TestGetByID_InstructorCanGetAnyUser(t *testing.T) {
	env := setupTestEnv(t)
	inst := seedUser(t, env.repo, "Inst", "inst@test.com", "instructor")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, inst.ID, inst.Email, inst.Role)
	rec := doRequest(t, env, http.MethodGet, fmt.Sprintf("/api/users/%d", user.ID), token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetByID_UserCanGetSelf(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	rec := doRequest(t, env, http.MethodGet, fmt.Sprintf("/api/users/%d", user.ID), token, nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}
}

func TestGetByID_UserCannotGetOther(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")
	other := seedUser(t, env.repo, "Other", "other@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	rec := doRequest(t, env, http.MethodGet, fmt.Sprintf("/api/users/%d", other.ID), token, nil)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestGetByID_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodGet, "/api/users/9999", token, nil)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

func TestGetByID_DeletedUserNotFound(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	deleted := seedUser(t, env.repo, "Deleted", "del@test.com", "user")

	if err := env.repo.SoftDelete(deleted.ID); err != nil {
		t.Fatalf("soft delete: %v", err)
	}

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodGet, fmt.Sprintf("/api/users/%d", deleted.ID), token, nil)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404 for deleted user, got %d", rec.Code)
	}
}

func TestGetByID_InvalidID(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodGet, "/api/users/abc", token, nil)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Update Tests ---

func TestUpdate_AdminCanUpdateAnyUser(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	newName := "Updated User"
	body := map[string]interface{}{
		"name": newName,
	}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d", user.ID), token, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp userResponse
	decodeJSON(t, rec, &resp)
	if resp.Name != newName {
		t.Errorf("expected name %q, got %s", newName, resp.Name)
	}
}

func TestUpdate_UserCanUpdateSelf(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	body := map[string]interface{}{
		"phone": "555-9999",
	}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d", user.ID), token, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp userResponse
	decodeJSON(t, rec, &resp)
	if resp.Phone != "555-9999" {
		t.Errorf("expected phone '555-9999', got %s", resp.Phone)
	}
}

func TestUpdate_UserCannotUpdateOther(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")
	other := seedUser(t, env.repo, "Other", "other@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	body := map[string]interface{}{"name": "Hacked"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d", other.ID), token, body)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestUpdate_EmptyNameRejected(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	emptyName := ""
	body := map[string]interface{}{"name": emptyName}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d", user.ID), token, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestUpdate_DuplicateEmailRejected(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	seedUser(t, env.repo, "User", "user@test.com", "user")
	other := seedUser(t, env.repo, "Other", "other@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"email": "user@test.com"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d", other.ID), token, body)

	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestUpdate_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"name": "Ghost"}
	rec := doRequest(t, env, http.MethodPut, "/api/users/9999", token, body)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- Delete Tests ---

func TestDelete_AdminCanDeleteUser(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodDelete, fmt.Sprintf("/api/users/%d", user.ID), token, nil)

	if rec.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rec.Code)
	}

	// Verify user is soft-deleted.
	var u models.User
	if err := env.repo.GetByID(user.ID, &u); err != nil {
		t.Fatalf("get after delete: %v", err)
	}
	if u.DeletedAt == nil {
		t.Error("expected deleted_at to be set")
	}
}

func TestDelete_NonAdminForbidden(t *testing.T) {
	env := setupTestEnv(t)
	seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	rec := doRequest(t, env, http.MethodDelete, fmt.Sprintf("/api/users/%d", user.ID), token, nil)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestDelete_InstructorForbidden(t *testing.T) {
	env := setupTestEnv(t)
	inst := seedUser(t, env.repo, "Inst", "inst@test.com", "instructor")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, inst.ID, inst.Email, inst.Role)
	rec := doRequest(t, env, http.MethodDelete, fmt.Sprintf("/api/users/%d", user.ID), token, nil)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestDelete_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	rec := doRequest(t, env, http.MethodDelete, "/api/users/9999", token, nil)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- ChangeRole Tests ---

func TestChangeRole_AdminCanChangeRole(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"role": "instructor"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d/role", user.ID), token, body)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var resp userResponse
	decodeJSON(t, rec, &resp)
	if resp.Role != "instructor" {
		t.Errorf("expected role 'instructor', got %s", resp.Role)
	}
}

func TestChangeRole_NonAdminForbidden(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	body := map[string]interface{}{"role": "admin"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d/role", user.ID), token, body)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestChangeRole_InvalidRole(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"role": "superadmin"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d/role", user.ID), token, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestChangeRole_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"role": "instructor"}
	rec := doRequest(t, env, http.MethodPut, "/api/users/9999/role", token, body)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}

// --- ChangePassword Tests ---

func TestChangePassword_AdminCanResetPassword(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"password": "newpassword123"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d/password", user.ID), token, body)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body.String())
	}

	// Verify new password works.
	var u models.User
	if err := env.repo.GetByID(user.ID, &u); err != nil {
		t.Fatalf("get user: %v", err)
	}
	if err := auth.CheckPassword(u.PasswordHash, "newpassword123"); err != nil {
		t.Error("expected new password to match")
	}
	if err := auth.CheckPassword(u.PasswordHash, "password123"); err == nil {
		t.Error("expected old password to no longer match")
	}
}

func TestChangePassword_NonAdminForbidden(t *testing.T) {
	env := setupTestEnv(t)
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, user.ID, user.Email, user.Role)
	body := map[string]interface{}{"password": "newpass"}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d/password", user.ID), token, body)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

func TestChangePassword_EmptyPasswordRejected(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")
	user := seedUser(t, env.repo, "User", "user@test.com", "user")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"password": ""}
	rec := doRequest(t, env, http.MethodPut, fmt.Sprintf("/api/users/%d/password", user.ID), token, body)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestChangePassword_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	admin := seedUser(t, env.repo, "Admin", "admin@test.com", "admin")

	token := tokenForUser(t, env.jwtSvc, admin.ID, admin.Email, admin.Role)
	body := map[string]interface{}{"password": "newpass"}
	rec := doRequest(t, env, http.MethodPut, "/api/users/9999/password", token, body)

	if rec.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rec.Code)
	}
}
