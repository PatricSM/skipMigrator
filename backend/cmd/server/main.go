// Command server runs the skip-migrator HTTP API + background worker pool.
//
// It serves the REST endpoints under /api and concurrently runs N worker goroutines
// that pick up migration jobs from Postgres via LISTEN/NOTIFY.
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/PatricSM/skip-migrator/backend/internal/api"
	"github.com/PatricSM/skip-migrator/backend/internal/db"
	"github.com/PatricSM/skip-migrator/backend/internal/storage"
	"github.com/PatricSM/skip-migrator/backend/internal/worker"
)

func main() {
	_ = godotenv.Load()

	cfg := loadConfig()
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	// Postgres pool
	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("db pool: %v", err)
	}
	defer pool.Close()

	// Supabase Storage client
	store := storage.NewSupabaseStorage(cfg.SupabaseURL, cfg.SupabaseServiceKey)

	// Worker pool — N goroutines watching the same NOTIFY channel
	wp := worker.New(pool, store, cfg.WorkerConcurrency)
	go wp.Start(ctx)

	// HTTP server
	mux := api.NewRouter(api.Deps{
		Pool:               pool,
		Storage:            store,
		SupabaseURL:        cfg.SupabaseURL,
		SupabaseServiceKey: cfg.SupabaseServiceKey,
		SupabaseJWKURL:     cfg.SupabaseJWKURL,
		SupabaseAud:        cfg.SupabaseProjectRef,
		SuperAdminEmails:   cfg.SuperAdminEmails,
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           mux,
		ReadHeaderTimeout: 10 * time.Second,
		WriteTimeout:      120 * time.Second,
		IdleTimeout:       120 * time.Second,
	}

	go func() {
		log.Printf("skip-migrator API listening on :%s (workers=%d)", cfg.Port, cfg.WorkerConcurrency)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("http server: %v", err)
		}
	}()

	<-ctx.Done()
	log.Println("shutting down…")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = srv.Shutdown(shutdownCtx)
	wp.Stop()
	log.Println("bye")
}

type config struct {
	Port               string
	DatabaseURL        string
	SupabaseURL        string
	SupabaseServiceKey string
	SupabaseJWKURL     string
	SupabaseProjectRef string
	WorkerConcurrency  int
	SuperAdminEmails   []string
}

func loadConfig() config {
	c := config{
		Port:               envOr("PORT", "8000"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		SupabaseURL:        os.Getenv("SUPABASE_URL"),
		SupabaseServiceKey: os.Getenv("SUPABASE_SERVICE_ROLE_KEY"),
		SupabaseJWKURL:     os.Getenv("SUPABASE_JWK_URL"),
		SupabaseProjectRef: os.Getenv("SUPABASE_PROJECT_REF"),
	}
	if n, err := strconv.Atoi(envOr("WORKER_CONCURRENCY", "2")); err == nil {
		c.WorkerConcurrency = n
	} else {
		c.WorkerConcurrency = 2
	}
	if raw := strings.TrimSpace(os.Getenv("SUPER_ADMIN_EMAILS")); raw != "" {
		for _, e := range strings.Split(raw, ",") {
			if e = strings.TrimSpace(strings.ToLower(e)); e != "" {
				c.SuperAdminEmails = append(c.SuperAdminEmails, e)
			}
		}
	}
	if c.DatabaseURL == "" {
		log.Fatal("DATABASE_URL is required")
	}
	if c.SupabaseURL == "" || c.SupabaseServiceKey == "" {
		log.Fatal("SUPABASE_URL and SUPABASE_SERVICE_ROLE_KEY are required")
	}
	return c
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
