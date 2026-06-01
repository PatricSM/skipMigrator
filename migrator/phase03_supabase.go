package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const supabaseClientTemplate = `// AVOID UPDATING THIS FILE DIRECTLY. It is automatically generated.
import { createClient } from '@supabase/supabase-js'
import type { Database } from './types'

const SUPABASE_URL = import.meta.env.VITE_SUPABASE_URL as string
const SUPABASE_PUBLISHABLE_KEY = import.meta.env.VITE_SUPABASE_PUBLISHABLE_KEY as string

// Import the supabase client like this:
// import { supabase } from "@/integrations/supabase/client";

export const supabase = createClient<Database>(SUPABASE_URL, SUPABASE_PUBLISHABLE_KEY, {
  auth: {
    storage: localStorage,
    persistSession: true,
    autoRefreshToken: true,
  },
})
`

// phase03SupabaseClient creates src/integrations/supabase/{client.ts,types.ts} in OutputDir.
// types.ts is copied from source if present.
func phase03SupabaseClient(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "03-supabase-client"}

	targetDir := filepath.Join(opts.OutputDir, "src", "integrations", "supabase")
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return log, err
	}

	// Write env-based client.ts
	if err := os.WriteFile(filepath.Join(targetDir, "client.ts"), []byte(supabaseClientTemplate), 0o644); err != nil {
		return log, err
	}

	// Copy types.ts from source if present
	srcCandidates := []string{
		"src/integrations/supabase/types.ts",
		"src/lib/supabase/types.ts",
	}
	copied := false
	for _, p := range srcCandidates {
		full := filepath.Join(opts.SourceDir, p)
		if fileExists(full) {
			if err := copyFile(full, filepath.Join(targetDir, "types.ts")); err != nil {
				return log, fmt.Errorf("copying types.ts: %w", err)
			}
			copied = true
			break
		}
	}

	// Remove Skip baseline's lib/supabase if it exists (path collision avoidance)
	_ = os.RemoveAll(filepath.Join(opts.OutputDir, "src", "lib", "supabase"))

	status := "ok"
	msg := "client.ts written; types.ts copied"
	if !copied {
		status = "warn"
		msg = "client.ts written; types.ts NOT found in source — generate via `supabase gen types`"
	}
	log.Status = status
	log.Message = msg
	log.Duration = time.Since(start).String()
	return log, nil
}
