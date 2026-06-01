package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestStatusIcon(t *testing.T) {
	cases := map[string]string{
		"ok":      "✅ ok",
		"warn":    "⚠️ warn",
		"skipped": "⏭ skipped",
		"error":   "❌ error",
		"weird":   "weird",
	}
	for in, want := range cases {
		if got := statusIcon(in); got != want {
			t.Errorf("statusIcon(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestEscapeTable(t *testing.T) {
	got := escapeTable("a | b | c")
	want := `a \| b \| c`
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestTail(t *testing.T) {
	if got := tail("hello", 100); got != "hello" {
		t.Errorf("short string should not be tailed, got %q", got)
	}
	long := strings.Repeat("x", 5000)
	got := tail(long, 100)
	if !strings.HasPrefix(got, "[…truncated…]") {
		t.Error("expected truncation marker")
	}
	if !strings.HasSuffix(got, strings.Repeat("x", 100)) {
		t.Error("expected exactly 100 trailing chars")
	}
}

// TestExtractSupabaseRef finds a 20-char ref from .env.local or client.ts.
func TestExtractSupabaseRef(t *testing.T) {
	dir := t.TempDir()

	// No file → ""
	if got := extractSupabaseRef(dir); got != "" {
		t.Errorf("expected empty ref, got %q", got)
	}

	// .env.local with valid URL
	mustWrite(t, filepath.Join(dir, ".env.local"),
		"VITE_SUPABASE_URL=https://abcd1234efgh5678ijkl.supabase.co\nVITE_SUPABASE_PUBLISHABLE_KEY=xxx\n")
	if got := extractSupabaseRef(dir); got != "abcd1234efgh5678ijkl" {
		t.Errorf("expected ref from env, got %q", got)
	}

	// Falls back to client.ts when env missing
	dir2 := t.TempDir()
	mustWrite(t, filepath.Join(dir2, "src", "integrations", "supabase", "client.ts"),
		`const SUPABASE_URL = "https://zyxw9876vuts5432rqpo.supabase.co"`)
	if got := extractSupabaseRef(dir2); got != "zyxw9876vuts5432rqpo" {
		t.Errorf("expected ref from client.ts, got %q", got)
	}
}

func TestReadSourcePackageMeta(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "package.json"), `{"name":"my-app","version":"1.2.3"}`)
	name, version := readSourcePackageMeta(dir)
	if name != "my-app" || version != "1.2.3" {
		t.Errorf("got name=%q version=%q", name, version)
	}

	emptyDir := t.TempDir()
	n2, v2 := readSourcePackageMeta(emptyDir)
	if n2 != "" || v2 != "" {
		t.Errorf("expected empty for missing package.json, got %q %q", n2, v2)
	}
}

// TestWriteReport renders the report and verifies it contains expected sections.
func TestWriteReport(t *testing.T) {
	dir := t.TempDir()
	// Minimum needed for the Supabase checklist to render
	mustWrite(t, filepath.Join(dir, ".env.local"),
		"VITE_SUPABASE_URL=https://abcd1234efgh5678ijkl.supabase.co\n")

	result := &Result{
		PhaseLogs: []PhaseLog{
			{Phase: "01-preflight", Status: "ok", Message: "ok", Duration: "1ms"},
			{Phase: "08-zod-refactor", Status: "ok", Message: "11 changes", Duration: "2ms",
				Details: []string{"`src/x.ts`: foo (1x)"}},
		},
		BuildLog: "✓ built in 5.44s",
	}
	if err := writeReport(Options{OutputDir: dir, Validate: true, SourceDir: dir}, result, nil); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(filepath.Join(dir, "MIGRATION_REPORT.md"))
	out := string(data)

	mustContain := []string{
		"# Skip Migrator — Relatório",
		"## ✅ Resultado: SUCESSO",
		"## Fases executadas",
		"01-preflight",
		"## Transformações de código aplicadas",
		"`src/x.ts`",
		"## Cauda do build log",
		"built in 5.44s",
		"## Checklist pós-migração: Supabase",
		"abcd1234efgh5678ijkl",
		"## O que NÃO foi validado",
		"## Próximos passos recomendados",
	}
	for _, s := range mustContain {
		if !strings.Contains(out, s) {
			t.Errorf("report missing: %q", s)
		}
	}
}
