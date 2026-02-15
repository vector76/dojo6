package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"dojo6/backend/internal/auth"
	"dojo6/backend/internal/database"
	"dojo6/backend/internal/models"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func newRouter(authHandler *auth.Handler, jwtSvc *auth.JWTService) chi.Router {
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)

	r.Route("/api", func(api chi.Router) {
		api.Get("/health", healthHandler)

		// Public auth routes (no JWT required).
		api.Post("/auth/login", authHandler.Login)
		api.Get("/auth/setup-status", authHandler.SetupStatus)
		api.Post("/auth/setup", authHandler.Setup)

		// Protected auth routes (JWT required).
		api.Group(func(protected chi.Router) {
			protected.Use(auth.JWTMiddleware(jwtSvc))
			protected.Get("/auth/me", authHandler.Me)
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

	r := newRouter(authHandler, jwtSvc)

	log.Printf("Server starting on :%s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
