package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"dojo6/backend/internal/repository"

	"github.com/go-chi/chi/v5"
)

// ClassHandlers holds dependencies for class-related HTTP handlers.
type ClassHandlers struct {
	DB *sql.DB
}

// JSON response types

type classTypeJSON struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
}

type classJSON struct {
	ID              int64  `json:"id"`
	ClassTypeID     int64  `json:"class_type_id"`
	InstructorID    int64  `json:"instructor_id"`
	StartTime       string `json:"start_time"`
	DurationMinutes int    `json:"duration_minutes"`
	Capacity        int    `json:"capacity"`
}

func toClassTypeJSON(ct repository.ClassType) classTypeJSON {
	return classTypeJSON{
		ID:          ct.ID,
		Name:        ct.Name,
		Description: ct.Description,
	}
}

func toClassJSON(c repository.Class) classJSON {
	return classJSON{
		ID:              c.ID,
		ClassTypeID:     c.ClassTypeID,
		InstructorID:    c.InstructorID,
		StartTime:       c.StartTime.Format(time.RFC3339),
		DurationMinutes: c.DurationMinutes,
		Capacity:        c.Capacity,
	}
}

func classWriteJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func classWriteError(w http.ResponseWriter, status int, msg string) {
	classWriteJSON(w, status, map[string]string{"error": msg})
}

// --- Class Type handlers ---

// ListClassTypes handles GET /api/class-types.
func (h *ClassHandlers) ListClassTypes(w http.ResponseWriter, r *http.Request) {
	types, err := repository.ListClassTypes(h.DB)
	if err != nil {
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	result := make([]classTypeJSON, len(types))
	for i, ct := range types {
		result[i] = toClassTypeJSON(ct)
	}
	classWriteJSON(w, http.StatusOK, result)
}

// GetClassType handles GET /api/class-types/{id}.
func (h *ClassHandlers) GetClassType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	ct, err := repository.GetClassType(h.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			classWriteError(w, http.StatusNotFound, "class type not found")
			return
		}
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	classWriteJSON(w, http.StatusOK, toClassTypeJSON(ct))
}

// CreateClassType handles POST /api/class-types.
func (h *ClassHandlers) CreateClassType(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		classWriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	ct, err := repository.CreateClassType(h.DB, req.Name, req.Description)
	if err != nil {
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	classWriteJSON(w, http.StatusCreated, toClassTypeJSON(ct))
}

// UpdateClassType handles PUT /api/class-types/{id}.
func (h *ClassHandlers) UpdateClassType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Name == "" {
		classWriteError(w, http.StatusBadRequest, "name is required")
		return
	}
	ct, err := repository.UpdateClassType(h.DB, id, req.Name, req.Description)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			classWriteError(w, http.StatusNotFound, "class type not found")
			return
		}
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	classWriteJSON(w, http.StatusOK, toClassTypeJSON(ct))
}

// DeleteClassType handles DELETE /api/class-types/{id}.
func (h *ClassHandlers) DeleteClassType(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := repository.DeleteClassType(h.DB, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			classWriteError(w, http.StatusNotFound, "class type not found")
			return
		}
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

// --- Class handlers ---

// ListClasses handles GET /api/classes.
func (h *ClassHandlers) ListClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := repository.ListClasses(h.DB)
	if err != nil {
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	result := make([]classJSON, len(classes))
	for i, c := range classes {
		result[i] = toClassJSON(c)
	}
	classWriteJSON(w, http.StatusOK, result)
}

// GetClass handles GET /api/classes/{id}.
func (h *ClassHandlers) GetClass(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	c, err := repository.GetClass(h.DB, id)
	if err != nil {
		if strings.Contains(err.Error(), "no rows") {
			classWriteError(w, http.StatusNotFound, "class not found")
			return
		}
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	classWriteJSON(w, http.StatusOK, toClassJSON(c))
}

// CreateClass handles POST /api/classes.
func (h *ClassHandlers) CreateClass(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ClassTypeID     int64  `json:"class_type_id"`
		InstructorID    int64  `json:"instructor_id"`
		StartTime       string `json:"start_time"`
		DurationMinutes int    `json:"duration_minutes"`
		Capacity        int    `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ClassTypeID == 0 || req.InstructorID == 0 || req.StartTime == "" || req.DurationMinutes <= 0 || req.Capacity <= 0 {
		classWriteError(w, http.StatusBadRequest, "class_type_id, instructor_id, start_time, duration_minutes, and capacity are required")
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "start_time must be in RFC3339 format")
		return
	}
	c, err := repository.CreateClass(h.DB, req.ClassTypeID, req.InstructorID, startTime, req.DurationMinutes, req.Capacity)
	if err != nil {
		classWriteError(w, http.StatusInternalServerError, "failed to create class")
		return
	}
	classWriteJSON(w, http.StatusCreated, toClassJSON(c))
}

// UpdateClass handles PUT /api/classes/{id}.
func (h *ClassHandlers) UpdateClass(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	var req struct {
		ClassTypeID     int64  `json:"class_type_id"`
		InstructorID    int64  `json:"instructor_id"`
		StartTime       string `json:"start_time"`
		DurationMinutes int    `json:"duration_minutes"`
		Capacity        int    `json:"capacity"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ClassTypeID == 0 || req.InstructorID == 0 || req.StartTime == "" || req.DurationMinutes <= 0 || req.Capacity <= 0 {
		classWriteError(w, http.StatusBadRequest, "class_type_id, instructor_id, start_time, duration_minutes, and capacity are required")
		return
	}
	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "start_time must be in RFC3339 format")
		return
	}
	c, err := repository.UpdateClass(h.DB, id, req.ClassTypeID, req.InstructorID, startTime, req.DurationMinutes, req.Capacity)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			classWriteError(w, http.StatusNotFound, "class not found")
			return
		}
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	classWriteJSON(w, http.StatusOK, toClassJSON(c))
}

// DeleteClass handles DELETE /api/classes/{id}.
func (h *ClassHandlers) DeleteClass(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	if err != nil {
		classWriteError(w, http.StatusBadRequest, "invalid id")
		return
	}
	if err := repository.DeleteClass(h.DB, id); err != nil {
		if strings.Contains(err.Error(), "not found") {
			classWriteError(w, http.StatusNotFound, "class not found")
			return
		}
		classWriteError(w, http.StatusInternalServerError, "internal error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
