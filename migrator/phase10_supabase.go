package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// phase10SupabaseVersioning copies source's migrations + edge functions + config.toml
// into output's supabase/ dir, replacing the baseline's default init_schema.
func phase10SupabaseVersioning(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "10-supabase-versioning"}

	srcSupa := filepath.Join(opts.SourceDir, "supabase")
	if !fileExists(srcSupa) {
		log.Status = "skipped"
		log.Message = "source has no supabase/ dir"
		log.Duration = time.Since(start).String()
		return log, nil
	}

	dstSupa := filepath.Join(opts.OutputDir, "supabase")

	// Remove baseline's default init_schema migration; source supplies its own
	dstMigrations := filepath.Join(dstSupa, "migrations")
	if entries, err := os.ReadDir(dstMigrations); err == nil {
		for _, e := range entries {
			_ = os.Remove(filepath.Join(dstMigrations, e.Name()))
		}
	}

	// Copy source migrations
	srcMig := filepath.Join(srcSupa, "migrations")
	if fileExists(srcMig) {
		if err := os.MkdirAll(dstMigrations, 0o755); err != nil {
			return log, err
		}
		if err := copyDir(srcMig, dstMigrations, nil); err != nil {
			return log, fmt.Errorf("copying migrations: %w", err)
		}
	}

	// Copy edge functions
	srcFn := filepath.Join(srcSupa, "functions")
	if fileExists(srcFn) {
		dstFn := filepath.Join(dstSupa, "functions")
		if err := os.MkdirAll(dstFn, 0o755); err != nil {
			return log, err
		}
		if err := copyDir(srcFn, dstFn, nil); err != nil {
			return log, fmt.Errorf("copying edge functions: %w", err)
		}
	}

	// Copy config.toml
	srcCfg := filepath.Join(srcSupa, "config.toml")
	if fileExists(srcCfg) {
		_ = copyFile(srcCfg, filepath.Join(dstSupa, "config.toml"))
	}

	log.Status = "ok"
	log.Message = "supabase/ (migrations, functions, config.toml) versioned from source"
	log.Duration = time.Since(start).String()
	return log, nil
}
