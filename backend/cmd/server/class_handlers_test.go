package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"
	"dojo6/backend/internal/repository"
)

func seedAdminToken(t *testing.T, env testEnv) string {
	t.Helper()
	hash, _ := auth.HashPassword("pass")
	u := models.User{Name: "Admin", Email: "admin@test.com", Role: "admin", PasswordHash: hash}
	if err := env.Auth.Users.Create(&u); err != nil {
		t.Fatal(err)
	}
	tok, _ := auth.GenerateToken(testJWTSecret, u.ID, u.Email, u.Role)
	return tok
}

func seedUserToken(t *testing.T, env testEnv, role string) (string, models.User) {
	t.Helper()
	hash, _ := auth.HashPassword("pass")
	u := models.User{Name: "User", Email: role + "@test.com", Role: role, PasswordHash: hash}
	if err := env.Auth.Users.Create(&u); err != nil {
		t.Fatal(err)
	}
	tok, _ := auth.GenerateToken(testJWTSecret, u.ID, u.Email, u.Role)
	return tok, u
}

// --- Class Type Tests ---

func TestClassTypes_CRUD(t *testing.T) {
	env := testEnv2(t)
	token := seedAdminToken(t, env)

	// Create
	body := `{"name":"Yoga","description":"Relaxation class"}`
	req := httptest.NewRequest(http.MethodPost, "/api/class-types", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created classTypeJSON
	json.NewDecoder(rec.Body).Decode(&created)
	if created.Name != "Yoga" {
		t.Errorf("expected name Yoga, got %s", created.Name)
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/class-types/%d", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", rec.Code)
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/api/class-types", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	var list []classTypeJSON
	json.NewDecoder(rec.Body).Decode(&list)
	if len(list) != 1 {
		t.Errorf("expected 1 class type, got %d", len(list))
	}

	// Update
	body = `{"name":"Karate","description":"Martial arts"}`
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/class-types/%d", created.ID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updated classTypeJSON
	json.NewDecoder(rec.Body).Decode(&updated)
	if updated.Name != "Karate" {
		t.Errorf("expected name Karate, got %s", updated.Name)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/class-types/%d", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}

	// Get after delete returns 404
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/class-types/%d", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("get deleted: expected 404, got %d", rec.Code)
	}
}

func TestClassTypes_AdminOnly(t *testing.T) {
	env := testEnv2(t)
	_ = seedAdminToken(t, env)
	userTok, _ := seedUserToken(t, env, "user")
	instTok, _ := seedUserToken(t, env, "instructor")

	tests := []struct {
		name   string
		token  string
		method string
		path   string
		body   string
		expect int
	}{
		{"user cannot create", userTok, http.MethodPost, "/api/class-types", `{"name":"X"}`, http.StatusForbidden},
		{"instructor cannot create", instTok, http.MethodPost, "/api/class-types", `{"name":"X"}`, http.StatusForbidden},
		{"user cannot update", userTok, http.MethodPut, "/api/class-types/1", `{"name":"X"}`, http.StatusForbidden},
		{"user cannot delete", userTok, http.MethodDelete, "/api/class-types/1", "", http.StatusForbidden},
		// List and get should work for all authenticated users
		{"user can list", userTok, http.MethodGet, "/api/class-types", "", http.StatusOK},
		{"instructor can list", instTok, http.MethodGet, "/api/class-types", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bodyReader *strings.Reader
			if tt.body != "" {
				bodyReader = strings.NewReader(tt.body)
			} else {
				bodyReader = strings.NewReader("")
			}
			req := httptest.NewRequest(tt.method, tt.path, bodyReader)
			req.Header.Set("Authorization", "Bearer "+tt.token)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			env.Router.ServeHTTP(rec, req)
			if rec.Code != tt.expect {
				t.Errorf("expected %d, got %d: %s", tt.expect, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestClassTypes_Unauthenticated(t *testing.T) {
	env := testEnv2(t)

	req := httptest.NewRequest(http.MethodGet, "/api/class-types", nil)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestClassTypes_ValidationErrors(t *testing.T) {
	env := testEnv2(t)
	token := seedAdminToken(t, env)

	// Missing name
	body := `{"description":"no name"}`
	req := httptest.NewRequest(http.MethodPost, "/api/class-types", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// --- Class Tests ---

func TestClasses_CRUD(t *testing.T) {
	env := testEnv2(t)
	token := seedAdminToken(t, env)

	// Create a class type and instructor first.
	ct, _ := repository.CreateClassType(env.DB, "Yoga", "")
	hash, _ := auth.HashPassword("pass")
	instructor := models.User{Name: "Inst", Email: "inst@test.com", Role: "instructor", PasswordHash: hash}
	env.Auth.Users.Create(&instructor)

	// Create class
	body := fmt.Sprintf(`{"class_type_id":%d,"instructor_id":%d,"start_time":"2026-03-01T10:00:00Z","duration_minutes":60,"capacity":20}`, ct.ID, instructor.ID)
	req := httptest.NewRequest(http.MethodPost, "/api/classes", strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var created classJSON
	json.NewDecoder(rec.Body).Decode(&created)
	if created.ClassTypeID != ct.ID {
		t.Errorf("expected class_type_id %d, got %d", ct.ID, created.ClassTypeID)
	}
	if created.Capacity != 20 {
		t.Errorf("expected capacity 20, got %d", created.Capacity)
	}

	// Get
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/classes/%d", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("get: expected 200, got %d", rec.Code)
	}

	// List
	req = httptest.NewRequest(http.MethodGet, "/api/classes", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	var classList []classJSON
	json.NewDecoder(rec.Body).Decode(&classList)
	if len(classList) != 1 {
		t.Errorf("expected 1 class, got %d", len(classList))
	}

	// Update
	body = fmt.Sprintf(`{"class_type_id":%d,"instructor_id":%d,"start_time":"2026-03-01T11:00:00Z","duration_minutes":90,"capacity":25}`, ct.ID, instructor.ID)
	req = httptest.NewRequest(http.MethodPut, fmt.Sprintf("/api/classes/%d", created.ID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("update: expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var updatedClass classJSON
	json.NewDecoder(rec.Body).Decode(&updatedClass)
	if updatedClass.Capacity != 25 {
		t.Errorf("expected capacity 25, got %d", updatedClass.Capacity)
	}
	if updatedClass.DurationMinutes != 90 {
		t.Errorf("expected duration 90, got %d", updatedClass.DurationMinutes)
	}

	// Delete
	req = httptest.NewRequest(http.MethodDelete, fmt.Sprintf("/api/classes/%d", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}

	// Get after delete
	req = httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/classes/%d", created.ID), nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusNotFound {
		t.Errorf("get deleted: expected 404, got %d", rec.Code)
	}
}

func TestClasses_AdminOnly(t *testing.T) {
	env := testEnv2(t)
	_ = seedAdminToken(t, env)
	userTok, _ := seedUserToken(t, env, "user")
	instTok, _ := seedUserToken(t, env, "instructor")

	tests := []struct {
		name   string
		token  string
		method string
		path   string
		body   string
		expect int
	}{
		{"user cannot create", userTok, http.MethodPost, "/api/classes", `{"class_type_id":1,"instructor_id":1,"start_time":"2026-03-01T10:00:00Z","duration_minutes":60,"capacity":20}`, http.StatusForbidden},
		{"instructor cannot create", instTok, http.MethodPost, "/api/classes", `{"class_type_id":1,"instructor_id":1,"start_time":"2026-03-01T10:00:00Z","duration_minutes":60,"capacity":20}`, http.StatusForbidden},
		{"user cannot update", userTok, http.MethodPut, "/api/classes/1", `{"class_type_id":1,"instructor_id":1,"start_time":"2026-03-01T10:00:00Z","duration_minutes":60,"capacity":20}`, http.StatusForbidden},
		{"user cannot delete", userTok, http.MethodDelete, "/api/classes/1", "", http.StatusForbidden},
		// List should work for all authenticated users
		{"user can list", userTok, http.MethodGet, "/api/classes", "", http.StatusOK},
		{"instructor can list", instTok, http.MethodGet, "/api/classes", "", http.StatusOK},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer "+tt.token)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			env.Router.ServeHTTP(rec, req)
			if rec.Code != tt.expect {
				t.Errorf("expected %d, got %d: %s", tt.expect, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestClasses_ValidationErrors(t *testing.T) {
	env := testEnv2(t)
	token := seedAdminToken(t, env)

	tests := []struct {
		name string
		body string
	}{
		{"missing fields", `{}`},
		{"invalid start_time", `{"class_type_id":1,"instructor_id":1,"start_time":"not-a-date","duration_minutes":60,"capacity":20}`},
		{"zero capacity", `{"class_type_id":1,"instructor_id":1,"start_time":"2026-03-01T10:00:00Z","duration_minutes":60,"capacity":0}`},
		{"zero duration", `{"class_type_id":1,"instructor_id":1,"start_time":"2026-03-01T10:00:00Z","duration_minutes":0,"capacity":20}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/api/classes", strings.NewReader(tt.body))
			req.Header.Set("Authorization", "Bearer "+token)
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			env.Router.ServeHTTP(rec, req)
			if rec.Code != http.StatusBadRequest {
				t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
			}
		})
	}
}
