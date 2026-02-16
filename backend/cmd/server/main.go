package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/database"
	"dojo6/backend/internal/handlers"
	"dojo6/backend/internal/models"
)

func newRouter(authHandler *auth.Handler, classHandler *ClassHandlers, attendanceHandler *AttendanceHandlers, userHandler *handlers.UserHandler, jwtSvc *auth.JWTService) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", healthHandler)

		// Public auth routes (no JWT required).
		api.Post("/auth/login", authHandler.Login)
		api.Get("/auth/setup-status", authHandler.SetupStatus)
		api.Post("/auth/setup", authHandler.Setup)

		// Protected routes (JWT required).
		api.Group(func(protected chi.Router) {
			protected.Use(auth.JWTMiddleware(jwtSvc))
			protected.Get("/auth/me", authHandler.Me)

			// Class types: list/get for all authenticated, mutate for admin only.
			protected.Get("/class-types", classHandler.ListClassTypes)
			protected.Get("/class-types/{id}", classHandler.GetClassType)
			protected.Group(func(admin chi.Router) {
				admin.Use(auth.RequireRole("admin"))
				admin.Post("/class-types", classHandler.CreateClassType)
				admin.Put("/class-types/{id}", classHandler.UpdateClassType)
				admin.Delete("/class-types/{id}", classHandler.DeleteClassType)
			})

			// Classes: list/get for all authenticated, mutate for admin only.
			protected.Get("/classes", classHandler.ListClasses)
			protected.Get("/classes/{id}", classHandler.GetClass)
			protected.Group(func(admin chi.Router) {
				admin.Use(auth.RequireRole("admin"))
				admin.Post("/classes", classHandler.CreateClass)
				admin.Put("/classes/{id}", classHandler.UpdateClass)
				admin.Delete("/classes/{id}", classHandler.DeleteClass)
			})

			// Attendance: class attendance for admin/instructor, user history for self/admin/instructor.
			protected.Group(func(staff chi.Router) {
				staff.Use(auth.RequireRole("admin", "instructor"))
				staff.Post("/classes/{id}/attendance", attendanceHandler.RecordAttendance)
				staff.Get("/classes/{id}/attendance", attendanceHandler.ListClassAttendance)
			})
			// User attendance: permission check is done in the handler (self or admin/instructor).
			protected.Get("/users/{id}/attendance", attendanceHandler.ListUserAttendance)

			// Users: CRUD with role-based access.
			protected.Get("/users", userHandler.List)
			protected.Post("/users", userHandler.Create)
			protected.Get("/users/{id}", userHandler.GetByID)
			protected.Put("/users/{id}", userHandler.Update)
			protected.Delete("/users/{id}", userHandler.Delete)
			protected.Put("/users/{id}/role", userHandler.ChangeRole)
			protected.Put("/users/{id}/password", userHandler.ChangePassword)
		})
	})

	// Serve embedded frontend for all non-API routes.
	r.NotFound(spaHandler().ServeHTTP)

	return r
}

func healthHandler(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func main() {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	dbPath := os.Getenv("DB_PATH")
	if dbPath == "" {
		dbPath = "dojo6.db"
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		// Generate a random secret for development.
		b := make([]byte, 32)
		if _, err := rand.Read(b); err != nil {
			log.Fatalf("Failed to generate JWT secret: %v", err)
		}
		jwtSecret = hex.EncodeToString(b)
		log.Println("No JWT_SECRET set; generated a random secret (tokens will not survive restarts)")
	}

	db, err := database.Open(dbPath)
	if err != nil {
		log.Fatalf("Failed to open database: %v", err)
	}
	defer db.Close()

	if err := database.Migrate(db, database.MigrationsFS); err != nil {
		log.Fatalf("Failed to run migrations: %v", err)
	}

	jwtSvc := auth.NewJWTService(jwtSecret, 24*time.Hour)

	authHandler := &auth.Handler{
		Users:  models.NewUserRepository(db),
		JWTSvc: jwtSvc,
	}

	classHandler := &ClassHandlers{DB: db}
	attendanceHandler := &AttendanceHandlers{Attendance: models.NewAttendanceRepository(db)}
	userHandler := handlers.NewUserHandler(models.NewUserRepository(db))

	r := newRouter(authHandler, classHandler, attendanceHandler, userHandler, jwtSvc)

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
