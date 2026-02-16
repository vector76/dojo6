package main

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

// AttendanceHandlers holds dependencies for attendance-related HTTP handlers.
type AttendanceHandlers struct {
	Attendance *models.AttendanceRepository
}

type attendanceJSON struct {
	ID          int64  `json:"id"`
	ClassID     int64  `json:"class_id"`
	UserID      int64  `json:"user_id"`
	CheckedInAt string `json:"checked_in_at"`
}

func toAttendanceJSON(a models.Attendance) attendanceJSON {
	return attendanceJSON{
		ID:          a.ID,
		ClassID:     a.ClassID,
		UserID:      a.UserID,
		CheckedInAt: a.CheckedInAt.Format(time.RFC3339),
	}
}

// RecordAttendance handles POST /api/classes/{id}/attendance.
// Only admin and instructor roles may record attendance.
func (h *AttendanceHandlers) RecordAttendance(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid class id")
		return
	}

	var req struct {
		UserID int64 `json:"user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.UserID == 0 {
		classWriteError(w, http.StatusBadRequest, "user_id is required")
		return
	}

	a := &models.Attendance{
		ClassID: classID,
		UserID:  req.UserID,
	}
	if err := h.Attendance.Create(a); err != nil {
		classWriteError(w, http.StatusInternalServerError, "failed to record attendance")
		return
	}

	classWriteJSON(w, http.StatusCreated, toAttendanceJSON(*a))
}

// ListClassAttendance handles GET /api/classes/{id}/attendance.
// Only admin and instructor roles may list class attendance.
func (h *AttendanceHandlers) ListClassAttendance(w http.ResponseWriter, r *http.Request) {
	classID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid class id")
		return
	}

	records, err := h.Attendance.ListByClass(classID)
	if err != nil {
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]attendanceJSON, len(records))
	for i, a := range records {
		result[i] = toAttendanceJSON(a)
	}
	classWriteJSON(w, http.StatusOK, result)
}

// ListUserAttendance handles GET /api/users/{id}/attendance.
// Accessible by the user themselves, or by admin/instructor roles.
func (h *AttendanceHandlers) ListUserAttendance(w http.ResponseWriter, r *http.Request) {
	userID, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid user id")
		return
	}

	claims := auth.GetClaims(r)
	if claims == nil {
		classWriteError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	// Allow self-access or admin/instructor access.
	if claims.UserID != userID && claims.Role != "admin" && claims.Role != "instructor" {
		classWriteError(w, http.StatusForbidden, "forbidden")
		return
	}

	records, err := h.Attendance.ListByUser(userID)
	if err != nil {
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}

	result := make([]attendanceJSON, len(records))
	for i, a := range records {
		result[i] = toAttendanceJSON(a)
	}
	classWriteJSON(w, http.StatusOK, result)
}
