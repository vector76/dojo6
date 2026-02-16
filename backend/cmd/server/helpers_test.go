package main

import (
	"database/sql"
	"testing"
	"time"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/database"
	"dojo6/backend/internal/handlers"
	"dojo6/backend/internal/models"

	"github.com/go-chi/chi/v5"
)

const testJWTSecret = "test-secret"

var testJWTSvc = auth.NewJWTService(testJWTSecret, 24*time.Hour)

type testEnv struct {
	Router   chi.Router
	Auth     *auth.Handler
	Payment  *PaymentHandler
	DB       *sql.DB
}

func testRouter(t *testing.T) (chi.Router, *auth.Handler, *PaymentHandler) {
	t.Helper()
	env := testEnv2(t)
	return env.Router, env.Auth, env.Payment
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

	userRepo := models.NewUserRepository(db)

	h := &auth.Handler{
		Users:  userRepo,
		JWTSvc: testJWTSvc,
	}
	ch := &ClassHandlers{DB: db}
	ah := &AttendanceHandlers{Attendance: models.NewAttendanceRepository(db)}
	uh := handlers.NewUserHandler(models.NewUserRepository(db))
	ph := &PaymentHandler{
		Payments: models.NewPaymentRepository(db),
		Users:    userRepo,
	}
	return testEnv{
		Router:  newRouter(h, ch, ah, uh, ph, testJWTSvc),
		Auth:    h,
		Payment: ph,
		DB:      db,
	}
}
