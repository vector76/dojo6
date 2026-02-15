package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSPAServing(t *testing.T) {
	router, _ := testRouter(t)

	t.Run("serves index.html at root", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<div id=\"root\">") {
			t.Errorf("expected index.html content with root div, got: %s", body)
		}
	})

	t.Run("falls back to index.html for unknown frontend paths", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/members/123", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, "<div id=\"root\">") {
			t.Errorf("expected index.html fallback, got: %s", body)
		}
	})

	t.Run("returns 404 for unknown API routes", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/nonexistent", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Errorf("expected status %d, got %d", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("API health still works", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/api/health", nil)
		rec := httptest.NewRecorder()

		router.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, rec.Code)
		}
	})
}
