package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"
	"dojo6/backend/internal/repository"
)

// seedAttendanceEnv creates prerequisite data: admin, instructor, user, class type, and class.
// Returns tokens and IDs needed for attendance tests.
type attendanceEnvResult struct {
	AdminToken     string
	InstructorToken string
	UserToken      string
	UserID         int64
	InstructorID   int64
	ClassID        int64
}

func seedAttendanceEnv(t *testing.T, env testEnv) attendanceEnvResult {
	t.Helper()
	hash, _ := auth.HashPassword("pass")

	admin := models.User{Name: "Admin", Email: "admin@att.com", Role: "admin", PasswordHash: hash}
	if err := env.Auth.Users.Create(&admin); err != nil {
		t.Fatal(err)
	}
	adminTok, _ := auth.GenerateToken(testJWTSecret, admin.ID, admin.Email, admin.Role)

	instructor := models.User{Name: "Instructor", Email: "inst@att.com", Role: "instructor", PasswordHash: hash}
	if err := env.Auth.Users.Create(&instructor); err != nil {
		t.Fatal(err)
	}
	instTok, _ := auth.GenerateToken(testJWTSecret, instructor.ID, instructor.Email, instructor.Role)

	user := models.User{Name: "Student", Email: "student@att.com", Role: "user", PasswordHash: hash}
	if err := env.Auth.Users.Create(&user); err != nil {
		t.Fatal(err)
	}
	userTok, _ := auth.GenerateToken(testJWTSecret, user.ID, user.Email, user.Role)

	ct, err := repository.CreateClassType(env.DB, "Yoga", "")
	if err != nil {
		t.Fatal(err)
	}
	class, err := repository.CreateClass(env.DB, ct.ID, instructor.ID, mustParseTime("2026-03-01T10:00:00Z"), 60, 20)
	if err != nil {
		t.Fatal(err)
	}

	return attendanceEnvResult{
		AdminToken:      adminTok,
		InstructorToken: instTok,
		UserToken:       userTok,
		UserID:          user.ID,
		InstructorID:    instructor.ID,
		ClassID:         class.ID,
	}
}

func mustParseTime(s string) time.Time {
	t, err := time.Parse(time.RFC3339, s)
	if err != nil {
		panic(err)
	}
	return t
}

// RecordAttendance tests

func TestRecordAttendance_Success(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	body := fmt.Sprintf(`{"user_id":%d}`, seed.UserID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seed.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var result attendanceJSON
	json.NewDecoder(rec.Body).Decode(&result)
	if result.ClassID != seed.ClassID {
		t.Errorf("expected class_id %d, got %d", seed.ClassID, result.ClassID)
	}
	if result.UserID != seed.UserID {
		t.Errorf("expected user_id %d, got %d", seed.UserID, result.UserID)
	}
	if result.CheckedInAt == "" {
		t.Error("expected checked_in_at to be set")
	}
}

func TestRecordAttendance_InstructorCanRecord(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	body := fmt.Sprintf(`{"user_id":%d}`, seed.UserID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seed.InstructorToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecordAttendance_UserForbidden(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	body := fmt.Sprintf(`{"user_id":%d}`, seed.UserID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seed.UserToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecordAttendance_Unauthenticated(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	body := fmt.Sprintf(`{"user_id":%d}`, seed.UserID)
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}

func TestRecordAttendance_MissingUserID(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(`{}`))
	req.Header.Set("Authorization", "Bearer "+seed.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestRecordAttendance_DuplicateCheckIn(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	body := fmt.Sprintf(`{"user_id":%d}`, seed.UserID)

	// First check-in succeeds.
	req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seed.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusCreated {
		t.Fatalf("first check-in: expected 201, got %d", rec.Code)
	}

	// Second check-in fails (unique constraint).
	req = httptest.NewRequest(http.MethodPost, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), strings.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+seed.AdminToken)
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)
	if rec.Code != http.StatusInternalServerError {
		t.Errorf("duplicate check-in: expected 500, got %d", rec.Code)
	}
}

// ListClassAttendance tests

func TestListClassAttendance_Success(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	// Record attendance first.
	a := &models.Attendance{ClassID: seed.ClassID, UserID: seed.UserID}
	if err := models.NewAttendanceRepository(env.DB).Create(a); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), nil)
	req.Header.Set("Authorization", "Bearer "+seed.AdminToken)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result []attendanceJSON
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("expected 1 record, got %d", len(result))
	}
}

func TestListClassAttendance_InstructorAllowed(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), nil)
	req.Header.Set("Authorization", "Bearer "+seed.InstructorToken)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestListClassAttendance_UserForbidden(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/classes/%d/attendance", seed.ClassID), nil)
	req.Header.Set("Authorization", "Bearer "+seed.UserToken)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d", rec.Code)
	}
}

// ListUserAttendance tests

func TestListUserAttendance_SelfAccess(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	// Record attendance.
	a := &models.Attendance{ClassID: seed.ClassID, UserID: seed.UserID}
	if err := models.NewAttendanceRepository(env.DB).Create(a); err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/attendance", seed.UserID), nil)
	req.Header.Set("Authorization", "Bearer "+seed.UserToken)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var result []attendanceJSON
	json.NewDecoder(rec.Body).Decode(&result)
	if len(result) != 1 {
		t.Errorf("expected 1 record, got %d", len(result))
	}
}

func TestListUserAttendance_AdminAccess(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/attendance", seed.UserID), nil)
	req.Header.Set("Authorization", "Bearer "+seed.AdminToken)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestListUserAttendance_InstructorAccess(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/attendance", seed.UserID), nil)
	req.Header.Set("Authorization", "Bearer "+seed.InstructorToken)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rec.Code)
	}
}

func TestListUserAttendance_OtherUserForbidden(t *testing.T) {
	env := testEnv2(t)
	seed := seedAttendanceEnv(t, env)

	// Create a second user.
	hash, _ := auth.HashPassword("pass")
	user2 := models.User{Name: "Other", Email: "other@att.com", Role: "user", PasswordHash: hash}
	if err := env.Auth.Users.Create(&user2); err != nil {
		t.Fatal(err)
	}
	user2Tok, _ := auth.GenerateToken(testJWTSecret, user2.ID, user2.Email, user2.Role)

	// user2 tries to view seed.UserID's attendance.
	req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/users/%d/attendance", seed.UserID), nil)
	req.Header.Set("Authorization", "Bearer "+user2Tok)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("expected 403, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestListUserAttendance_Unauthenticated(t *testing.T) {
	env := testEnv2(t)

	req := httptest.NewRequest(http.MethodGet, "/api/users/1/attendance", nil)
	rec := httptest.NewRecorder()
	env.Router.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("expected 401, got %d", rec.Code)
	}
}
