package main

import (
	"database/sql"
	"testing"
	"time"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/database"
	"dojo6/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

const testJWTSecret = "test-secret"

var testJWTSvc = auth.NewJWTService(testJWTSecret, 24*time.Hour)

type testEnv struct {
	Router chi.Router
	Auth   *auth.Handler
	DB     *sql.DB
}

func testRouter(t *testing.T) (chi.Router, *auth.Handler) {
	t.Helper()
	env := testEnv2(t)
	return env.Router, env.Auth
}

func testEnv2(t *testing.T) testEnv {
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
	ch := &ClassHandlers{DB: db}
	return testEnv{
		Router: newRouter(h, ch, testJWTSvc),
		Auth:   h,
		DB:     db,
	}
}
