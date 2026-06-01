package migrator

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

// phase01Preflight inspects the source Lovable project, extracts Supabase credentials,
// and writes a .env.local in the output dir.
func phase01Preflight(opts Options) (PhaseLog, error) {
	start := time.Now()
	log := PhaseLog{Phase: "01-preflight"}

	if !fileExists(filepath.Join(opts.SourceDir, "package.json")) {
		return log, fmt.Errorf("source %s does not look like a Node project (no package.json)", opts.SourceDir)
	}

	url, anonKey := opts.SupabaseURL, opts.SupabaseAnonKey
	if url == "" || anonKey == "" {
		extractedURL, extractedKey, err := extractSupabaseCreds(opts.SourceDir)
		if err == nil {
			if url == "" {
				url = extractedURL
			}
			if anonKey == "" {
				anonKey = extractedKey
			}
		}
	}

	envBody := fmt.Sprintf("VITE_SUPABASE_URL=%s\nVITE_SUPABASE_PUBLISHABLE_KEY=%s\n", url, anonKey)
	if err := os.WriteFile(filepath.Join(opts.OutputDir, ".env.local"), []byte(envBody), 0o600); err != nil {
		return log, fmt.Errorf("writing .env.local: %w", err)
	}

	log.Status = "ok"
	log.Message = fmt.Sprintf(".env.local written (url=%s, key=%s)", maskURL(url), maskKey(anonKey))
	log.Duration = time.Since(start).String()
	return log, nil
}

var (
	supabaseURLRe = regexp.MustCompile(`(?:SUPABASE_URL|supabaseUrl)\s*[:=]\s*["']([^"']+)["']`)
	supabaseKeyRe = regexp.MustCompile(`(?:SUPABASE_PUBLISHABLE_KEY|SUPABASE_ANON_KEY|supabaseKey|supabaseAnonKey)\s*[:=]\s*["']([^"']+)["']`)
)

// extractSupabaseCreds looks for hardcoded Supabase URL/anon key in common locations.
// Lovable embeds these in src/integrations/supabase/client.ts.
func extractSupabaseCreds(srcDir string) (url, anonKey string, err error) {
	candidates := []string{
		"src/integrations/supabase/client.ts",
		"src/lib/supabase/client.ts",
		"src/lib/supabase.ts",
	}
	for _, p := range candidates {
		full := filepath.Join(srcDir, p)
		data, err := os.ReadFile(full)
		if err != nil {
			continue
		}
		content := string(data)
		if m := supabaseURLRe.FindStringSubmatch(content); len(m) > 1 && url == "" {
			url = strings.TrimSpace(m[1])
		}
		if m := supabaseKeyRe.FindStringSubmatch(content); len(m) > 1 && anonKey == "" {
			anonKey = strings.TrimSpace(m[1])
		}
		if url != "" && anonKey != "" {
			return url, anonKey, nil
		}
	}
	if url == "" || anonKey == "" {
		return url, anonKey, fmt.Errorf("could not extract Supabase credentials from source")
	}
	return url, anonKey, nil
}

func maskURL(u string) string {
	if u == "" {
		return "<empty>"
	}
	return u
}

func maskKey(k string) string {
	if k == "" {
		return "<empty>"
	}
	if len(k) <= 12 {
		return "***"
	}
	return k[:6] + "..." + k[len(k)-4:]
}
