// Package api defines the HTTP routes for skip-migrator.
package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/PatricSM/skip-migrator/backend/internal/storage"
)

// Deps holds the dependencies handlers need.
type Deps struct {
	Pool               *pgxpool.Pool
	Storage            *storage.SupabaseStorage
	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseJWKURL     string // e.g. https://<ref>.supabase.co/auth/v1/.well-known/jwks.json
	SupabaseAud        string // project ref (audience claim)
	SuperAdminEmails   []string
}

// NewRouter builds the chi router with all middleware + routes wired.
func NewRouter(deps Deps) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(5 * 60 * 1e9))

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"*"}, // tighten in production via env
		AllowedMethods:   []string{"GET", "POST", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		ExposedHeaders:   []string{"Content-Length", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	h := &handlers{deps: deps}

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("ok"))
	})

	r.Route("/api", func(r chi.Router) {
		r.Use(authMiddleware(deps))
		r.Get("/me", h.me)
		r.Post("/migrations", h.createMigration)
		r.Get("/migrations", h.listMigrations)
		r.Get("/migrations/{id}", h.getMigration)
		r.Get("/migrations/{id}/download", h.downloadOutput)

		// Super-admin only
		r.Group(func(r chi.Router) {
			r.Use(requireSuperAdmin)
			r.Get("/admin/users", h.adminListUsers)
			r.Post("/admin/users", h.adminCreateUser)
			r.Delete("/admin/users/{id}", h.adminDeleteUser)
		})
	})

	return r
}
