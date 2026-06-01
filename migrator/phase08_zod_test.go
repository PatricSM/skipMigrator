package migrator

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestZodReplacements verifies each Zod 3→4 substitution pattern in isolation,
// then verifies they're idempotent (applying twice doesn't double-mangle).
func TestZodReplacements(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			"required_error",
			`z.number({ required_error: 'Required' })`,
			`z.number({ error: 'Required' })`,
		},
		{
			"invalid_type_error",
			`z.string({ invalid_type_error: 'wrong type' })`,
			`z.string({ error: 'wrong type' })`,
		},
		{
			".errors[0]",
			`const msg = err.errors[0].message`,
			`const msg = err.issues[0].message`,
		},
		{
			"?.errors?.[0]",
			`const msg = err?.errors?.[0]?.message`,
			`const msg = err?.issues?.[0]?.message`,
		},
		{
			"errorMap",
			`z.string({ errorMap: () => ({ message: 'x' }) })`,
			`z.string({ error: () => ({ message: 'x' }) })`,
		},
		{
			"no change when zod patterns absent",
			`const foo = bar.baz`,
			`const foo = bar.baz`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tc.input
			for _, r := range zodReplacements {
				got = r.pattern.ReplaceAllString(got, r.replace)
			}
			if got != tc.want {
				t.Errorf("got %q, want %q", got, tc.want)
			}

			// Idempotency: second pass should be a no-op.
			second := got
			for _, r := range zodReplacements {
				second = r.pattern.ReplaceAllString(second, r.replace)
			}
			if second != got {
				t.Errorf("not idempotent: 1st=%q 2nd=%q", got, second)
			}
		})
	}
}

// TestPhase08ZodRefactor runs the phase end-to-end against a fake project tree.
// Exercises the file-walker, gate (skip files without zod imports), and details capture.
func TestPhase08ZodRefactor(t *testing.T) {
	dir := t.TempDir()
	mustWrite(t, filepath.Join(dir, "src", "lib", "validations.ts"),
		`import { z } from 'zod'
const s = z.number({ required_error: 'r' })
const m = err.errors[0].message`)
	mustWrite(t, filepath.Join(dir, "src", "unrelated.ts"),
		`export const foo = 'bar'`) // should be skipped (no zod imports)

	log, err := phase08ZodRefactor(Options{OutputDir: dir})
	if err != nil {
		t.Fatalf("phase08ZodRefactor returned err: %v", err)
	}
	if log.Status != "ok" {
		t.Errorf("status = %q, want ok", log.Status)
	}
	if !strings.Contains(log.Message, "2 substituições em 1 arquivos") {
		t.Errorf("message = %q; expected 2 replacements / 1 file", log.Message)
	}
	if len(log.Details) != 1 {
		t.Errorf("details len = %d, want 1", len(log.Details))
	}

	out, _ := os.ReadFile(filepath.Join(dir, "src", "lib", "validations.ts"))
	if strings.Contains(string(out), "required_error") {
		t.Error("required_error still present after refactor")
	}
	if !strings.Contains(string(out), "err.issues[0]") {
		t.Error("errors[0] not replaced with issues[0]")
	}
}

func mustWrite(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}
