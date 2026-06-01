// Package db owns the Postgres connection pool and CRUD over the `migrations` table.
package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool opens a pgxpool against databaseURL with sensible defaults.
func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parsing DATABASE_URL: %w", err)
	}
	cfg.MaxConns = 20
	cfg.MinConns = 2
	cfg.MaxConnLifetime = 30 * time.Minute
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("pinging db: %w", err)
	}
	return pool, nil
}

// Status values for the `migrations.status` column.
const (
	StatusQueued  = "queued"
	StatusRunning = "running"
	StatusSuccess = "success"
	StatusFailed  = "failed"
)

// Migration mirrors the `migrations` row.
type Migration struct {
	ID             string
	UserID         string
	Status         string
	CreatedAt      time.Time
	UpdatedAt      time.Time
	SourceZipPath  string
	OutputZipPath  *string
	BuildLog       *string
	ErrorMessage   *string
	PixelPerfect   bool
	Validate       bool
	SupabaseStrategy string
}

// Insert creates a new migration row in `queued` status and returns the ID.
func Insert(ctx context.Context, pool *pgxpool.Pool, m Migration) (string, error) {
	const q = `
INSERT INTO migrations (user_id, status, source_zip_path, pixel_perfect, validate, supabase_strategy)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING id::text
`
	var id string
	err := pool.QueryRow(ctx, q, m.UserID, StatusQueued, m.SourceZipPath, m.PixelPerfect, m.Validate, m.SupabaseStrategy).Scan(&id)
	if err != nil {
		return "", err
	}
	// Notify worker pool that a new job is available.
	_, _ = pool.Exec(ctx, "SELECT pg_notify('migrations_queue', $1)", id)
	return id, nil
}

// FetchAndLockNext picks the oldest queued migration and marks it running atomically.
// Returns (Migration, true) on success, (zero, false) if nothing is available.
func FetchAndLockNext(ctx context.Context, pool *pgxpool.Pool) (Migration, bool, error) {
	const q = `
UPDATE migrations
SET status = $1, updated_at = NOW()
WHERE id = (
  SELECT id FROM migrations
  WHERE status = $2
  ORDER BY created_at
  FOR UPDATE SKIP LOCKED
  LIMIT 1
)
RETURNING id::text, user_id::text, status, created_at, updated_at, source_zip_path, pixel_perfect, validate, supabase_strategy
`
	var m Migration
	err := pool.QueryRow(ctx, q, StatusRunning, StatusQueued).Scan(
		&m.ID, &m.UserID, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		&m.SourceZipPath, &m.PixelPerfect, &m.Validate, &m.SupabaseStrategy,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return Migration{}, false, nil
	}
	if err != nil {
		return Migration{}, false, err
	}
	return m, true, nil
}

// MarkSuccess finalizes a migration: status=success, attaches output path and build log.
func MarkSuccess(ctx context.Context, pool *pgxpool.Pool, id, outputPath, buildLog string) error {
	_, err := pool.Exec(ctx, `
UPDATE migrations
SET status = $1, output_zip_path = $2, build_log = $3, updated_at = NOW()
WHERE id = $4
`, StatusSuccess, outputPath, buildLog, id)
	return err
}

// MarkFailed records an error and the partial build log (if any).
func MarkFailed(ctx context.Context, pool *pgxpool.Pool, id, errMsg, buildLog string) error {
	_, err := pool.Exec(ctx, `
UPDATE migrations
SET status = $1, error_message = $2, build_log = $3, updated_at = NOW()
WHERE id = $4
`, StatusFailed, errMsg, buildLog, id)
	return err
}

// GetByID fetches a single migration row.
func GetByID(ctx context.Context, pool *pgxpool.Pool, id, userID string) (Migration, error) {
	const q = `
SELECT id::text, user_id::text, status, created_at, updated_at, source_zip_path, output_zip_path, build_log, error_message, pixel_perfect, validate, supabase_strategy
FROM migrations
WHERE id = $1 AND user_id = $2
`
	var m Migration
	err := pool.QueryRow(ctx, q, id, userID).Scan(
		&m.ID, &m.UserID, &m.Status, &m.CreatedAt, &m.UpdatedAt,
		&m.SourceZipPath, &m.OutputZipPath, &m.BuildLog, &m.ErrorMessage,
		&m.PixelPerfect, &m.Validate, &m.SupabaseStrategy,
	)
	return m, err
}

// ListByUser returns the user's migrations newest-first (up to limit).
func ListByUser(ctx context.Context, pool *pgxpool.Pool, userID string, limit int) ([]Migration, error) {
	const q = `
SELECT id::text, user_id::text, status, created_at, updated_at, source_zip_path, output_zip_path, build_log, error_message, pixel_perfect, validate, supabase_strategy
FROM migrations
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2
`
	rows, err := pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Migration
	for rows.Next() {
		var m Migration
		if err := rows.Scan(&m.ID, &m.UserID, &m.Status, &m.CreatedAt, &m.UpdatedAt,
			&m.SourceZipPath, &m.OutputZipPath, &m.BuildLog, &m.ErrorMessage,
			&m.PixelPerfect, &m.Validate, &m.SupabaseStrategy); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}
