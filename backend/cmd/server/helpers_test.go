package main

import (
	"testing"
	"time"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/database"
	"dojo6/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

const testJWTSecret = "test-secret"

var testJWTSvc = auth.NewJWTService(testJWTSecret, 24*time.Hour)

func testRouter(t *testing.T) (chi.Router, *auth.Handler) {
	t.Helper()
	db, err := database.Open(":memory:")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })

	if err := database.Migrate(db, database.MigrationsFS); err != nil {
		t.Fatal(err)
	}

	h := &auth.Handler{
		Users:  models.NewUserRepository(db),
		JWTSvc: testJWTSvc,
	}
	return newRouter(h, testJWTSvc), h
}
